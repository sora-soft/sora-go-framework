package packet

import "encoding/json"

type ReqPacketData struct {
	Opcode  PacketOpcode      `json:"opcode"`
	Method  string            `json:"method"`
	Service string            `json:"service"`
	Headers map[string]string `json:"headers"`
	Payload json.RawMessage   `json:"payload"`
}

func NewReqPacket(method string, service string, headers map[string]string, payload json.RawMessage) *ReqPacketData {
	return &ReqPacketData{
		Opcode:  PacketOpcodeRequest,
		Method:  method,
		Service: service,
		Headers: headers,
		Payload: payload,
	}
}
