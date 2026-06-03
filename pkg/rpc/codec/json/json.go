package json

import (
	"encoding/json"

	"github.com/sora-soft/sora-go-framework.git/pkg/rpc"
	"github.com/sora-soft/sora-go-framework.git/pkg/rpc/packet"
	"github.com/sora-soft/sora-go-framework.git/pkg/utility/errorx"
)

type JSONBufferCodec struct{}

func (c *JSONBufferCodec) GetCode() string {
	return "json"
}

func (c *JSONBufferCodec) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (c *JSONBufferCodec) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

type requestWire struct {
	Opcode  packet.PacketOpcode `json:"opcode" yaml:"opcode"`
	Method  string              `json:"method" yaml:"method"`
	Service string              `json:"service" yaml:"service"`
	Headers map[string]string   `json:"headers" yaml:"headers"`
	Payload json.RawMessage     `json:"payload" yaml:"payload"`
}

type responseWire struct {
	Opcode  packet.PacketOpcode `json:"opcode" yaml:"opcode"`
	Headers map[string]string   `json:"headers" yaml:"headers"`
	Payload json.RawMessage     `json:"payload" yaml:"payload"`
}

type notifyWire struct {
	Opcode  packet.PacketOpcode `json:"opcode" yaml:"opcode"`
	Method  string              `json:"method" yaml:"method"`
	Service string              `json:"service" yaml:"service"`
	Headers map[string]string   `json:"headers" yaml:"headers"`
	Payload json.RawMessage     `json:"payload" yaml:"payload"`
}

type commandWire struct {
	Opcode  packet.PacketOpcode     `json:"opcode" yaml:"opcode"`
	Command packet.ConnectorCommand `json:"command" yaml:"command"`
	Args    json.RawMessage         `json:"args" yaml:"args"`
}

func ensureHeaders(h map[string]string) map[string]string {
	if h == nil {
		return map[string]string{}
	}
	return h
}

func (c *JSONBufferCodec) EncodePacket(pkt packet.Packet) ([]byte, error) {
	switch pkt.Opcode {
	case packet.PacketOpcodeRequest:
		return json.Marshal(requestWire{
			Opcode:  pkt.Opcode,
			Method:  pkt.Method,
			Service: pkt.Service,
			Headers: ensureHeaders(pkt.Headers),
			Payload: pkt.Payload(),
		})
	case packet.PacketOpcodeResponse:
		return json.Marshal(responseWire{
			Opcode:  pkt.Opcode,
			Headers: ensureHeaders(pkt.Headers),
			Payload: pkt.Payload(),
		})
	case packet.PacketOpcodeNotify:
		return json.Marshal(notifyWire{
			Opcode:  pkt.Opcode,
			Method:  pkt.Method,
			Service: pkt.Service,
			Headers: ensureHeaders(pkt.Headers),
			Payload: pkt.Payload(),
		})
	case packet.PacketOpcodeCommand:
		return json.Marshal(commandWire{
			Opcode:  pkt.Opcode,
			Command: packet.ConnectorCommand(pkt.Method),
			Args:    pkt.Payload(),
		})
	default:
		return nil, errorx.New("ERR_CODEC_UNSUPPORTED_OPCODE", errorx.LevelUnexpected, "JSONCodecError", "unsupported opcode", map[string]any{"opcode": pkt.Opcode})
	}
}

func (c *JSONBufferCodec) DecodePacket(data []byte) (packet.Packet, error) {
	var header struct {
		Opcode packet.PacketOpcode `json:"opcode" yaml:"opcode"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return packet.Packet{}, errorx.Wrap(err, "ERR_CODEC_DECODE_FAILED", errorx.LevelUnexpected, "JSONCodecError", "failed to decode opcode", nil)
	}

	switch header.Opcode {
	case packet.PacketOpcodeRequest:
		var w requestWire
		if err := json.Unmarshal(data, &w); err != nil {
			return packet.Packet{}, errorx.Wrap(err, "ERR_CODEC_DECODE_FAILED", errorx.LevelUnexpected, "JSONCodecError", "failed to decode request packet", nil)
		}
		return packet.NewDecodedPacket(w.Opcode, w.Method, w.Service, w.Headers, w.Payload, c), nil
	case packet.PacketOpcodeResponse:
		var w responseWire
		if err := json.Unmarshal(data, &w); err != nil {
			return packet.Packet{}, errorx.Wrap(err, "ERR_CODEC_DECODE_FAILED", errorx.LevelUnexpected, "JSONCodecError", "failed to decode response packet", nil)
		}
		return packet.NewDecodedPacket(w.Opcode, "", "", w.Headers, w.Payload, c), nil
	case packet.PacketOpcodeNotify:
		var w notifyWire
		if err := json.Unmarshal(data, &w); err != nil {
			return packet.Packet{}, errorx.Wrap(err, "ERR_CODEC_DECODE_FAILED", errorx.LevelUnexpected, "JSONCodecError", "failed to decode notify packet", nil)
		}
		return packet.NewDecodedPacket(w.Opcode, w.Method, w.Service, w.Headers, w.Payload, c), nil
	case packet.PacketOpcodeCommand:
		var w commandWire
		if err := json.Unmarshal(data, &w); err != nil {
			return packet.Packet{}, errorx.Wrap(err, "ERR_CODEC_DECODE_FAILED", errorx.LevelUnexpected, "JSONCodecError", "failed to decode command packet", nil)
		}
		return packet.NewDecodedPacket(w.Opcode, string(w.Command), "", nil, w.Args, c), nil
	default:
		return packet.Packet{}, errorx.New("ERR_CODEC_UNSUPPORTED_OPCODE", errorx.LevelUnexpected, "JSONCodecError", "unsupported opcode", map[string]any{"opcode": header.Opcode})
	}
}

var _ rpc.Codec = (*JSONBufferCodec)(nil)
