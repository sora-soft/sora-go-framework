package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/sora-soft/sora-go-framework.git/pkg/rpc/packet"
	"github.com/sora-soft/sora-go-framework.git/pkg/utility"
	"github.com/sora-soft/sora-go-framework.git/pkg/utility/errorx"
)

type ConnectorState int

const (
	ConnectorStateInit       ConnectorState = 1
	ConnectorStateConnecting ConnectorState = 2
	ConnectorStateReady      ConnectorState = 4
	ConnectorStateStopping   ConnectorState = 5
	ConnectorStateStopped    ConnectorState = 6
	ConnectorStateError      ConnectorState = 100
)

func (s ConnectorState) String() string {
	switch s {
	case ConnectorStateInit:
		return "Init"
	case ConnectorStateConnecting:
		return "Connecting"
	case ConnectorStateReady:
		return "Ready"
	case ConnectorStateStopping:
		return "Stopping"
	case ConnectorStateStopped:
		return "Stopped"
	case ConnectorStateError:
		return "Error"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}

type Connection struct {
	transport Transport
	options   ConnectorOptions

	ctx    context.Context
	cancel context.CancelFunc

	codec  Codec
	pinger *Pinger
	target ListenerInfo

	OnRequest func(conn *Connection, req *packet.ReqPacketData)
	OnNotify  func(conn *Connection, notify *packet.NotifyPacketData)

	mu sync.Mutex

	LifeCycle *utility.LifeCycle[ConnectorState]
}

func NewConnection(transport Transport, options ConnectorOptions) *Connection {
	c := &Connection{
		transport: transport,
		LifeCycle: utility.NewLifeCycle(ConnectorStateInit, false),
		options:   options,
	}

	stateCh := c.LifeCycle.Listen()
	go func() {
		for state := range stateCh {
			switch state {
			case ConnectorStateReady:
				c.enablePingPong()
			case ConnectorStateStopping:
				c.disablePingPong()
			case ConnectorStateStopped, ConnectorStateError:
				transport.Close()
				c.LifeCycle.RemoveListen(stateCh)
				return
			}
		}
	}()

	return c
}

func (c *Connection) Start(ctx context.Context, target ListenerInfo, codec Codec) error {
	c.mu.Lock()
	c.ctx, c.cancel = context.WithCancel(ctx)
	c.target = target
	c.codec = codec
	c.mu.Unlock()

	if err := c.LifeCycle.SetState(ConnectorStateConnecting); err != nil {
		c.LifeCycle.SetStateWithError(ConnectorStateError, err)
		return err
	}

	errCh := make(chan error, 1)
	var confirmedCodec string
	go func() {
		var err error
		confirmedCodec, err = c.transport.Connect(c.ctx, target.Endpoint, codec.GetCode())
		errCh <- err
	}()

	select {
	case err := <-errCh:
		if err != nil {
			c.LifeCycle.SetStateWithError(ConnectorStateError, err)
			return err
		}
	case <-c.ctx.Done():
		return c.ctx.Err()
	}

	if confirmedCodec != codec.GetCode() {
		err := errorx.New("ERR_CODEC_MISMATCH", errorx.LevelUnexpected, "ConnectorError", "confirmed codec does not match requested codec", nil)
		c.LifeCycle.SetStateWithError(ConnectorStateError, err)
		return err
	}

	go c.readLoop()

	if err := c.LifeCycle.SetState(ConnectorStateReady); err != nil {
		c.LifeCycle.SetStateWithError(ConnectorStateError, err)
		return err
	}

	return nil
}

func (c *Connection) Serve(listener *Listener) error {
	c.mu.Lock()
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.mu.Unlock()

	if err := c.LifeCycle.SetState(ConnectorStateConnecting); err != nil {
		c.LifeCycle.SetStateWithError(ConnectorStateError, err)
		return err
	}

	codecName, err := c.transport.Handshake(c.ctx)
	if err != nil {
		c.LifeCycle.SetStateWithError(ConnectorStateError, err)
		return err
	}

	var codec Codec
	for _, c := range listener.codecs {
		if c.GetCode() == codecName {
			codec = c
			break
		}
	}
	if codec == nil {
		err := errorx.New("ERR_CODEC_NOT_FOUND", errorx.LevelUnexpected, "ConnectorError", "codec not found in listener", map[string]any{"code": codecName})
		c.LifeCycle.SetStateWithError(ConnectorStateError, err)
		return err
	}

	c.mu.Lock()
	c.codec = codec
	c.mu.Unlock()

	go c.readLoop()

	if err := c.LifeCycle.SetState(ConnectorStateReady); err != nil {
		c.LifeCycle.SetStateWithError(ConnectorStateError, err)
		return err
	}

	return nil
}

func (c *Connection) readLoop() {
	for {
		data, err := c.transport.Recv(c.ctx)
		if err != nil {
			c.LifeCycle.SetStateWithError(ConnectorStateError, err)
			return
		}

		pkt, err := c.codec.Decode(data)
		if err != nil {
			c.LifeCycle.SetStateWithError(ConnectorStateError, err)
			return
		}

		if err := c.handlePacket(pkt); err != nil {
			c.LifeCycle.SetStateWithError(ConnectorStateError, err)
			return
		}
	}
}

func (c *Connection) handlePacket(pkt packet.Packet) error {
	switch pkt.Opcode {
	case packet.PacketOpcodeCommand:
		if pkt.Cmd == nil {
			return errorx.New("ERR_CONNECTION_INVALID_PACKET", errorx.LevelUnexpected, "RpcError", "command packet has nil Cmd field", nil)
		}
		return c.handleCommand(pkt.Cmd)
	case packet.PacketOpcodeRequest:
		if pkt.Req == nil {
			return errorx.New("ERR_CONNECTION_INVALID_PACKET", errorx.LevelUnexpected, "RpcError", "request packet has nil Req field", nil)
		}
		if c.OnRequest != nil {
			c.OnRequest(c, pkt.Req)
		}
	case packet.PacketOpcodeNotify:
		if pkt.Notify == nil {
			return errorx.New("ERR_CONNECTION_INVALID_PACKET", errorx.LevelUnexpected, "RpcError", "notify packet has nil Notify field", nil)
		}
		if c.OnNotify != nil {
			c.OnNotify(c, pkt.Notify)
		}
	}
	return nil
}

func (c *Connection) handleCommand(cmd *packet.CommandPacketData) error {
	switch cmd.Command {
	case packet.ConnectorCommandPing:
		var args struct {
			Id int `json:"id"`
		}
		if err := json.Unmarshal(cmd.Args, &args); err != nil {
			return err
		}
		return c.sendPong(args.Id)
	case packet.ConnectorCommandPong:
		var args struct {
			Id int `json:"id"`
		}
		if err := json.Unmarshal(cmd.Args, &args); err != nil {
			return err
		}
		c.mu.Lock()
		if c.pinger != nil {
			c.pinger.OnPong(args.Id, nil)
		}
		c.mu.Unlock()
	}
	return nil
}

func (c *Connection) SendRaw(ctx context.Context, pkt packet.Packet) error {
	data, err := c.codec.Encode(pkt)
	if err != nil {
		return err
	}
	return c.transport.Send(ctx, data)
}

func (c *Connection) SendRequest(ctx context.Context, req *packet.ReqPacketData) error {
	return c.SendRaw(ctx, packet.Packet{
		Opcode: packet.PacketOpcodeRequest,
		Req:    req,
	})
}

func (c *Connection) SendCommand(ctx context.Context, cmd *packet.CommandPacketData) error {
	return c.SendRaw(ctx, packet.Packet{
		Opcode: packet.PacketOpcodeCommand,
		Cmd:    cmd,
	})
}

func (c *Connection) SendResponse(ctx context.Context, res *packet.ResPacketData) error {
	return c.SendRaw(ctx, packet.Packet{
		Opcode: packet.PacketOpcodeResponse,
		Res:    res,
	})
}

func (c *Connection) Disconnect() error {
	c.mu.Lock()
	if c.cancel != nil {
		c.cancel()
	}
	c.mu.Unlock()

	if err := c.LifeCycle.SetState(ConnectorStateStopping); err != nil {
		return err
	}

	if err := c.LifeCycle.SetState(ConnectorStateStopped); err != nil {
		return err
	}

	return nil
}

func (c *Connection) sendPing(id int) error {
	args, err := json.Marshal(struct {
		Id int `json:"id"`
	}{Id: id})
	if err != nil {
		return err
	}
	return c.SendCommand(c.ctx, packet.NewCommandPacket(packet.ConnectorCommandPing, args))
}

func (c *Connection) sendPong(id int) error {
	args, err := json.Marshal(struct {
		Id int `json:"id"`
	}{Id: id})
	if err != nil {
		return err
	}
	return c.SendCommand(c.ctx, packet.NewCommandPacket(packet.ConnectorCommandPong, args))
}

func (c *Connection) enablePingPong() {
	p := NewPinger(func(u int) error {
		return c.sendPing(u)
	}, func(err error) {
		c.LifeCycle.SetStateWithError(ConnectorStateError, err)
	}, func() ConnectorState {
		return c.LifeCycle.GetState()
	}, c.options.Ping)

	p.Start()

	c.mu.Lock()
	c.pinger = p
	c.mu.Unlock()
}

func (c *Connection) disablePingPong() {
	c.mu.Lock()
	p := c.pinger
	c.pinger = nil
	c.mu.Unlock()

	if p != nil {
		p.Stop()
	}
}
