package etcd

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func (e *EtcdComponent) grantLease(ctx context.Context) error {
	lease := clientv3.NewLease(e.client)
	resp, err := lease.Grant(ctx, e.options.TTL)
	if err != nil {
		lease.Close()
		return err
	}
	e.lease = lease
	e.leaseID = resp.ID
	return nil
}

func (e *EtcdComponent) startKeepAlive() {
	e.keepAliveCtx, e.keepAliveCancel = context.WithCancel(context.Background())
	ch, err := e.lease.KeepAlive(e.keepAliveCtx, e.leaseID)
	if err != nil {
		etcdLogger().Error("EtcdComponent", err, "keepalive start failed")
		go e.reconnect(err)
		return
	}

	go func() {
		for range ch {
		}
		etcdLogger().Warn("EtcdComponent", "lease keepalive channel closed, triggering reconnect")
		go e.reconnect(nil)
	}()
}
