package packet

import "encoding/json"

type NotifyPacketData struct {
	Opcode  PacketOpcode      `json:"opcode"`
	Method  string            `json:"method"`
	Service string            `json:"service"`
	Headers map[string]string `json:"headers"`
	Payload json.RawMessage   `json:"payload"`
}

func NewNotifyPacket(method string, service string, headers map[string]string, payload json.RawMessage) *NotifyPacketData {
	return &NotifyPacketData{
		Opcode:  PacketOpcodeNotify,
		Method:  method,
		Service: service,
		Headers: headers,
		Payload: payload,
	}
}
