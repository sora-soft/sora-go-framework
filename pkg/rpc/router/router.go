package router

import (
	"context"
	"fmt"

	"github.com/sora-soft/sora-go-framework.git/pkg/logger"
	"github.com/sora-soft/sora-go-framework.git/pkg/rpc"
	"github.com/sora-soft/sora-go-framework.git/pkg/rpc/packet"
	"github.com/sora-soft/sora-go-framework.git/pkg/runtime"
	"github.com/sora-soft/sora-go-framework.git/pkg/utility/errorx"
)

type DispatchFunc func(ctx rpc.HandlerContext) error

type Middleware func(next DispatchFunc) DispatchFunc

type Router struct {
	methodTable map[string]DispatchFunc
	notifyTable map[string]DispatchFunc
	middlewares []Middleware
	logger      *logger.Logger
}

type RouterOption func(*Router)

func WithLogger(l *logger.Logger) RouterOption {
	return func(r *Router) {
		r.logger = l
	}
}

func NewRouter(opts ...RouterOption) *Router {
	r := &Router{
		methodTable: make(map[string]DispatchFunc),
		notifyTable: make(map[string]DispatchFunc),
		logger:      runtime.RT.RpcLogger,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func Method[Req any, Resp any](r *Router, method string, handler func(ctx *rpc.RequestContext, req Req) (Resp, error)) {
	r.methodTable[method] = func(ctx rpc.HandlerContext) error {
		rc := ctx.(*rpc.RequestContext)
		req, err := packet.Decode[Req](rc.Reader().Packet())
		if err != nil {
			r.sendDecodeErrorResponse(rc, err)
			return nil
		}
		resp, err := handler(rc, req)
		if err != nil {
			r.sendErrorResponse(rc, err)
			return nil
		}
		r.sendSuccessResponse(rc, resp)
		return nil
	}
}

func Notify[Msg any](r *Router, method string, handler func(ctx *rpc.NotifyContext, msg Msg) error) {
	r.notifyTable[method] = func(ctx rpc.HandlerContext) error {
		nc, ok := ctx.(*rpc.NotifyContext)
		if !ok {
			return nil
		}
		msg, err := packet.Decode[Msg](nc.Reader().Packet())
		if err != nil {
			r.logger.Error("notify.decode-failed", err, map[string]any{"method": method})
			return nil
		}
		if err := handler(nc, msg); err != nil {
			r.logger.Error("notify.handler-error", err, map[string]any{"method": method})
		}
		return nil
	}
}

func (r *Router) Use(mw Middleware) {
	r.middlewares = append(r.middlewares, mw)
}

func (r *Router) OnRequestCB() func(conn *rpc.Connection, pkt packet.Packet) {
	chain := r.buildChain(r.dispatchRequest)
	return func(conn *rpc.Connection, pkt packet.Packet) {
		ctx := rpc.NewRequestContext(conn, pkt)
		defer func() {
			if rec := recover(); rec != nil {
				r.sendErrorResponse(ctx, fmt.Errorf("panic: %v", rec))
			}
		}()
		if err := chain(ctx); err != nil {
			r.sendErrorResponse(ctx, err)
		}
	}
}

func (r *Router) OnNotifyCB() func(conn *rpc.Connection, pkt packet.Packet) {
	chain := r.buildChain(r.dispatchNotify)
	return func(conn *rpc.Connection, pkt packet.Packet) {
		ctx := rpc.NewNotifyContext(conn, pkt)
		defer func() {
			if rec := recover(); rec != nil {
				r.logger.Error("notify.panic", fmt.Errorf("%v", rec), map[string]any{"method": ctx.Reader().Method()})
			}
		}()
		if err := chain(ctx); err != nil {
			r.logger.Error("notify.middleware-error", err, map[string]any{"method": ctx.Reader().Method()})
		}
	}
}

func (r *Router) buildChain(core DispatchFunc) DispatchFunc {
	chain := core
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		chain = r.middlewares[i](chain)
	}
	return chain
}

func (r *Router) dispatchRequest(ctx rpc.HandlerContext) error {
	entry, ok := r.methodTable[ctx.Reader().Method()]
	if !ok {
		rc := ctx.(*rpc.RequestContext)
		r.sendMethodNotFoundResponse(rc, ctx.Reader().Method())
		return nil
	}
	return entry(ctx)
}

func (r *Router) dispatchNotify(ctx rpc.HandlerContext) error {
	entry, ok := r.notifyTable[ctx.Reader().Method()]
	if !ok {
		r.logger.Warn("notify.method-not-found", map[string]any{"method": ctx.Reader().Method()})
		return nil
	}
	return entry(ctx)
}

func mergeHeaders(base map[string]string, override map[string]string) map[string]string {
	merged := make(map[string]string, len(base)+len(override))
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range override {
		merged[k] = v
	}
	return merged
}

func (r *Router) sendSuccessResponse(ctx *rpc.RequestContext, resp any) {
	pkt := ctx.Reader().Packet()
	codec := pkt.Codec()
	payload := packet.Response[any]{
		Error:  nil,
		Result: resp,
	}
	data, err := codec.Marshal(payload)
	if err != nil {
		r.logger.Error("response.marshal-failed", err, nil)
		return
	}
	headers := mergeHeaders(pkt.Headers, ctx.Res().Headers())
	respPkt := packet.NewDecodedPacket(packet.PacketOpcodeResponse, "", "", headers, data, codec)
	if err := ctx.Conn().SendResponse(context.Background(), respPkt); err != nil {
		r.logger.Error("response.send-failed", err, nil)
	}
}

func (r *Router) sendErrorResponse(ctx *rpc.RequestContext, err error) {
	pkt := ctx.Reader().Packet()
	var pe *packet.PayloadError
	switch e := err.(type) {
	case *errorx.Error:
		pe = &packet.PayloadError{
			Code:    e.Code,
			Message: e.Message,
			Level:   int(e.Level),
			Name:    e.Name,
			Args:    e.Extra,
		}
	default:
		pe = &packet.PayloadError{
			Code:    "ERR_INTERNAL",
			Message: err.Error(),
			Level:   int(errorx.LevelUnexpected),
			Name:    "InternalError",
		}
	}
	codec := pkt.Codec()
	payload := packet.Response[any]{
		Error:  pe,
		Result: nil,
	}
	data, marshalErr := codec.Marshal(payload)
	if marshalErr != nil {
		r.logger.Error("response.marshal-failed", marshalErr, nil)
		return
	}
	headers := mergeHeaders(pkt.Headers, ctx.Res().Headers())
	respPkt := packet.NewDecodedPacket(packet.PacketOpcodeResponse, "", "", headers, data, codec)
	if sendErr := ctx.Conn().SendResponse(context.Background(), respPkt); sendErr != nil {
		r.logger.Error("response.send-failed", sendErr, nil)
	}
}

func (r *Router) sendMethodNotFoundResponse(ctx *rpc.RequestContext, method string) {
	pkt := ctx.Reader().Packet()
	pe := &packet.PayloadError{
		Code:    "ERR_METHOD_NOT_FOUND",
		Message: fmt.Sprintf("method '%s' not found", method),
		Level:   int(errorx.LevelExpected),
		Name:    "MethodNotFoundError",
	}
	codec := pkt.Codec()
	payload := packet.Response[any]{
		Error:  pe,
		Result: nil,
	}
	data, err := codec.Marshal(payload)
	if err != nil {
		r.logger.Error("response.marshal-failed", err, nil)
		return
	}
	headers := mergeHeaders(pkt.Headers, ctx.Res().Headers())
	respPkt := packet.NewDecodedPacket(packet.PacketOpcodeResponse, "", "", headers, data, codec)
	if err := ctx.Conn().SendResponse(context.Background(), respPkt); err != nil {
		r.logger.Error("response.send-failed", err, nil)
	}
}

func (r *Router) sendDecodeErrorResponse(ctx *rpc.RequestContext, decodeErr error) {
	pkt := ctx.Reader().Packet()
	method := pkt.Method
	pe := &packet.PayloadError{
		Code:    "ERR_DECODE_FAILED",
		Message: fmt.Sprintf("failed to decode request for method '%s': %s", method, decodeErr.Error()),
		Level:   int(errorx.LevelExpected),
		Name:    "DecodeError",
	}
	codec := pkt.Codec()
	payload := packet.Response[any]{
		Error:  pe,
		Result: nil,
	}
	data, err := codec.Marshal(payload)
	if err != nil {
		r.logger.Error("response.marshal-failed", err, nil)
		return
	}
	headers := mergeHeaders(pkt.Headers, ctx.Res().Headers())
	respPkt := packet.NewDecodedPacket(packet.PacketOpcodeResponse, "", "", headers, data, codec)
	if err := ctx.Conn().SendResponse(context.Background(), respPkt); err != nil {
		r.logger.Error("response.send-failed", err, nil)
	}
}
