package provider

import "time"

type CallOptions struct {
	Timeout  time.Duration
	TargetID string
}

type CallOption func(*CallOptions)

func WithTimeout(d time.Duration) CallOption {
	return func(o *CallOptions) {
		o.Timeout = d
	}
}

func WithTarget(serviceId string) CallOption {
	return func(o *CallOptions) {
		o.TargetID = serviceId
	}
}

func defaultCallOptions() CallOptions {
	return CallOptions{
		Timeout: 10 * time.Second,
	}
}
