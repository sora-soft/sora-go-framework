package rpc

import (
	"context"
	"sync"
)

type Transport interface {
	Connect(ctx context.Context, endpoint string, codec string) (string, error)
	Handshake(ctx context.Context) (string, error)
	Send(ctx context.Context, data []byte) error
	Recv(ctx context.Context) ([]byte, error)
	Close() error
}

type TransportMetaInfo struct {
	Protocol string `json:"protocol"`
	Endpoint string `json:"endpoint"`
}

type TransportFactory func() Transport

type TransportConfig struct {
	Factory  TransportFactory
	Options  ConnectorOptions
}

var (
	transportMu    sync.RWMutex
	transportConfs = make(map[string]TransportConfig)
)

func RegisterTransport(protocol string, factory TransportFactory, opts ConnectorOptions) {
	transportMu.Lock()
	defer transportMu.Unlock()
	transportConfs[protocol] = TransportConfig{Factory: factory, Options: opts}
}

func GetTransportConfig(protocol string) (TransportConfig, bool) {
	transportMu.RLock()
	defer transportMu.RUnlock()
	c, ok := transportConfs[protocol]
	return c, ok
}

const HeaderRpcId = "x-sora-rpc-id"

type TransportListener interface {
	Accept(ctx context.Context) (*Connection, error)
	Close() error
	GetMetaInfo() TransportMetaInfo
}
