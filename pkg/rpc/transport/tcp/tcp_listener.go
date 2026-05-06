package tcp

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/sora-soft/sora-go-framework.git/pkg/rpc"
	"github.com/sora-soft/sora-go-framework.git/pkg/utility/errorx"
)

type TCPListenerOptions struct {
	Host      string
	Port      int
	PortRange []int
}

type TCPListener struct {
	listener net.Listener
	connWg   sync.WaitGroup
}

func NewTCPListener(opts TCPListenerOptions) (*TCPListener, error) {
	if opts.Port != 0 && len(opts.PortRange) > 0 {
		return nil, errorx.New("ERR_TCP_OPTIONS_CONFLICT", errorx.LevelExpected, "TCPListenerError", "Port and PortRange are mutually exclusive", nil)
	}
	if opts.Port == 0 && len(opts.PortRange) == 0 {
		return nil, errorx.New("ERR_TCP_OPTIONS_REQUIRED", errorx.LevelExpected, "TCPListenerError", "either Port or PortRange must be set", nil)
	}

	var listener net.Listener
	var err error

	if opts.Port != 0 {
		addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
		listener, err = net.Listen("tcp", addr)
		if err != nil {
			return nil, err
		}
	} else {
		min := opts.PortRange[0]
		max := opts.PortRange[1]
		current := min
		for {
			addr := fmt.Sprintf("%s:%d", opts.Host, current)
			listener, err = net.Listen("tcp", addr)
			if err == nil {
				break
			}
			step := rand.Intn(5) + 1
			current += step
			if current > max {
				return nil, errorx.New("ERR_TCP_NO_AVAILABLE_PORT", errorx.LevelUnexpected, "TCPListenerError", "no available port in range", map[string]any{"min": min, "max": max})
			}
		}
	}

	return &TCPListener{
		listener: listener,
	}, nil
}

func (l *TCPListener) Accept(ctx context.Context) (*rpc.Connection, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}

	transport := NewServerTCPTransport(conn)
	rpcConn := rpc.NewConnection(transport, rpc.ConnectorOptions{
		Ping: rpc.ConnectorPingOptions{Enabled: false},
	})

	l.connWg.Add(1)
	stateCh := rpcConn.LifeCycle.Listen()
	go func() {
		defer l.connWg.Done()
		for state := range stateCh {
			if state == rpc.ConnectorStateError || state == rpc.ConnectorStateStopped {
				rpcConn.LifeCycle.RemoveListen(stateCh)
				return
			}
		}
	}()

	return rpcConn, nil
}

func (l *TCPListener) Close() error {
	err := l.listener.Close()

	done := make(chan struct{})
	go func() {
		l.connWg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
	}

	return err
}

func (l *TCPListener) GetMetaInfo() rpc.TransportMetaInfo {
	return rpc.TransportMetaInfo{
		Protocol: "tcp",
		Endpoint: l.listener.Addr().String(),
	}
}

func (l *TCPListener) Addr() string {
	return l.listener.Addr().String()
}

var _ rpc.TransportListener = (*TCPListener)(nil)
