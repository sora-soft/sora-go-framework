package types

import "github.com/sora-soft/sora-go-framework/pkg/utility"

type WorkerOptions struct {
	Alias *string `json:"alias,omitempty" yaml:"alias,omitempty"`
}

type ServiceOptions struct {
	WorkerOptions
	Labels utility.Labels `json:"labels,omitempty" yaml:"labels,omitempty"`
}
