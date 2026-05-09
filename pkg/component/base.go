package component

import (
	"context"
	"sync"
)

type BaseComponent[T ComponentImpl] struct {
	impl     T
	Name     ComponentName
	ready    bool
	refCount int
	mu       sync.Mutex
}

func NewBaseComponent[T ComponentImpl](name ComponentName, impl T) *BaseComponent[T] {
	return &BaseComponent[T]{
		impl:     impl,
		Name:     name,
		ready:    false,
		refCount: 0,
	}
}

func (b *BaseComponent[T]) Start(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.refCount > 0 {
		b.refCount++
		return nil
	}

	if err := b.impl.Connect(ctx); err != nil {
		return err
	}

	b.ready = true
	b.refCount = 1
	return nil
}

func (b *BaseComponent[T]) Stop() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.refCount <= 0 {
		return nil
	}

	b.refCount--
	if b.refCount > 0 {
		return nil
	}

	err := b.impl.Disconnect()
	b.ready = false
	return err
}

func (b *BaseComponent[T]) LoadOptions(opts any) error {
	return b.impl.SetOptions(opts)
}

func (b *BaseComponent[T]) GetMetaInfo() ComponentMetadata {
	b.mu.Lock()
	defer b.mu.Unlock()

	return ComponentMetadata{
		Name:    string(b.Name),
		Ready:   b.ready,
		Version: b.impl.GetVersion(),
		Options: b.impl.GetOptions(),
	}
}

func (b *BaseComponent[T]) Impl() T {
	return b.impl
}
