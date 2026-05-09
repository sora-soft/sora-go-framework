package etcd

import (
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func (e *EtcdComponent) reconnect(cause error) {
	e.reconnectMu.Lock()
	if e.reconnecting {
		e.reconnectMu.Unlock()
		return
	}
	e.reconnecting = true
	e.reconnectMu.Unlock()

	defer func() {
		e.reconnectMu.Lock()
		e.reconnecting = false
		e.reconnectMu.Unlock()
	}()

	etcdLogger().Warn("EtcdComponent", "starting reconnect loop")

	interval := 100 * time.Millisecond
	for {
		if e.destroyed {
			etcdLogger().Warn("EtcdComponent", "reconnect stopped: component destroyed")
			return
		}

		if e.lease != nil && e.leaseID != 0 {
			_, _ = e.lease.Revoke(context.Background(), e.leaseID)
			e.lease = nil
			e.leaseID = 0
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := e.grantLease(ctx)
		cancel()
		if err != nil {
			etcdLogger().Error("EtcdComponent", err, "reconnect grant lease failed, retrying")
			time.Sleep(interval)
			interval = min(interval*2, 30*time.Second)
			continue
		}

		e.startKeepAlive()

		if err := e.restorePersistValues(context.Background()); err != nil {
			etcdLogger().Error("EtcdComponent", err, "reconnect restore persist values failed")
		}

		etcdLogger().Success("EtcdComponent", "reconnect successful")

		for _, fn := range e.onLeaseReconnect {
			fn(e.leaseID, cause)
		}

		return
	}
}

func (e *EtcdComponent) restorePersistValues(ctx context.Context) error {
	if len(e.persistValues) == 0 {
		return nil
	}
	for key, value := range e.persistValues {
		_, err := e.client.Put(ctx, key, value, clientv3.WithLease(e.leaseID))
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *EtcdComponent) PersistPut(ctx context.Context, key string, value string) error {
	if e.leaseID == 0 {
		return newLeaseNotFoundError()
	}

	_, err := e.client.Put(ctx, key, value, clientv3.WithLease(e.leaseID))
	if err != nil {
		return err
	}

	e.persistValues[key] = value
	return nil
}

func (e *EtcdComponent) PersistDel(ctx context.Context, key string) error {
	if e.client == nil {
		return newNotConnectedError()
	}

	_, err := e.client.Delete(ctx, key)
	if err != nil {
		return err
	}

	delete(e.persistValues, key)
	return nil
}
