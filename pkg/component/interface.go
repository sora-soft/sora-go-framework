package component

import "context"

type Component interface {
	Start(ctx context.Context) error
	Stop() error
	LoadOptions(opts any) error
	GetMetaInfo() ComponentMetadata
}

type componentImpl interface {
	Connect(ctx context.Context) error
	Disconnect() error
	SetOptions(opts any) error
	GetOptions() any
	GetVersion() string
}

type ComponentMetadata struct {
	Name    string `json:"name"`
	Ready   bool   `json:"ready"`
	Version string `json:"version"`
	Options any    `json:"options,omitempty"`
}
