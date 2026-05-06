package utility

import (
	"sync"

	"github.com/sora-soft/sora-go-framework.git/pkg/utility/errorx"
)

type State interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

type listener[T State] struct {
	ch   chan T
	done chan struct{}
}

type LifeCycle[T State] struct {
	state         T
	backTrackable bool
	listeners     []listener[T]
	mu            sync.RWMutex
	lastError     error
}

func NewLifeCycle[T State](state T, backTrackable bool) *LifeCycle[T] {
	return &LifeCycle[T]{
		state:         state,
		backTrackable: backTrackable,
	}
}

func (lc *LifeCycle[T]) SetState(state T) error {
	if lc.state > state && !lc.backTrackable {
		return errorx.New("ERR_LIFECYCLE_STATE", errorx.LevelUnexpected, "LifeCycleError", "lifecycle state can not back track", nil)
	}

	lc.mu.Lock()
	defer lc.mu.Unlock()

	if lc.state == state {
		return nil
	}
	lc.state = state

	for _, listener := range lc.listeners {
		select {
		case listener.ch <- state:
		default:
		}
	}

	return nil
}

func (lc *LifeCycle[T]) SetStateWithError(state T, err error) error {
	if e := lc.SetState(state); e != nil {
		return e
	}
	lc.mu.Lock()
	lc.lastError = err
	lc.mu.Unlock()
	return nil
}

func (lc *LifeCycle[T]) GetLastError() error {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return lc.lastError
}

func (lc *LifeCycle[T]) GetState() T {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return lc.state
}

func (lc *LifeCycle[T]) Listen() chan T {
	ch := make(chan T, 1)
	lc.mu.Lock()
	defer lc.mu.Unlock()

	lc.listeners = append(lc.listeners, listener[T]{ch: ch})
	return ch
}

func (lc *LifeCycle[T]) RemoveListen(ch chan T) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	for i, listener := range lc.listeners {
		if listener.ch == ch {
			lc.listeners = append(lc.listeners[:i], lc.listeners[i+1:]...)
			close(listener.ch) // 由发送方关闭
			return
		}
	}
}
