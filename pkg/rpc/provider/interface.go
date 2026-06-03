package provider

import (
	"context"

	"github.com/sora-soft/sora-go-framework/pkg/rpc/packet"
	"github.com/sora-soft/sora-go-framework/pkg/utility/errorx"
)

type Provider interface {
	Start(ctx context.Context) error
	Stop() error
	CallRpc(ctx context.Context, method string, req any, opts ...CallOption) (packet.Packet, error)
	SendNotify(ctx context.Context, method string, req any, opts ...NotifyOption) error
}

func CallRpc[Resp any](p Provider, ctx context.Context, method string, req any, opts ...CallOption) (Resp, error) {
	var zero Resp
	rawPkt, err := p.CallRpc(ctx, method, req, opts...)
	if err != nil {
		return zero, err
	}

	resp, err := packet.Decode[packet.Response[Resp]](rawPkt)
	if err != nil {
		return zero, err
	}

	if resp.Error != nil {
		return zero, &errorx.Error{
			Code:    resp.Error.Code,
			Level:   errorx.ErrorLevel(resp.Error.Level),
			Name:    "RpcResponseError",
			Message: resp.Error.Message,
			Args:    resp.Error.Args,
		}
	}

	return resp.Result, nil
}
