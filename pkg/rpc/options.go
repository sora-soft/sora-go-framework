package rpc

import "time"

type ConnectorPingOptions struct {
	Enabled  bool
	Timeout  time.Duration
	Interval time.Duration
}

type ConnectorOptions struct {
	Ping ConnectorPingOptions
}
