package packet

import "encoding/json"

type ResPacketData struct {
	Opcode  PacketOpcode      `json:"opcode"`
	Headers map[string]string `json:"headers"`
	Payload json.RawMessage   `json:"payload"`
}

func NewResPacket(headers map[string]string, payload json.RawMessage) *ResPacketData {
	return &ResPacketData{
		Opcode:  PacketOpcodeResponse,
		Headers: headers,
		Payload: payload,
	}
}
