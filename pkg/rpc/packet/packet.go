package packet

type PayloadCodec interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}

type PacketOpcode int

const (
	PacketOpcodeRequest  PacketOpcode = 1
	PacketOpcodeResponse PacketOpcode = 2
	PacketOpcodeNotify   PacketOpcode = 3
	PacketOpcodeCommand  PacketOpcode = 4
)

type PayloadError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Level   int    `json:"level"`
	Name    string `json:"name"`
	Args    any    `json:"args"`
}

type Response[T any] struct {
	Error  *PayloadError `json:"error"`
	Result T             `json:"result"`
}

type ConnectorCommand string

const (
	ConnectorCommandPing  ConnectorCommand = "ping"
	ConnectorCommandPong  ConnectorCommand = "pong"
	ConnectorCommandError ConnectorCommand = "error"
	ConnectorCommandOff   ConnectorCommand = "off"
	ConnectorCommandClose ConnectorCommand = "close"
)

type Packet struct {
	Opcode  PacketOpcode
	Method  string
	Service string
	Headers map[string]string
	payload []byte
	codec   PayloadCodec
}

func Decode[T any](p Packet) (T, error) {
	var zero T
	if p.codec == nil {
		return zero, nil
	}
	var v T
	if err := p.codec.Unmarshal(p.payload, &v); err != nil {
		return zero, err
	}
	return v, nil
}

func (p Packet) Payload() []byte {
	return p.payload
}

func (p Packet) Codec() PayloadCodec {
	return p.codec
}

func NewDecodedPacket(opcode PacketOpcode, method string, service string, headers map[string]string, payload []byte, codec PayloadCodec) Packet {
	return Packet{
		Opcode:  opcode,
		Method:  method,
		Service: service,
		Headers: headers,
		payload: payload,
		codec:   codec,
	}
}

func NewRequest[T any](codec PayloadCodec, method string, service string, headers map[string]string, req T) (Packet, error) {
	payload, err := codec.Marshal(req)
	if err != nil {
		return Packet{}, err
	}
	return Packet{
		Opcode:  PacketOpcodeRequest,
		Method:  method,
		Service: service,
		Headers: headers,
		payload: payload,
		codec:   codec,
	}, nil
}

func NewResponse[T any](codec PayloadCodec, headers map[string]string, resp T) (Packet, error) {
	payload, err := codec.Marshal(resp)
	if err != nil {
		return Packet{}, err
	}
	return Packet{
		Opcode:  PacketOpcodeResponse,
		Headers: headers,
		payload: payload,
		codec:   codec,
	}, nil
}

func NewNotify[T any](codec PayloadCodec, method string, service string, headers map[string]string, notify T) (Packet, error) {
	payload, err := codec.Marshal(notify)
	if err != nil {
		return Packet{}, err
	}
	return Packet{
		Opcode:  PacketOpcodeNotify,
		Method:  method,
		Service: service,
		Headers: headers,
		payload: payload,
		codec:   codec,
	}, nil
}

func NewCommandPacket(command ConnectorCommand, args []byte) Packet {
	return Packet{
		Opcode: PacketOpcodeCommand,
		Method: string(command),
		payload: args,
	}
}
