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

type DispatchFunc func(conn *rpc.Connection, pkt packet.Packet) error

type Middleware func(next DispatchFunc) DispatchFunc

type Router struct {
	methodTable map[string]DispatchFunc
	notifyTable map[string]func(conn *rpc.Connection, pkt packet.Packet)
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
		notifyTable: make(map[string]func(conn *rpc.Connection, pkt packet.Packet)),
		logger:      runtime.RT.RpcLogger,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func Method[Req any, Resp any](r *Router, method string, handler func(conn *rpc.Connection, req Req) (Resp, error)) {
	r.methodTable[method] = func(conn *rpc.Connection, pkt packet.Packet) error {
		req, err := packet.Decode[Req](pkt)
		if err != nil {
			r.sendDecodeErrorResponse(conn, pkt, method, err)
			return nil
		}
		resp, err := handler(conn, req)
		if err != nil {
			r.sendErrorResponse(conn, pkt, err)
			return nil
		}
		r.sendSuccessResponse(conn, pkt, resp)
		return nil
	}
}

func Notify[Msg any](r *Router, method string, handler func(conn *rpc.Connection, msg Msg) error) {
	r.notifyTable[method] = func(conn *rpc.Connection, pkt packet.Packet) {
		msg, err := packet.Decode[Msg](pkt)
		if err != nil {
			r.logger.Error("notify.decode-failed", err, map[string]any{"method": method})
			return
		}
		if err := handler(conn, msg); err != nil {
			r.logger.Error("notify.handler-error", err, map[string]any{"method": method})
		}
	}
}

func (r *Router) Use(mw Middleware) {
	r.middlewares = append(r.middlewares, mw)
}

func (r *Router) OnRequestCB() func(conn *rpc.Connection, pkt packet.Packet) {
	chain := r.buildChain(r.dispatchRequest)
	return func(conn *rpc.Connection, pkt packet.Packet) {
		defer func() {
			if rec := recover(); rec != nil {
				r.sendErrorResponse(conn, pkt, fmt.Errorf("panic: %v", rec))
			}
		}()
		if err := chain(conn, pkt); err != nil {
			r.sendErrorResponse(conn, pkt, err)
		}
	}
}

func (r *Router) OnNotifyCB() func(conn *rpc.Connection, pkt packet.Packet) {
	chain := r.buildChain(r.dispatchNotify)
	return func(conn *rpc.Connection, pkt packet.Packet) {
		defer func() {
			if rec := recover(); rec != nil {
				r.logger.Error("notify.panic", fmt.Errorf("%v", rec), map[string]any{"method": pkt.Method})
			}
		}()
		if err := chain(conn, pkt); err != nil {
			r.logger.Error("notify.middleware-error", err, map[string]any{"method": pkt.Method})
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

func (r *Router) dispatchRequest(conn *rpc.Connection, pkt packet.Packet) error {
	entry, ok := r.methodTable[pkt.Method]
	if !ok {
		r.sendMethodNotFoundResponse(conn, pkt, pkt.Method)
		return nil
	}
	return entry(conn, pkt)
}

func (r *Router) dispatchNotify(conn *rpc.Connection, pkt packet.Packet) error {
	entry, ok := r.notifyTable[pkt.Method]
	if !ok {
		r.logger.Warn("notify.method-not-found", map[string]any{"method": pkt.Method})
		return nil
	}
	entry(conn, pkt)
	return nil
}

func (r *Router) sendSuccessResponse(conn *rpc.Connection, pkt packet.Packet, resp any) {
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
	respPkt := packet.NewDecodedPacket(packet.PacketOpcodeResponse, "", "", pkt.Headers, data, codec)
	if err := conn.SendResponse(context.Background(), respPkt); err != nil {
		r.logger.Error("response.send-failed", err, nil)
	}
}

func (r *Router) sendErrorResponse(conn *rpc.Connection, pkt packet.Packet, err error) {
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
	respPkt := packet.NewDecodedPacket(packet.PacketOpcodeResponse, "", "", pkt.Headers, data, codec)
	if sendErr := conn.SendResponse(context.Background(), respPkt); sendErr != nil {
		r.logger.Error("response.send-failed", sendErr, nil)
	}
}

func (r *Router) sendMethodNotFoundResponse(conn *rpc.Connection, pkt packet.Packet, method string) {
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
	respPkt := packet.NewDecodedPacket(packet.PacketOpcodeResponse, "", "", pkt.Headers, data, codec)
	if err := conn.SendResponse(context.Background(), respPkt); err != nil {
		r.logger.Error("response.send-failed", err, nil)
	}
}

func (r *Router) sendDecodeErrorResponse(conn *rpc.Connection, pkt packet.Packet, method string, decodeErr error) {
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
	respPkt := packet.NewDecodedPacket(packet.PacketOpcodeResponse, "", "", pkt.Headers, data, codec)
	if err := conn.SendResponse(context.Background(), respPkt); err != nil {
		r.logger.Error("response.send-failed", err, nil)
	}
}
