package etcd

import (
	"context"
	"strings"
	"sync"

	"github.com/sora-soft/sora-go-framework/pkg/component"
	"github.com/sora-soft/sora-go-framework/pkg/logger"
	"github.com/sora-soft/sora-go-framework/pkg/runtime"
	"github.com/sora-soft/sora-go-framework/pkg/utility/errorx"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type LeaseReconnectFunc func(leaseID clientv3.LeaseID, err error)

type EtcdAuth struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

type EtcdClientConfig struct {
	Hosts []string  `json:"hosts" yaml:"hosts"`
	Auth  *EtcdAuth `json:"auth,omitempty" yaml:"auth,omitempty"`
}

type EtcdComponentOptions struct {
	Etcd   EtcdClientConfig `json:"etcd" yaml:"etcd"`
	TTL    int64            `json:"ttl" yaml:"ttl"`
	Prefix string           `json:"prefix" yaml:"prefix"`
}

type EtcdComponent struct {
	client  *clientv3.Client
	options *EtcdComponentOptions
	version string

	lease           clientv3.Lease
	leaseID         clientv3.LeaseID
	keepAliveCtx    context.Context
	keepAliveCancel context.CancelFunc

	persistValues    map[string]string
	onLeaseReconnect []LeaseReconnectFunc
	reconnecting     bool
	destroyed        bool
	reconnectMu      sync.Mutex
}

func NewEtcdComponent(name component.ComponentName) *component.BaseComponent[*EtcdComponent] {
	impl := &EtcdComponent{
		version:       "0.1.0",
		persistValues: make(map[string]string),
	}
	return component.NewBaseComponent(name, impl)
}

func (e *EtcdComponent) Connect(ctx context.Context) error {
	if e.options == nil {
		return newOptionsNotSetError()
	}

	cfg := clientv3.Config{
		Endpoints: e.options.Etcd.Hosts,
	}
	if e.options.Etcd.Auth != nil && e.options.Etcd.Auth.Username != "" {
		cfg.Username = e.options.Etcd.Auth.Username
		cfg.Password = e.options.Etcd.Auth.Password
	}

	client, err := clientv3.New(cfg)
	if err != nil {
		return err
	}

	if _, err := client.Get(ctx, "health"); err != nil {
		client.Close()
		return err
	}

	e.client = client

	if err := e.grantLease(ctx); err != nil {
		return err
	}

	e.startKeepAlive()
	return nil
}

func (e *EtcdComponent) Disconnect() error {
	e.destroyed = true

	if e.keepAliveCancel != nil {
		e.keepAliveCancel()
		e.keepAliveCancel = nil
	}

	if e.lease != nil && e.leaseID != 0 {
		_, _ = e.lease.Revoke(context.Background(), e.leaseID)
		e.lease = nil
		e.leaseID = 0
	}

	if e.client != nil {
		err := e.client.Close()
		e.client = nil
		return err
	}
	return nil
}

func (e *EtcdComponent) SetOptions(opts any) error {
	o, ok := opts.(*EtcdComponentOptions)
	if !ok {
		return newInvalidOptionsError()
	}

	if len(o.Etcd.Hosts) == 0 {
		return newHostsEmptyError()
	}
	if o.TTL <= 0 {
		return newTTLInvalidError(o.TTL)
	}
	if o.Prefix == "" {
		return newPrefixEmptyError()
	}

	e.options = o
	return nil
}

func (e *EtcdComponent) GetOptions() any {
	if e.options == nil {
		return nil
	}
	return *e.options
}

func (e *EtcdComponent) GetVersion() string {
	return e.version
}

func (e *EtcdComponent) Client() *clientv3.Client {
	return e.client
}

func (e *EtcdComponent) LeaseID() clientv3.LeaseID {
	return e.leaseID
}

func (e *EtcdComponent) OnLeaseReconnect(fn LeaseReconnectFunc) {
	e.onLeaseReconnect = append(e.onLeaseReconnect, fn)
}

func (e *EtcdComponent) Keys(args ...string) string {
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, e.options.Prefix)
	parts = append(parts, args...)
	return strings.Join(parts, "/")
}

func (e *EtcdComponent) Lock(ctx context.Context, key string, fn func() error, ttlSec int) error {
	if e.client == nil {
		return newNotConnectedError()
	}

	session, err := concurrency.NewSession(e.client, concurrency.WithTTL(ttlSec))
	if err != nil {
		return errorx.Wrap(err, "ERR_ETCD_LOCK_SESSION", errorx.LevelUnexpected, "EtcdComponentError", "failed to create lock session", nil)
	}
	defer session.Close()

	fullKey := e.Keys(key)
	mu := concurrency.NewMutex(session, fullKey)

	if err := mu.Lock(ctx); err != nil {
		return errorx.Wrap(err, "ERR_ETCD_LOCK_ACQUIRE", errorx.LevelUnexpected, "EtcdComponentError", "failed to acquire lock", map[string]any{"key": fullKey})
	}
	defer mu.Unlock(ctx)

	return fn()
}

func etcdLogger() *logger.Logger {
	return runtime.RT.FrameLogger
}
