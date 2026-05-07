package provider

type NotifyOptions struct {
	TargetID string
}

type NotifyOption func(*NotifyOptions)

func WithNotifyTarget(targetID string) NotifyOption {
	return func(o *NotifyOptions) {
		o.TargetID = targetID
	}
}
