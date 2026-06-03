package etcd

import (
	"context"
	"encoding/json"
	"path/filepath"
	"sync"

	"github.com/sora-soft/sora-go-framework/pkg/component"
	etcdcomp "github.com/sora-soft/sora-go-framework/pkg/component/etcd"
	"github.com/sora-soft/sora-go-framework/pkg/discovery"
	"github.com/sora-soft/sora-go-framework/pkg/runtime"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdBackendOptions struct {
	EtcdComponentName string `json:"etcdComponentName" yaml:"etcdComponentName"`
	Scope             string `json:"scope" yaml:"scope"`
}

type EtcdBackend struct {
	options  EtcdBackendOptions
	comp     component.Component
	client   *clientv3.Client
	store    *store
	registry *etcdRegistry
	discover *etcdDiscovery

	watchCtx    context.Context
	watchCancel context.CancelFunc
	watchers    []clientv3.Watcher

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
	comp, err := runtime.GetComponentOf[*etcdcomp.EtcdComponent](component.ComponentName(b.options.EtcdComponentName))
	if err != nil {
		return err
	}

	if err := comp.Start(ctx); err != nil {
		return err
	}
	b.comp = comp

	impl := comp.Impl()
	b.client = impl.Client()
	if b.client == nil {
		return newNotConnectedError()
	}

	impl.OnLeaseReconnect(func(leaseID clientv3.LeaseID, err error) {
		b.handleLeaseReconnect(leaseID, err)
	})

	b.watchCtx, b.watchCancel = context.WithCancel(context.Background())

	if err := b.initFromEtcd(ctx); err != nil {
		b.Disconnect()
		return err
	}

	if err := b.startWatchers(); err != nil {
		b.Disconnect()
		return err
	}

	b.registry = newEtcdRegistry(b, impl)
	b.discover = &etcdDiscovery{store: b.store}
	return nil
}

func (b *EtcdBackend) handleLeaseReconnect(leaseID clientv3.LeaseID, _ error) {
	ctx := context.Background()
	b.registry.reRegisterAll(ctx, leaseID)
}

func (b *EtcdBackend) initFromEtcd(ctx context.Context) error {
	entityTypes := []entityType{entityService, entityEndpoint, entityNode}

	for _, et := range entityTypes {
		prefix := entityPrefix(b.options.Scope, et)
		resp, err := b.client.Get(ctx, prefix, clientv3.WithPrefix())
		if err != nil {
			return err
		}
		for _, kv := range resp.Kvs {
			key := string(kv.Key)
			id := key[len(prefix)+1:]
			switch et {
			case entityNode:
				b.store.updateNode(&mvccpb.KeyValue{
					Key:            []byte(id),
					CreateRevision: kv.CreateRevision,
					ModRevision:    kv.ModRevision,
					Value:          kv.Value,
				})
			case entityService:
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
		{entityNode, b.store.updateNode, b.handleDeleteNode, entityPrefix(b.options.Scope, entityNode)},
		{entityService, b.store.updateService, b.handleDeleteService, entityPrefix(b.options.Scope, entityService)},
		{entityWorker, b.store.updateWorker, b.handleDeleteWorker, entityPrefix(b.options.Scope, entityWorker)},
		{entityEndpoint, b.store.updateEndpoint, b.handleDeleteEndpoint, entityPrefix(b.options.Scope, entityEndpoint)},
	}

	for _, target := range watchTargets {
		watcher := clientv3.NewWatcher(b.client)
		b.watchers = append(b.watchers, watcher)

		wc := watcher.Watch(b.watchCtx, target.prefix, clientv3.WithPrefix())
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
	if b.watchCancel != nil {
		b.watchCancel()
		b.watchCancel = nil
	}

	for _, w := range b.watchers {
		w.Close()
	}
	b.watchers = nil

	b.client = nil

	if b.comp != nil {
		b.comp.Stop()
		b.comp = nil
	}

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

	electionPath := filepath.Join(b.options.Scope, "singleton", name)
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

func (b *EtcdBackend) putWithLease(ctx context.Context, key string, value any, leaseID clientv3.LeaseID) error {
	if b.client == nil {
		return newNotConnectedError()
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, err = b.client.Put(ctx, key, string(data), clientv3.WithLease(leaseID))
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
