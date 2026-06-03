package rpc

import (
	"context"

	"github.com/sora-soft/sora-go-framework/pkg/rpc/packet"
)

type HandlerContext interface {
	Conn() *Connection
	Context() context.Context
	Reader() *RequestReader
	Set(k string, v any)
	Get(k string) any
}

type baseContext struct {
	conn   *Connection
	ctx    context.Context
	reader *RequestReader
	store  map[string]any
}

func (bc *baseContext) Conn() *Connection {
	return bc.conn
}

func (bc *baseContext) Context() context.Context {
	return bc.ctx
}

func (bc *baseContext) Reader() *RequestReader {
	return bc.reader
}

func (bc *baseContext) Set(k string, v any) {
	bc.store[k] = v
}

func (bc *baseContext) Get(k string) any {
	return bc.store[k]
}

type RequestContext struct {
	baseContext
	res *ResponseWriter
}

func NewRequestContext(conn *Connection, pkt packet.Packet) *RequestContext {
	return &RequestContext{
		baseContext: baseContext{
			conn:   conn,
			ctx:    context.Background(),
			reader: newRequestReader(pkt),
			store:  make(map[string]any),
		},
		res: newResponseWriter(),
	}
}

func (rc *RequestContext) Res() *ResponseWriter {
	return rc.res
}

func (rc *RequestContext) WithContext(ctx context.Context) *RequestContext {
	rc.ctx = ctx
	return rc
}

type NotifyContext struct {
	baseContext
}

func NewNotifyContext(conn *Connection, pkt packet.Packet) *NotifyContext {
	return &NotifyContext{
		baseContext: baseContext{
			conn:   conn,
			ctx:    context.Background(),
			reader: newRequestReader(pkt),
			store:  make(map[string]any),
		},
	}
}

func (nc *NotifyContext) WithContext(ctx context.Context) *NotifyContext {
	nc.ctx = ctx
	return nc
}
