package component

import (
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdOptions struct {
	Endpoints   []string      `json:"endpoints"`
	DialTimeout time.Duration `json:"dialTimeout"`
	Username    string        `json:"username,omitempty"`
	Password    string        `json:"password,omitempty"`
}

type EtcdComponent struct {
	client  *clientv3.Client
	options *EtcdOptions
	version string
}

func NewEtcdComponent(name string) Component {
	impl := &EtcdComponent{
		version: "0.1.0",
	}
	return NewBaseComponent(name, impl)
}

func (e *EtcdComponent) Connect(ctx context.Context) error {
	if e.options == nil {
		return ErrEtcdOptionsNotSet
	}

	cfg := clientv3.Config{
		Endpoints:   e.options.Endpoints,
		DialTimeout: e.options.DialTimeout,
	}
	if e.options.Username != "" {
		cfg.Username = e.options.Username
		cfg.Password = e.options.Password
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
	return nil
}

func (e *EtcdComponent) Disconnect() error {
	if e.client != nil {
		err := e.client.Close()
		e.client = nil
		return err
	}
	return nil
}

func (e *EtcdComponent) SetOptions(opts any) error {
	o, ok := opts.(*EtcdOptions)
	if !ok {
		return ErrEtcdInvalidOptions
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

type BaseEtcdComponent struct {
	*baseComponent
}

func NewBaseEtcdComponent(name string) *BaseEtcdComponent {
	impl := &EtcdComponent{
		version: "0.1.0",
	}
	return &BaseEtcdComponent{
		baseComponent: NewBaseComponent(name, impl),
	}
}

func (b *BaseEtcdComponent) Client() *clientv3.Client {
	return b.baseComponent.componentImpl.(*EtcdComponent).Client()
}
