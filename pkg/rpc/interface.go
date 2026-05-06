package rpc

import (
	"github.com/sora-soft/sora-go-framework.git/pkg/rpc/packet"
)

type Codec interface {
	GetCode() string
	Encode(pkt packet.Packet) ([]byte, error)
	Decode(data []byte) (packet.Packet, error)
}
