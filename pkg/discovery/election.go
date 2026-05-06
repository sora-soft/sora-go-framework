package discovery

import "context"

type Election interface {
	Campaign(ctx context.Context, id string) error
	Resign(ctx context.Context) error
	Leader(ctx context.Context) (string, error)
	Watch(ctx context.Context) <-chan string
}
