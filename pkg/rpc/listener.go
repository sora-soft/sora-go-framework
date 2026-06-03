package rpc

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/sora-soft/sora-go-framework/pkg/rpc/packet"
	"github.com/sora-soft/sora-go-framework/pkg/utility"
	"github.com/sora-soft/sora-go-framework/pkg/utility/errorx"
)

type ListenerInfo struct {
	Protocol string   `json:"protocol" yaml:"protocol"`
	Endpoint string   `json:"endpoint" yaml:"endpoint"`
	Codecs   []string `json:"codecs" yaml:"codecs"`
	Labels   utility.Labels
}

type ListenerState int

const (
	ListenerStateInit     ListenerState = 1
	ListenerStateStarting ListenerState = 2
	ListenerStateReady    ListenerState = 3
	ListenerStateStopping ListenerState = 4
	ListenerStateStopped  ListenerState = 5
	ListenerStateError    ListenerState = 100
)

func (s ListenerState) String() string {
	switch s {
	case ListenerStateInit:
		return "Init"
	case ListenerStateStarting:
		return "Starting"
	case ListenerStateReady:
		return "Ready"
	case ListenerStateStopping:
		return "Stopping"
	case ListenerStateStopped:
		return "Stopped"
	case ListenerStateError:
		return "Error"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}

type ListenerCallbacks struct {
	OnRequest      func(conn *Connection, pkt packet.Packet)
	OnNotify       func(conn *Connection, pkt packet.Packet)
	OnSessionOpen  func(conn *Connection, sessionId string)
	OnSessionClose func(conn *Connection, sessionId string)
}

type ListenerMetaInfo struct {
	Id       string         `json:"id" yaml:"id"`
	Protocol string         `json:"protocol" yaml:"protocol"`
	Endpoint string         `json:"endpoint" yaml:"endpoint"`
	State    ListenerState  `json:"state" yaml:"state"`
	Labels   utility.Labels `json:"labels" yaml:"labels"`
	Codecs   []string       `json:"codecs" yaml:"codecs"`
}

type Listener struct {
	id        string
	tl        TransportListener
	labels    utility.Labels
	codecs    []Codec
	callbacks ListenerCallbacks

	sessions  map[string]*Connection
	sessionMu sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc

	LifeCycle *utility.LifeCycle[ListenerState]
}

func NewListener(tl TransportListener, labels utility.Labels, codecs []Codec, callbacks ListenerCallbacks) *Listener {
	return &Listener{
		id:        uuid.New().String(),
		tl:        tl,
		labels:    labels,
		codecs:    codecs,
		callbacks: callbacks,
		sessions:  make(map[string]*Connection),
		LifeCycle: utility.NewLifeCycle(ListenerStateInit, false),
	}
}

func (l *Listener) Id() string {
	return l.id
}

func (l *Listener) GetMetaInfo() ListenerMetaInfo {
	tlMeta := l.tl.GetMetaInfo()
	codecNames := make([]string, 0, len(l.codecs))
	for _, c := range l.codecs {
		codecNames = append(codecNames, c.GetCode())
	}
	return ListenerMetaInfo{
		Id:       l.id,
		Protocol: tlMeta.Protocol,
		Endpoint: tlMeta.Endpoint,
		State:    l.LifeCycle.GetState(),
		Labels:   l.labels,
		Codecs:   codecNames,
	}
}

func (l *Listener) Start(ctx context.Context) error {
	l.ctx, l.cancel = context.WithCancel(ctx)

	if err := l.LifeCycle.SetState(ListenerStateStarting); err != nil {
		l.LifeCycle.SetStateWithError(ListenerStateError, err)
		return err
	}

	if err := l.tl.StartListen(ctx); err != nil {
		l.LifeCycle.SetStateWithError(ListenerStateError, err)
		return err
	}

	go l.acceptLoop()

	if err := l.LifeCycle.SetState(ListenerStateReady); err != nil {
		l.LifeCycle.SetStateWithError(ListenerStateError, err)
		return err
	}

	return nil
}

func (l *Listener) acceptLoop() {
	defer func() {
		if r := recover(); r != nil {
			FrameLogger.Error("listener", fmt.Errorf("%v", r), map[string]any{"event": "goroutine-panic", "recover": r})
		}
	}()
	for {
		conn, err := l.tl.Accept(l.ctx)
		if err != nil {
			if l.ctx.Err() != nil {
				return
			}
			continue
		}

		conn.OnRequest = l.callbacks.OnRequest
		conn.OnNotify = l.callbacks.OnNotify

		sessionId := uuid.New().String()

		if err := conn.Serve(l); err != nil {
			if l.callbacks.OnSessionClose != nil {
				l.callbacks.OnSessionClose(conn, sessionId)
			}
			continue
		}

		l.newConnector(sessionId, conn)

		if l.callbacks.OnSessionOpen != nil {
			l.callbacks.OnSessionOpen(conn, sessionId)
		}
	}
}

func (l *Listener) newConnector(sessionId string, conn *Connection) {
	l.sessionMu.Lock()
	l.sessions[sessionId] = conn
	l.sessionMu.Unlock()

	stateCh := conn.LifeCycle.Listen()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				FrameLogger.Error("listener", fmt.Errorf("%v", r), map[string]any{"event": "goroutine-panic", "recover": r, "session": sessionId})
			}
		}()
		for state := range stateCh {
			if state == ConnectorStateError || state == ConnectorStateStopped {
				l.sessionMu.Lock()
				delete(l.sessions, sessionId)
				l.sessionMu.Unlock()

				RpcLogger.Info("listener", map[string]any{"event": "session-closed", "session": sessionId})

				if l.callbacks.OnSessionClose != nil {
					l.callbacks.OnSessionClose(conn, sessionId)
				}
				conn.LifeCycle.RemoveListen(stateCh)
				return
			}
		}
	}()
}

func (l *Listener) Stop() error {
	if err := l.LifeCycle.SetState(ListenerStateStopping); err != nil {
		return err
	}

	if l.cancel != nil {
		l.cancel()
	}

	l.sessionMu.RLock()
	sessions := make([]*Connection, 0, len(l.sessions))
	for _, conn := range l.sessions {
		sessions = append(sessions, conn)
	}
	l.sessionMu.RUnlock()

	for _, conn := range sessions {
		_ = conn.SendCommand(context.Background(), packet.NewCommandPacket(packet.ConnectorCommandOff, nil))
	}

	l.tl.Close()

	if err := l.LifeCycle.SetState(ListenerStateStopped); err != nil {
		return err
	}

	FrameLogger.Info("listener", map[string]any{"event": "listener-stopped", "id": l.id})
	return nil
}

func (l *Listener) CloseSession(sessionId string) error {
	l.sessionMu.Lock()
	conn, ok := l.sessions[sessionId]
	if !ok {
		l.sessionMu.Unlock()
		return errorx.New("ERR_SESSION_NOT_FOUND", errorx.LevelExpected, "ListenerError", "session not found", map[string]any{"sessionId": sessionId})
	}
	delete(l.sessions, sessionId)
	l.sessionMu.Unlock()

	return conn.Disconnect()
}

func (l *Listener) GetSession(sessionId string) (*Connection, bool) {
	l.sessionMu.RLock()
	defer l.sessionMu.RUnlock()
	conn, ok := l.sessions[sessionId]
	return conn, ok
}
