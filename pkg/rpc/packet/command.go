package packet

import "encoding/json"

type ConnectorCommand string

const (
	ConnectorCommandPing  ConnectorCommand = "ping"
	ConnectorCommandPong  ConnectorCommand = "pong"
	ConnectorCommandError ConnectorCommand = "error"
	ConnectorCommandOff   ConnectorCommand = "off"
	ConnectorCommandClose ConnectorCommand = "close"
)

type CommandPacketData struct {
	Opcode  PacketOpcode     `json:"opcode"`
	Command ConnectorCommand `json:"command"`
	Args    json.RawMessage  `json:"args"`
}

func NewCommandPacket(command ConnectorCommand, args json.RawMessage) *CommandPacketData {
	return &CommandPacketData{
		Opcode:  PacketOpcodeCommand,
		Command: command,
		Args:    args,
	}
}
