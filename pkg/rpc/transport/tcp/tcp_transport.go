package tcp

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"context"
	"encoding/binary"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/sora-soft/sora-go-framework.git/pkg/rpc"
	"github.com/sora-soft/sora-go-framework.git/pkg/utility/errorx"
)

const (
	maxRetries     = 3
	initialDelay   = 500 * time.Millisecond
	maxDelay       = 8 * time.Second
	connectTimeout = 5 * time.Second
)

type TCPTransport struct {
	conn   net.Conn
	reader *bufio.Reader
	mu     sync.Mutex
}

func NewTCPTransport() *TCPTransport {
	return &TCPTransport{}
}

func NewServerTCPTransport(conn net.Conn) *TCPTransport {
	return &TCPTransport{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}
}

func (t *TCPTransport) Connect(ctx context.Context, endpoint string, codec string) (string, error) {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		confirmed, err := t.dialAndHandshake(ctx, endpoint, codec)
		if err == nil {
			return confirmed, nil
		}
		lastErr = err

		if attempt < maxRetries-1 {
			delay := t.backoffDelay(attempt)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
	}
	return "", lastErr
}

func (t *TCPTransport) dialAndHandshake(ctx context.Context, endpoint string, codec string) (string, error) {
	connectCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	dialer := net.Dialer{}
	socket, err := dialer.DialContext(connectCtx, "tcp", endpoint)
	if err != nil {
		return "", err
	}

	t.conn = socket
	t.reader = bufio.NewReader(socket)

	t.mu.Lock()
	_, err = socket.Write([]byte(codec + "\n"))
	t.mu.Unlock()
	if err != nil {
		socket.Close()
		t.conn = nil
		t.reader = nil
		return "", err
	}

	line, err := t.reader.ReadString('\n')
	if err != nil {
		socket.Close()
		t.conn = nil
		t.reader = nil
		return "", err
	}

	confirmedCodec := strings.TrimSpace(line)
	if confirmedCodec != codec {
		socket.Close()
		t.conn = nil
		t.reader = nil
		return "", errorx.New("ERR_TCP_CODEC_MISMATCH", errorx.LevelUnexpected, "TCPError", "server confirmed codec does not match requested codec", nil)
	}

	return confirmedCodec, nil
}

func (t *TCPTransport) backoffDelay(attempt int) time.Duration {
	delay := initialDelay * time.Duration(1<<uint(attempt))
	if delay > maxDelay {
		delay = maxDelay
	}
	return delay
}

func (t *TCPTransport) Handshake(ctx context.Context) (string, error) {
	if t.conn == nil || t.reader == nil {
		return "", net.ErrClosed
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	line, err := t.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	codec := strings.TrimSpace(line)

	t.mu.Lock()
	_, err = t.conn.Write([]byte(codec + "\n"))
	t.mu.Unlock()
	if err != nil {
		return "", err
	}

	return codec, nil
}

func (t *TCPTransport) Send(ctx context.Context, data []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.conn == nil {
		return net.ErrClosed
	}

	compressed, err := zlibCompress(data)
	if err != nil {
		return err
	}

	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(compressed)))
	if _, err := t.conn.Write(lenBuf); err != nil {
		return err
	}
	_, err = t.conn.Write(compressed)
	return err
}

func (t *TCPTransport) Recv(ctx context.Context) ([]byte, error) {
	if t.reader == nil {
		return nil, net.ErrClosed
	}

	var lenBuf [4]byte
	if _, err := io.ReadFull(t.reader, lenBuf[:]); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lenBuf[:])

	compressed := make([]byte, length)
	if _, err := io.ReadFull(t.reader, compressed); err != nil {
		return nil, err
	}
	return zlibDecompress(compressed)
}

func (t *TCPTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.conn != nil {
		err := t.conn.Close()
		t.conn = nil
		t.reader = nil
		return err
	}
	return nil
}

func zlibCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		w.Close()
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func zlibDecompress(data []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

var _ rpc.Transport = (*TCPTransport)(nil)

func init() {
	rpc.RegisterTransport("tcp", func() rpc.Transport {
		return NewTCPTransport()
	}, rpc.ConnectorOptions{
		Ping: rpc.ConnectorPingOptions{
			Enabled:  true,
			Timeout:  5 * time.Second,
			Interval: 10 * time.Second,
		},
	})
}
