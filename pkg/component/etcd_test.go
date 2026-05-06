package component

import (
	"context"
	"testing"
	"time"
)

func TestNewEtcdComponent(t *testing.T) {
	c := NewEtcdComponent("etcd-main")

	meta := c.GetMetaInfo()
	if meta.Name != "etcd-main" {
		t.Fatalf("expected name etcd-main, got %s", meta.Name)
	}
	if meta.Ready {
		t.Fatal("expected ready to be false")
	}
	if meta.Version != "0.1.0" {
		t.Fatalf("expected version 0.1.0, got %s", meta.Version)
	}
}

func TestEtcdComponent_SetOptions(t *testing.T) {
	impl := &EtcdComponent{version: "0.1.0"}

	err := impl.SetOptions(&EtcdOptions{
		Endpoints:   []string{"http://127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}

	opts := impl.GetOptions().(EtcdOptions)
	if len(opts.Endpoints) != 1 || opts.Endpoints[0] != "http://127.0.0.1:2379" {
		t.Fatalf("unexpected endpoints: %v", opts.Endpoints)
	}
}

func TestEtcdComponent_SetOptions_InvalidType(t *testing.T) {
	impl := &EtcdComponent{version: "0.1.0"}

	err := impl.SetOptions("not-valid")
	if err != ErrEtcdInvalidOptions {
		t.Fatalf("expected ErrEtcdInvalidOptions, got %v", err)
	}
}

func TestEtcdComponent_LoadOptions_Delegates(t *testing.T) {
	c := NewEtcdComponent("etcd-test")

	err := c.LoadOptions(&EtcdOptions{
		Endpoints:   []string{"http://127.0.0.1:2379"},
		DialTimeout: 3 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}

	meta := c.GetMetaInfo()
	opts, ok := meta.Options.(EtcdOptions)
	if !ok {
		t.Fatal("expected options to be EtcdOptions")
	}
	if opts.DialTimeout != 3*time.Second {
		t.Fatalf("expected dial timeout 3s, got %v", opts.DialTimeout)
	}
}

func TestEtcdComponent_Connect_WithoutOptions(t *testing.T) {
	impl := &EtcdComponent{version: "0.1.0"}

	err := impl.Connect(context.Background())
	if err != ErrEtcdOptionsNotSet {
		t.Fatalf("expected ErrEtcdOptionsNotSet, got %v", err)
	}
}

func TestEtcdComponent_Disconnect_WhenNil(t *testing.T) {
	impl := &EtcdComponent{version: "0.1.0"}

	err := impl.Disconnect()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestEtcdComponent_GetOptions_Nil(t *testing.T) {
	impl := &EtcdComponent{version: "0.1.0"}

	if opts := impl.GetOptions(); opts != nil {
		t.Fatalf("expected nil, got %v", opts)
	}
}

func TestEtcdComponent_SetOptions_Overwrites(t *testing.T) {
	impl := &EtcdComponent{version: "0.1.0"}

	_ = impl.SetOptions(&EtcdOptions{
		Endpoints:   []string{"http://host1:2379"},
		DialTimeout: 5 * time.Second,
	})

	_ = impl.SetOptions(&EtcdOptions{
		Endpoints:   []string{"http://host2:2379"},
		DialTimeout: 10 * time.Second,
	})

	opts := impl.GetOptions().(EtcdOptions)
	if len(opts.Endpoints) != 1 || opts.Endpoints[0] != "http://host2:2379" {
		t.Fatalf("expected host2, got %v", opts.Endpoints)
	}
	if opts.DialTimeout != 10*time.Second {
		t.Fatalf("expected 10s, got %v", opts.DialTimeout)
	}
}
