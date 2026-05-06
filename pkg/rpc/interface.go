package rpc

import (
	"github.com/sora-soft/sora-go-framework.git/pkg/rpc/packet"
)

type Codec interface {
	GetCode() string
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
	EncodePacket(pkt packet.Packet) ([]byte, error)
	DecodePacket(data []byte) (packet.Packet, error)
}
