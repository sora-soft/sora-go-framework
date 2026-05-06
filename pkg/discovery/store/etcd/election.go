package etcd

import (
	"context"
	"sync"

	"github.com/sora-soft/sora-go-framework.git/pkg/discovery"
	"github.com/sora-soft/sora-go-framework.git/pkg/utility/errorx"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type etcdElection struct {
	client      *clientv3.Client
	electionKey string

	mu        sync.Mutex
	leader    *string
	waiters   []chan struct{}
	campaigns map[string]bool
}

func (e *etcdElection) Campaign(ctx context.Context, id string) error {
	e.mu.Lock()
	if e.leader == nil {
		e.leader = &id
		if e.campaigns == nil {
			e.campaigns = make(map[string]bool)
		}
		e.campaigns[id] = true
		e.mu.Unlock()
		return e.persistLeader(ctx, id)
	}

	waitCh := make(chan struct{})
	e.waiters = append(e.waiters, waitCh)
	e.mu.Unlock()

	select {
	case <-waitCh:
		e.mu.Lock()
		e.leader = &id
		if e.campaigns == nil {
			e.campaigns = make(map[string]bool)
		}
		e.campaigns[id] = true
		e.mu.Unlock()
		return e.persistLeader(ctx, id)
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

func (e *etcdElection) persistLeader(ctx context.Context, id string) error {
	if e.client == nil {
		return newNotConnectedError()
	}
	_, err := e.client.Put(ctx, e.electionKey, id)
	return err
}

func (e *etcdElection) Resign(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.leader == nil {
		return nil
	}

	if e.client != nil {
		_, _ = e.client.Delete(ctx, e.electionKey)
	}

	e.leader = nil

	if len(e.waiters) > 0 {
		waitCh := e.waiters[0]
		e.waiters = e.waiters[1:]
		close(waitCh)
	}
	return nil
}

func (e *etcdElection) Leader(_ context.Context) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.leader != nil {
		return *e.leader, nil
	}

	if e.client != nil {
		resp, err := e.client.Get(context.Background(), e.electionKey)
		if err != nil {
			return "", errorx.Wrap(err, "ERR_ETCD_OPERATION", errorx.LevelUnexpected, "ETCDElectionError", "failed to get leader", nil)
		}
		if len(resp.Kvs) > 0 {
			val := string(resp.Kvs[0].Value)
			e.leader = &val
			return val, nil
		}
	}

	return "", nil
}

func (e *etcdElection) Watch(ctx context.Context) <-chan string {
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

var _ discovery.Election = (*etcdElection)(nil)
