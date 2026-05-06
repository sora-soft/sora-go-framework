package component

import (
	"context"
	"sync"
)

type baseComponent struct {
	componentImpl
	Name     string
	ready    bool
	refCount int
	mu       sync.Mutex
}

func NewBaseComponent(name string, impl componentImpl) *baseComponent {
	return &baseComponent{
		componentImpl: impl,
		Name:          name,
		ready:         false,
		refCount:      0,
	}
}

func (b *baseComponent) Start(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.refCount > 0 {
		b.refCount++
		return nil
	}

	if err := b.componentImpl.Connect(ctx); err != nil {
		return err
	}

	b.ready = true
	b.refCount = 1
	return nil
}

func (b *baseComponent) Stop() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.refCount <= 0 {
		return nil
	}

	b.refCount--
	if b.refCount > 0 {
		return nil
	}

	err := b.componentImpl.Disconnect()
	b.ready = false
	return err
}

func (b *baseComponent) LoadOptions(opts any) error {
	return b.componentImpl.SetOptions(opts)
}

func (b *baseComponent) GetMetaInfo() ComponentMetadata {
	b.mu.Lock()
	defer b.mu.Unlock()

	return ComponentMetadata{
		Name:    b.Name,
		Ready:   b.ready,
		Version: b.componentImpl.GetVersion(),
		Options: b.componentImpl.GetOptions(),
	}
}
