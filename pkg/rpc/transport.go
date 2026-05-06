package rpc

import "context"

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

type TransportListener interface {
	Accept(ctx context.Context) (*Connection, error)
	Close() error
	GetMetaInfo() TransportMetaInfo
}
