package rpc

import "github.com/sora-soft/sora-go-framework.git/pkg/rpc/packet"

type RequestReader struct {
	pkt packet.Packet
}

func newRequestReader(pkt packet.Packet) *RequestReader {
	return &RequestReader{pkt: pkt}
}

func (r *RequestReader) Header(key string) string {
	if r.pkt.Headers == nil {
		return ""
	}
	return r.pkt.Headers[key]
}

func (r *RequestReader) Headers() map[string]string {
	return r.pkt.Headers
}

func (r *RequestReader) Method() string {
	return r.pkt.Method
}

func (r *RequestReader) Service() string {
	return r.pkt.Service
}

func (r *RequestReader) Packet() packet.Packet {
	return r.pkt
}
