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

func (c *JSONBufferCodec) Encode(pkt packet.Packet) ([]byte, error) {
	switch pkt.Opcode {
	case packet.PacketOpcodeRequest:
		if pkt.Req == nil {
			return nil, errorx.New("ERR_CODEC_NIL_FIELD", errorx.LevelUnexpected, "JSONCodecError", "request packet has nil Req field", nil)
		}
		return json.Marshal(pkt.Req)
	case packet.PacketOpcodeResponse:
		if pkt.Res == nil {
			return nil, errorx.New("ERR_CODEC_NIL_FIELD", errorx.LevelUnexpected, "JSONCodecError", "response packet has nil Res field", nil)
		}
		return json.Marshal(pkt.Res)
	case packet.PacketOpcodeNotify:
		if pkt.Notify == nil {
			return nil, errorx.New("ERR_CODEC_NIL_FIELD", errorx.LevelUnexpected, "JSONCodecError", "notify packet has nil Notify field", nil)
		}
		return json.Marshal(pkt.Notify)
	case packet.PacketOpcodeCommand:
		if pkt.Cmd == nil {
			return nil, errorx.New("ERR_CODEC_NIL_FIELD", errorx.LevelUnexpected, "JSONCodecError", "command packet has nil Cmd field", nil)
		}
		return json.Marshal(pkt.Cmd)
	default:
		return nil, errorx.New("ERR_CODEC_UNSUPPORTED_OPCODE", errorx.LevelUnexpected, "JSONCodecError", "unsupported opcode", map[string]any{"opcode": pkt.Opcode})
	}
}

func (c *JSONBufferCodec) Decode(data []byte) (packet.Packet, error) {
	var header struct {
		Opcode packet.PacketOpcode `json:"opcode"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return packet.Packet{}, errorx.Wrap(err, "ERR_CODEC_DECODE_FAILED", errorx.LevelUnexpected, "JSONCodecError", "failed to decode opcode", nil)
	}

	switch header.Opcode {
	case packet.PacketOpcodeRequest:
		var req packet.ReqPacketData
		if err := json.Unmarshal(data, &req); err != nil {
			return packet.Packet{}, errorx.Wrap(err, "ERR_CODEC_DECODE_FAILED", errorx.LevelUnexpected, "JSONCodecError", "failed to decode request packet", nil)
		}
		return packet.Packet{Opcode: packet.PacketOpcodeRequest, Req: &req}, nil
	case packet.PacketOpcodeResponse:
		var res packet.ResPacketData
		if err := json.Unmarshal(data, &res); err != nil {
			return packet.Packet{}, errorx.Wrap(err, "ERR_CODEC_DECODE_FAILED", errorx.LevelUnexpected, "JSONCodecError", "failed to decode response packet", nil)
		}
		return packet.Packet{Opcode: packet.PacketOpcodeResponse, Res: &res}, nil
	case packet.PacketOpcodeNotify:
		var ntfy packet.NotifyPacketData
		if err := json.Unmarshal(data, &ntfy); err != nil {
			return packet.Packet{}, errorx.Wrap(err, "ERR_CODEC_DECODE_FAILED", errorx.LevelUnexpected, "JSONCodecError", "failed to decode notify packet", nil)
		}
		return packet.Packet{Opcode: packet.PacketOpcodeNotify, Notify: &ntfy}, nil
	case packet.PacketOpcodeCommand:
		var cmd packet.CommandPacketData
		if err := json.Unmarshal(data, &cmd); err != nil {
			return packet.Packet{}, errorx.Wrap(err, "ERR_CODEC_DECODE_FAILED", errorx.LevelUnexpected, "JSONCodecError", "failed to decode command packet", nil)
		}
		return packet.Packet{Opcode: packet.PacketOpcodeCommand, Cmd: &cmd}, nil
	default:
		return packet.Packet{}, errorx.New("ERR_CODEC_UNSUPPORTED_OPCODE", errorx.LevelUnexpected, "JSONCodecError", "unsupported opcode", map[string]any{"opcode": header.Opcode})
	}
}

var _ rpc.Codec = (*JSONBufferCodec)(nil)
