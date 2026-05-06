package ram

import (
	"context"
	"sync"
)

type ramElection struct {
	name    string
	mu      sync.Mutex
	leader  *string
	waiters []chan struct{}
}

func (e *ramElection) Campaign(ctx context.Context, id string) error {
	e.mu.Lock()
	if e.leader == nil {
		e.leader = &id
		e.mu.Unlock()
		return nil
	}

	waitCh := make(chan struct{})
	e.waiters = append(e.waiters, waitCh)
	e.mu.Unlock()

	select {
	case <-waitCh:
		e.mu.Lock()
		e.leader = &id
		e.mu.Unlock()
		return nil
	case <-ctx.Done():
		e.mu.Lock()
		for i, ch := range e.waiters {
			if ch == waitCh {
				e.waiters = append(e.waiters[:i], e.waiters[i+1:]...)
				break
			}
		}
		e.mu.Unlock()
		return ctx.Err()
	}
}

func (e *ramElection) Resign(_ context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.leader == nil {
		return nil
	}
	e.leader = nil

	if len(e.waiters) > 0 {
		waitCh := e.waiters[0]
		e.waiters = e.waiters[1:]
		close(waitCh)
	}
	return nil
}

func (e *ramElection) Leader(_ context.Context) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.leader != nil {
		return *e.leader, nil
	}
	return "", nil
}

func (e *ramElection) Watch(ctx context.Context) <-chan string {
	ch := make(chan string, 8)

	e.mu.Lock()
	if e.leader != nil {
		ch <- *e.leader
	} else {
		ch <- ""
	}
	e.mu.Unlock()

	return ch
}
