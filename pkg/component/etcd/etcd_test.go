package etcd

import (
	"context"
	"errors"
	"testing"

	"github.com/sora-soft/sora-go-framework.git/pkg/utility/errorx"
)

func validOptions() *EtcdComponentOptions {
	return &EtcdComponentOptions{
		Etcd: EtcdClientConfig{
			Hosts: []string{"http://127.0.0.1:2379"},
		},
		TTL:    60,
		Prefix: "/sora",
	}
}

func assertErrorCode(t *testing.T, err error, code string) {
	t.Helper()
	var exErr *errorx.Error
	if !errors.As(err, &exErr) {
		t.Fatalf("expected *errorx.Error, got %T: %v", err, err)
	}
	if exErr.Code != code {
		t.Fatalf("expected code %s, got %s", code, exErr.Code)
	}
}

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
	impl := &EtcdComponent{version: "0.1.0", persistValues: make(map[string]string)}

	err := impl.SetOptions(validOptions())
	if err != nil {
		t.Fatal(err)
	}

	opts := impl.GetOptions().(EtcdComponentOptions)
	if len(opts.Etcd.Hosts) != 1 || opts.Etcd.Hosts[0] != "http://127.0.0.1:2379" {
		t.Fatalf("unexpected hosts: %v", opts.Etcd.Hosts)
	}
	if opts.TTL != 60 {
		t.Fatalf("expected ttl 60, got %d", opts.TTL)
	}
	if opts.Prefix != "/sora" {
		t.Fatalf("expected prefix /sora, got %s", opts.Prefix)
	}
}

func TestEtcdComponent_SetOptions_InvalidType(t *testing.T) {
	impl := &EtcdComponent{version: "0.1.0", persistValues: make(map[string]string)}

	err := impl.SetOptions("not-valid")
	assertErrorCode(t, err, ErrCodeInvalidOptions)
}

func TestEtcdComponent_SetOptions_HostsEmpty(t *testing.T) {
	impl := &EtcdComponent{version: "0.1.0", persistValues: make(map[string]string)}

	opts := validOptions()
	opts.Etcd.Hosts = nil

	err := impl.SetOptions(opts)
	assertErrorCode(t, err, ErrCodeHostsEmpty)
}

func TestEtcdComponent_SetOptions_DialTimeoutZero(t *testing.T) {
	impl := &EtcdComponent{version: "0.1.0", persistValues: make(map[string]string)}

	opts := validOptions()

	err := impl.SetOptions(opts)
	assertErrorCode(t, err, ErrCodeDialTimeoutZero)
}

func TestEtcdComponent_SetOptions_TTLZero(t *testing.T) {
	impl := &EtcdComponent{version: "0.1.0", persistValues: make(map[string]string)}

	opts := validOptions()
	opts.TTL = 0

	err := impl.SetOptions(opts)
	assertErrorCode(t, err, ErrCodeTTLInvalid)
}

func TestEtcdComponent_SetOptions_TTLNegative(t *testing.T) {
	impl := &EtcdComponent{version: "0.1.0", persistValues: make(map[string]string)}

	opts := validOptions()
	opts.TTL = -1

	err := impl.SetOptions(opts)
	assertErrorCode(t, err, ErrCodeTTLInvalid)
}

func TestEtcdComponent_SetOptions_PrefixEmpty(t *testing.T) {
	impl := &EtcdComponent{version: "0.1.0", persistValues: make(map[string]string)}

	opts := validOptions()
	opts.Prefix = ""

	err := impl.SetOptions(opts)
	assertErrorCode(t, err, ErrCodePrefixEmpty)
}

func TestEtcdComponent_SetOptions_AuthOptional(t *testing.T) {
	impl := &EtcdComponent{version: "0.1.0", persistValues: make(map[string]string)}

	opts := validOptions()
	opts.Etcd.Auth = nil

	err := impl.SetOptions(opts)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestEtcdComponent_SetOptions_WithAuth(t *testing.T) {
	impl := &EtcdComponent{version: "0.1.0", persistValues: make(map[string]string)}

	opts := validOptions()
	opts.Etcd.Auth = &EtcdAuth{
		Username: "admin",
		Password: "secret",
	}

	err := impl.SetOptions(opts)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got := impl.GetOptions().(EtcdComponentOptions)
	if got.Etcd.Auth == nil || got.Etcd.Auth.Username != "admin" || got.Etcd.Auth.Password != "secret" {
		t.Fatalf("expected auth admin:secret, got %v", got.Etcd.Auth)
	}
}

func TestEtcdComponent_LoadOptions_Delegates(t *testing.T) {
	c := NewEtcdComponent("etcd-test")

	err := c.LoadOptions(validOptions())
	if err != nil {
		t.Fatal(err)
	}

	meta := c.GetMetaInfo()
	opts, ok := meta.Options.(EtcdComponentOptions)
	if !ok {
		t.Fatal("expected options to be EtcdComponentOptions")
	}
	if opts.TTL != 60 {
		t.Fatalf("expected ttl 60, got %d", opts.TTL)
	}
	if opts.Prefix != "/sora" {
		t.Fatalf("expected prefix /sora, got %s", opts.Prefix)
	}
}

func TestEtcdComponent_Connect_WithoutOptions(t *testing.T) {
	impl := &EtcdComponent{version: "0.1.0", persistValues: make(map[string]string)}

	err := impl.Connect(context.Background())
	assertErrorCode(t, err, ErrCodeOptionsNotSet)
}

func TestEtcdComponent_Disconnect_WhenNil(t *testing.T) {
	impl := &EtcdComponent{version: "0.1.0", persistValues: make(map[string]string)}

	err := impl.Disconnect()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestEtcdComponent_GetOptions_Nil(t *testing.T) {
	impl := &EtcdComponent{version: "0.1.0", persistValues: make(map[string]string)}

	if opts := impl.GetOptions(); opts != nil {
		t.Fatalf("expected nil, got %v", opts)
	}
}

func TestEtcdComponent_SetOptions_Overwrites(t *testing.T) {
	impl := &EtcdComponent{version: "0.1.0", persistValues: make(map[string]string)}

	opts1 := validOptions()
	opts1.Etcd.Hosts = []string{"http://host1:2379"}

	_ = impl.SetOptions(opts1)

	opts2 := validOptions()
	opts2.Etcd.Hosts = []string{"http://host2:2379"}
	opts2.TTL = 120
	opts2.Prefix = "/app"

	_ = impl.SetOptions(opts2)

	got := impl.GetOptions().(EtcdComponentOptions)
	if len(got.Etcd.Hosts) != 1 || got.Etcd.Hosts[0] != "http://host2:2379" {
		t.Fatalf("expected host2, got %v", got.Etcd.Hosts)
	}
	if got.TTL != 120 {
		t.Fatalf("expected ttl 120, got %d", got.TTL)
	}
	if got.Prefix != "/app" {
		t.Fatalf("expected prefix /app, got %s", got.Prefix)
	}
}
