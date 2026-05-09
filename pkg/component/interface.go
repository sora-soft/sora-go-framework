package component

import "context"

type Component interface {
	Start(ctx context.Context) error
	Stop() error
	LoadOptions(opts any) error
	GetMetaInfo() ComponentMetadata
}

type ComponentImpl interface {
	Connect(ctx context.Context) error
	Disconnect() error
	SetOptions(opts any) error
	GetOptions() any
	GetVersion() string
}

type ComponentMetadata struct {
	Name    string `json:"name" yaml:"name"`
	Ready   bool   `json:"ready" yaml:"ready"`
	Version string `json:"version" yaml:"version"`
	Options any    `json:"options,omitempty" yaml:"options,omitempty"`
}

type ComponentName string
