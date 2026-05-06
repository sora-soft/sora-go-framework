package packet

type PacketOpcode int

const (
	PacketOpcodeRequest  PacketOpcode = 1
	PacketOpcodeResponse PacketOpcode = 2
	PacketOpcodeNotify   PacketOpcode = 3
	PacketOpcodeCommand  PacketOpcode = 4
)

type Packet struct {
	Opcode PacketOpcode
	Req    *ReqPacketData
	Notify *NotifyPacketData
	Res    *ResPacketData
	Cmd    *CommandPacketData
}
