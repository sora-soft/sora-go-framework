package etcd

import (
	"context"
	"encoding/json"
	"path/filepath"
	"sync"

	"github.com/sora-soft/sora-go-framework.git/pkg/component"
	"github.com/sora-soft/sora-go-framework.git/pkg/discovery"
	"github.com/sora-soft/sora-go-framework.git/pkg/runtime"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdBackendOptions struct {
	ComponentName string
	Prefix        string
	TTL           int64
}

type EtcdBackend struct {
	options  EtcdBackendOptions
	client   *clientv3.Client
	lease    clientv3.Lease
	leaseID  clientv3.LeaseID
	store    *store
	registry *etcdRegistry
	discover *etcdDiscovery

	keepAliveCtx    context.Context
	keepAliveCancel context.CancelFunc

	watchers []clientv3.Watcher

	electionsMu sync.Mutex
	elections   map[string]*etcdElection
}

func NewEtcdBackend(options EtcdBackendOptions) *EtcdBackend {
	s := newStore()
	return &EtcdBackend{
		options:   options,
		store:     s,
		elections: make(map[string]*etcdElection),
	}
}

func (b *EtcdBackend) Connect(ctx context.Context) error {
	c, ok := runtime.RT.GetComponent(b.options.ComponentName)
	if !ok {
		return newComponentNotFoundError(b.options.ComponentName)
	}

	etcdComp, ok := c.(*component.BaseEtcdComponent)
	if !ok {
		return newComponentTypeError(b.options.ComponentName)
	}

	b.client = etcdComp.Client()
	if b.client == nil {
		return newNotConnectedError()
	}

	lease := clientv3.NewLease(b.client)
	resp, err := lease.Grant(ctx, b.options.TTL)
	if err != nil {
		return err
	}
	b.lease = lease
	b.leaseID = resp.ID

	b.keepAliveCtx, b.keepAliveCancel = context.WithCancel(context.Background())
	ch, err := lease.KeepAlive(b.keepAliveCtx, resp.ID)
	if err != nil {
		lease.Revoke(ctx, resp.ID)
		return err
	}
	go b.drainKeepAlive(ch)

	if err := b.initFromEtcd(ctx); err != nil {
		b.Disconnect()
		return err
	}

	if err := b.startWatchers(); err != nil {
		b.Disconnect()
		return err
	}

	b.registry = &etcdRegistry{backend: b}
	b.discover = &etcdDiscovery{store: b.store}
	return nil
}

func (b *EtcdBackend) drainKeepAlive(ch <-chan *clientv3.LeaseKeepAliveResponse) {
	for range ch {
	}
}

func (b *EtcdBackend) initFromEtcd(ctx context.Context) error {
	entityTypes := []entityType{entityService, entityEndpoint, entityNode}

	for _, et := range entityTypes {
		prefix := entityPrefix(b.options.Prefix, et)
		resp, err := b.client.Get(ctx, prefix, clientv3.WithPrefix())
		if err != nil {
			return err
		}
		for _, kv := range resp.Kvs {
			key := string(kv.Key)
			id := key[len(prefix)+1:]
			println("etcd init key: " + key)
			switch et {
			case entityNode:
				println("entityNode!!!")
				b.store.updateNode(&mvccpb.KeyValue{
					Key:            []byte(id),
					CreateRevision: kv.CreateRevision,
					ModRevision:    kv.ModRevision,
					Value:          kv.Value,
				})
			case entityService:
				println("entityService!!!" + id)
				b.store.updateService(&mvccpb.KeyValue{
					Key:            []byte(id),
					CreateRevision: kv.CreateRevision,
					ModRevision:    kv.ModRevision,
					Value:          kv.Value,
				})
			case entityEndpoint:
				b.store.updateEndpoint(&mvccpb.KeyValue{
					Key:            []byte(id),
					CreateRevision: kv.CreateRevision,
					ModRevision:    kv.ModRevision,
					Value:          kv.Value,
				})
			}
		}
	}
	return nil
}

func (b *EtcdBackend) startWatchers() error {
	watchTargets := []struct {
		et     entityType
		onPut  func(kv *mvccpb.KeyValue)
		onDel  func(kv *mvccpb.KeyValue)
		prefix string
	}{
		{entityNode, b.store.updateNode, b.handleDeleteNode, entityPrefix(b.options.Prefix, entityNode)},
		{entityService, b.store.updateService, b.handleDeleteService, entityPrefix(b.options.Prefix, entityService)},
		{entityWorker, b.store.updateWorker, b.handleDeleteWorker, entityPrefix(b.options.Prefix, entityWorker)},
		{entityEndpoint, b.store.updateEndpoint, b.handleDeleteEndpoint, entityPrefix(b.options.Prefix, entityEndpoint)},
	}

	for _, target := range watchTargets {
		watcher := clientv3.NewWatcher(b.client)
		b.watchers = append(b.watchers, watcher)

		wc := watcher.Watch(context.Background(), target.prefix, clientv3.WithPrefix())
		go b.watchLoop(wc, target.et, target.prefix, target.onPut, target.onDel)
	}

	return nil
}

func (b *EtcdBackend) handleDeleteNode(kv *mvccpb.KeyValue) {
	b.store.deleteNode(string(kv.Key))
}

func (b *EtcdBackend) handleDeleteService(kv *mvccpb.KeyValue) {
	b.store.deleteService(string(kv.Key))
}

func (b *EtcdBackend) handleDeleteWorker(kv *mvccpb.KeyValue) {
	b.store.deleteWorker(string(kv.Key))
}

func (b *EtcdBackend) handleDeleteEndpoint(kv *mvccpb.KeyValue) {
	b.store.deleteEndpoint(string(kv.Key))
}

func (b *EtcdBackend) watchLoop(wc clientv3.WatchChan, et entityType, prefix string, onPut func(kv *mvccpb.KeyValue), onDel func(kv *mvccpb.KeyValue)) {
	for resp := range wc {
		for _, ev := range resp.Events {
			kv := ev.Kv
			id := string(kv.Key)[len(prefix)+1:]
			normalized := &mvccpb.KeyValue{
				Key:            []byte(id),
				CreateRevision: kv.CreateRevision,
				ModRevision:    kv.ModRevision,
				Value:          kv.Value,
			}
			switch ev.Type {
			case mvccpb.PUT:
				onPut(normalized)
			case mvccpb.DELETE:
				onDel(normalized)
			}
		}
	}
}

func (b *EtcdBackend) Disconnect() error {
	if b.keepAliveCancel != nil {
		b.keepAliveCancel()
	}

	for _, w := range b.watchers {
		w.Close()
	}
	b.watchers = nil

	if b.lease != nil {
		b.lease.Close()
		b.lease = nil
	}

	b.client = nil

	return nil
}

func (b *EtcdBackend) Registry() discovery.Registry {
	return b.registry
}

func (b *EtcdBackend) Discovery() discovery.Discovery {
	return b.discover
}

func (b *EtcdBackend) NewElection(name string) discovery.Election {
	b.electionsMu.Lock()
	defer b.electionsMu.Unlock()

	if e, ok := b.elections[name]; ok {
		return e
	}

	electionPath := filepath.Join(b.options.Prefix, "singleton", name)
	e := &etcdElection{
		client:      b.client,
		electionKey: electionPath,
	}
	b.elections[name] = e
	return e
}

func (b *EtcdBackend) GetInfo() discovery.BackendInfo {
	return discovery.BackendInfo{
		Type:    "etcd",
		Version: "0.1.0",
	}
}

func (b *EtcdBackend) putWithLease(ctx context.Context, key string, value any) error {
	if b.client == nil {
		return newNotConnectedError()
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, err = b.client.Put(ctx, key, string(data), clientv3.WithLease(b.leaseID))
	return err
}

func (b *EtcdBackend) deleteKey(ctx context.Context, key string) error {
	if b.client == nil {
		return newNotConnectedError()
	}

	_, err := b.client.Delete(ctx, key)
	return err
}

var _ discovery.Backend = (*EtcdBackend)(nil)
