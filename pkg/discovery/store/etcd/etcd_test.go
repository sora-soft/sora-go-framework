package etcd

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/sora-soft/sora-go-framework.git/pkg/component"
	"github.com/sora-soft/sora-go-framework.git/pkg/discovery"
	"github.com/sora-soft/sora-go-framework.git/pkg/runtime"
)

func etcdEndpoints() []string {
	if v := os.Getenv("ETCD_ENDPOINTS"); v != "" {
		return []string{v}
	}
	return []string{"localhost:2379"}
}

const testComponentName = "etcd-test"

func setupTestComponent(t *testing.T) {
	t.Helper()

	etcdComp := component.NewBaseEtcdComponent(testComponentName)
	err := etcdComp.LoadOptions(&component.EtcdOptions{
		Endpoints:   etcdEndpoints(),
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("LoadOptions: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := etcdComp.Start(ctx); err != nil {
		t.Fatalf("Start component: %v", err)
	}

	runtime.RT.RegisterComponent(testComponentName, etcdComp)

	t.Cleanup(func() {
		etcdComp.Stop()
	})
}

func newTestBackend(t *testing.T) *EtcdBackend {
	t.Helper()

	setupTestComponent(t)

	b := NewEtcdBackend(EtcdBackendOptions{
		ComponentName: testComponentName,
		Prefix:        "/sora-test/discovery",
		TTL:           10,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := b.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	t.Cleanup(func() {
		reg := b.Registry()
		disc := b.Discovery()
		ctx := context.Background()

		workers, _ := disc.ListWorkers(ctx)
		for _, w := range workers {
			reg.UnregisterWorker(ctx, w.ID)
		}
		services, _ := disc.ListServices(ctx)
		for _, s := range services {
			reg.UnregisterService(ctx, s.ID)
		}
		endpoints, _ := disc.ListEndpoints(ctx)
		for _, e := range endpoints {
			reg.UnregisterEndpoint(ctx, e.ID)
		}
		nodes, _ := disc.ListNodes(ctx)
		for _, n := range nodes {
			reg.UnregisterNode(ctx, n.ID)
		}

		b.Disconnect()
	})

	return b
}

func TestRegistryNodeRoundTrip(t *testing.T) {
	b := newTestBackend(t)
	reg := b.Registry()
	disc := b.Discovery()
	ctx := context.Background()

	node := discovery.NodeMeta{
		ID:        "node-1",
		Host:      "localhost",
		Pid:       1234,
		State:     3,
		StartTime: 1000,
	}
	if err := reg.RegisterNode(ctx, node); err != nil {
		t.Fatalf("RegisterNode: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	got, err := disc.GetNode(ctx, "node-1")
	if err != nil {
		t.Fatalf("GetNode: %v", err)
	}
	if got == nil || got.ID != "node-1" {
		t.Fatalf("expected node-1, got %v", got)
	}

	if err := reg.UnregisterNode(ctx, "node-1"); err != nil {
		t.Fatalf("UnregisterNode: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	got, err = disc.GetNode(ctx, "node-1")
	if err != nil {
		t.Fatalf("GetNode after unregister: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil after unregister, got %v", got)
	}
}

func TestRegistryServiceRoundTrip(t *testing.T) {
	b := newTestBackend(t)
	reg := b.Registry()
	disc := b.Discovery()
	ctx := context.Background()

	svc := discovery.ServiceMeta{
		Name:      "auth",
		ID:        "svc-1",
		State:     3,
		NodeID:    "node-1",
		StartTime: 1000,
		Labels:    map[string]string{"env": "test"},
	}
	if err := reg.RegisterService(ctx, svc); err != nil {
		t.Fatalf("RegisterService: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	got, err := disc.GetService(ctx, "svc-1")
	if err != nil {
		t.Fatalf("GetService: %v", err)
	}
	if got == nil || got.ID != "svc-1" || got.Name != "auth" {
		t.Fatalf("expected svc-1 auth, got %v", got)
	}

	byName, err := disc.ListServicesByName(ctx, "auth")
	if err != nil {
		t.Fatalf("ListServicesByName: %v", err)
	}
	if len(byName) != 1 || byName[0].ID != "svc-1" {
		t.Fatalf("expected 1 service named auth, got %v", byName)
	}

	if err := reg.UnregisterService(ctx, "svc-1"); err != nil {
		t.Fatalf("UnregisterService: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	got, _ = disc.GetService(ctx, "svc-1")
	if got != nil {
		t.Fatalf("expected nil after unregister")
	}
}

func TestRegistryWorkerRoundTrip(t *testing.T) {
	b := newTestBackend(t)
	reg := b.Registry()
	disc := b.Discovery()
	ctx := context.Background()

	w := discovery.WorkerMeta{
		Name:      "job-runner",
		ID:        "w-1",
		State:     3,
		NodeID:    "node-1",
		StartTime: 1000,
	}
	if err := reg.RegisterWorker(ctx, w); err != nil {
		t.Fatalf("RegisterWorker: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	got, err := disc.GetWorker(ctx, "w-1")
	if err != nil {
		t.Fatalf("GetWorker: %v", err)
	}
	if got == nil || got.ID != "w-1" {
		t.Fatalf("expected w-1, got %v", got)
	}

	byName, err := disc.ListWorkersByName(ctx, "job-runner")
	if err != nil {
		t.Fatalf("ListWorkersByName: %v", err)
	}
	if len(byName) != 1 {
		t.Fatalf("expected 1 worker, got %d", len(byName))
	}

	if err := reg.UnregisterWorker(ctx, "w-1"); err != nil {
		t.Fatalf("UnregisterWorker: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	got, _ = disc.GetWorker(ctx, "w-1")
	if got != nil {
		t.Fatalf("expected nil after unregister")
	}
}

func TestRegistryEndpointRoundTrip(t *testing.T) {
	b := newTestBackend(t)
	reg := b.Registry()
	disc := b.Discovery()
	ctx := context.Background()

	ep := discovery.EndpointMeta{
		ID:         "ep-1",
		Protocol:   "tcp",
		Endpoint:   "0.0.0.0:8080",
		State:      3,
		Weight:     100,
		TargetID:   "svc-1",
		TargetName: "auth",
	}
	if err := reg.RegisterEndpoint(ctx, ep); err != nil {
		t.Fatalf("RegisterEndpoint: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	got, err := disc.GetEndpoint(ctx, "ep-1")
	if err != nil {
		t.Fatalf("GetEndpoint: %v", err)
	}
	if got == nil || got.ID != "ep-1" {
		t.Fatalf("expected ep-1, got %v", got)
	}

	bySvc, err := disc.ListEndpointsByService(ctx, "svc-1")
	if err != nil {
		t.Fatalf("ListEndpointsByService: %v", err)
	}
	if len(bySvc) != 1 || bySvc[0].ID != "ep-1" {
		t.Fatalf("expected 1 endpoint for svc-1, got %v", bySvc)
	}

	if err := reg.UnregisterEndpoint(ctx, "ep-1"); err != nil {
		t.Fatalf("UnregisterEndpoint: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	got, _ = disc.GetEndpoint(ctx, "ep-1")
	if got != nil {
		t.Fatalf("expected nil after unregister")
	}
}

func TestWatchUpdateOnRegister(t *testing.T) {
	b := newTestBackend(t)
	reg := b.Registry()
	ctx := context.Background()

	ch := b.Discovery().WatchServices(ctx)

	select {
	case snapshot := <-ch:
		_ = snapshot
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for initial snapshot")
	}

	reg.RegisterService(ctx, discovery.ServiceMeta{ID: "s1", Name: "auth"})

	select {
	case snapshot := <-ch:
		found := false
		for _, s := range snapshot {
			if s.ID == "s1" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected snapshot with s1, got %v", snapshot)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for update after register")
	}
}

func TestWatchContextCancel(t *testing.T) {
	b := newTestBackend(t)
	ctx, cancel := context.WithCancel(context.Background())

	ch := b.Discovery().WatchServices(ctx)
	<-ch

	cancel()

	_, ok := <-ch
	if ok {
		t.Fatal("expected channel to be closed after context cancel")
	}
}

func TestElectionFirstCandidateWins(t *testing.T) {
	b := newTestBackend(t)
	ctx := context.Background()
	e := b.NewElection("test-election")

	if err := e.Campaign(ctx, "node-1"); err != nil {
		t.Fatalf("Campaign: %v", err)
	}

	leader, err := e.Leader(ctx)
	if err != nil {
		t.Fatalf("Leader: %v", err)
	}
	if leader != "node-1" {
		t.Fatalf("expected node-1, got %s", leader)
	}
}

func TestElectionResignTransfersLeadership(t *testing.T) {
	b := newTestBackend(t)
	ctx := context.Background()
	e := b.NewElection("test-election-2")

	e.Campaign(ctx, "node-1")

	doneCh := make(chan error, 1)
	go func() {
		doneCh <- e.Campaign(ctx, "node-2")
	}()

	time.Sleep(50 * time.Millisecond)

	if err := e.Resign(ctx); err != nil {
		t.Fatalf("Resign: %v", err)
	}

	select {
	case err := <-doneCh:
		if err != nil {
			t.Fatalf("second campaign after resign: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("second campaign should complete after resign")
	}

	leader, _ := e.Leader(ctx)
	if leader != "node-2" {
		t.Fatalf("expected node-2 as leader after resign, got %s", leader)
	}
}

func TestBackendGetInfo(t *testing.T) {
	b := newTestBackend(t)
	info := b.GetInfo()
	if info.Type != "etcd" {
		t.Fatalf("expected type etcd, got %s", info.Type)
	}
}

func TestNewElectionSameNameShares(t *testing.T) {
	b := newTestBackend(t)
	ctx := context.Background()

	e1 := b.NewElection("shared")
	e2 := b.NewElection("shared")

	e1.Campaign(ctx, "node-1")

	leader, _ := e2.Leader(ctx)
	if leader != "node-1" {
		t.Fatalf("expected e2 to see node-1 as leader, got %s", leader)
	}
}

func TestConnect_ComponentNotFound(t *testing.T) {
	b := NewEtcdBackend(EtcdBackendOptions{
		ComponentName: "nonexistent",
		Prefix:        "/test",
		TTL:           10,
	})

	err := b.Connect(context.Background())
	if err == nil {
		t.Fatal("expected error for missing component")
	}
}

func TestConnect_ComponentNotEtcd(t *testing.T) {
	runtime.RT.RegisterComponent("bad-component", component.NewEtcdComponent("bad"))
	// NewEtcdComponent returns Component interface (baseComponent), not *BaseEtcdComponent

	b := NewEtcdBackend(EtcdBackendOptions{
		ComponentName: "bad-component",
		Prefix:        "/test",
		TTL:           10,
	})

	err := b.Connect(context.Background())
	if err == nil {
		t.Fatal("expected error for wrong component type")
	}
}
