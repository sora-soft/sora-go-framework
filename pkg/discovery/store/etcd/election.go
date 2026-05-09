package etcd

import (
	"context"
	"time"

	"github.com/sora-soft/sora-go-framework.git/pkg/discovery"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type etcdElection struct {
	client      *clientv3.Client
	electionKey string

	session  *concurrency.Session
	election *concurrency.Election
}

func (e *etcdElection) ensureSession(ctx context.Context) error {
	if e.session != nil {
		return nil
	}

	session, err := concurrency.NewSession(e.client)
	if err != nil {
		return err
	}
	e.session = session
	e.election = concurrency.NewElection(session, e.electionKey)
	return nil
}

func (e *etcdElection) Campaign(ctx context.Context, id string) error {
	if err := e.ensureSession(ctx); err != nil {
		return err
	}
	return e.election.Campaign(ctx, id)
}

func (e *etcdElection) Resign(ctx context.Context) error {
	if e.election == nil {
		return nil
	}
	err := e.election.Resign(ctx)
	if e.session != nil {
		e.session.Close()
		e.session = nil
		e.election = nil
	}
	return err
}

func (e *etcdElection) Leader(ctx context.Context) (string, error) {
	if e.election == nil {
		return "", nil
	}
	resp, err := e.election.Leader(ctx)
	if err != nil {
		return "", nil
	}
	return string(resp.Kvs[0].Value), nil
}

func (e *etcdElection) Watch(ctx context.Context) <-chan string {
	ch := make(chan string, 8)

	go func() {
		defer close(ch)

		for {
			current, err := e.Leader(ctx)
			if err != nil {
				select {
				case ch <- "":
				case <-ctx.Done():
					return
				}
			} else {
				select {
				case ch <- current:
				case <-ctx.Done():
					return
				}
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Second):
			}
		}
	}()

	return ch
}

var _ discovery.Election = (*etcdElection)(nil)
