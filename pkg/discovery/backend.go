package discovery

import "context"

type Backend interface {
	Connect(ctx context.Context) error
	Disconnect() error
	Registry() Registry
	Discovery() Discovery
	NewElection(name string) Election
	GetInfo() BackendInfo
}
