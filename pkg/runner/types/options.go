package types

import "github.com/sora-soft/sora-go-framework.git/pkg/utility"

type WorkerOptions struct {
	Alias *string `json:"alias,omitempty"`
}

type ServiceOptions struct {
	WorkerOptions
	Labels utility.Labels `json:"labels,omitempty"`
}
