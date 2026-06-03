package ram

import (
	"context"
	"testing"
	"time"

	"github.com/sora-soft/sora-go-framework/pkg/discovery"
)

func newTestBackend() *RamBackend {
	return NewRamBackend()
}

func TestRegistryRoundTrip(t *testing.T) {
	b := newTestBackend()
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
	got, err = disc.GetNode(ctx, "node-1")
	if err != nil {
		t.Fatalf("GetNode after unregister: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil after unregister, got %v", got)
	}
}

func TestRegistryServiceRoundTrip(t *testing.T) {
	b := newTestBackend()
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
	got, _ = disc.GetService(ctx, "svc-1")
	if got != nil {
		t.Fatalf("expected nil after unregister")
	}
}

func TestRegistryEndpointRoundTrip(t *testing.T) {
	b := newTestBackend()
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
	got, _ = disc.GetEndpoint(ctx, "ep-1")
	if got != nil {
		t.Fatalf("expected nil after unregister")
	}
}

func TestRegistryWorkerRoundTrip(t *testing.T) {
	b := newTestBackend()
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
	got, _ = disc.GetWorker(ctx, "w-1")
	if got != nil {
		t.Fatalf("expected nil after unregister")
	}
}

func TestListEmpty(t *testing.T) {
	b := newTestBackend()
	disc := b.Discovery()
	ctx := context.Background()

	nodes, _ := disc.ListNodes(ctx)
	if nodes == nil || len(nodes) != 0 {
		t.Fatalf("expected empty non-nil slice, got %v", nodes)
	}
	services, _ := disc.ListServices(ctx)
	if services == nil || len(services) != 0 {
		t.Fatalf("expected empty non-nil slice, got %v", services)
	}
	workers, _ := disc.ListWorkers(ctx)
	if workers == nil || len(workers) != 0 {
		t.Fatalf("expected empty non-nil slice, got %v", workers)
	}
	endpoints, _ := disc.ListEndpoints(ctx)
	if endpoints == nil || len(endpoints) != 0 {
		t.Fatalf("expected empty non-nil slice, got %v", endpoints)
	}
}

func TestWatchInitialSnapshot(t *testing.T) {
	b := newTestBackend()
	reg := b.Registry()
	ctx := context.Background()

	reg.RegisterService(ctx, discovery.ServiceMeta{ID: "s1", Name: "a"})
	reg.RegisterService(ctx, discovery.ServiceMeta{ID: "s2", Name: "b"})

	ch := b.Discovery().WatchServices(ctx)

	select {
	case snapshot := <-ch:
		if len(snapshot) != 2 {
			t.Fatalf("expected 2 services in initial snapshot, got %d", len(snapshot))
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for initial snapshot")
	}
}

func TestWatchUpdateOnRegister(t *testing.T) {
	b := newTestBackend()
	reg := b.Registry()
	ctx := context.Background()

	ch := b.Discovery().WatchServices(ctx)

	select {
	case snapshot := <-ch:
		if len(snapshot) != 0 {
			t.Fatalf("expected empty initial snapshot, got %d", len(snapshot))
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for initial snapshot")
	}

	reg.RegisterService(ctx, discovery.ServiceMeta{ID: "s1", Name: "auth"})

	select {
	case snapshot := <-ch:
		if len(snapshot) != 1 || snapshot[0].ID != "s1" {
			t.Fatalf("expected snapshot with s1, got %v", snapshot)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for update after register")
	}
}

func TestWatchUpdateOnUnregister(t *testing.T) {
	b := newTestBackend()
	reg := b.Registry()
	ctx := context.Background()

	reg.RegisterService(ctx, discovery.ServiceMeta{ID: "s1", Name: "auth"})

	ch := b.Discovery().WatchServices(ctx)
	<-ch // drain initial

	reg.UnregisterService(ctx, "s1")

	select {
	case snapshot := <-ch:
		if len(snapshot) != 0 {
			t.Fatalf("expected empty snapshot after unregister, got %d", len(snapshot))
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for update after unregister")
	}
}

func TestWatchContextCancel(t *testing.T) {
	b := newTestBackend()
	ctx, cancel := context.WithCancel(context.Background())

	ch := b.Discovery().WatchServices(ctx)
	<-ch // drain initial

	cancel()

	_, ok := <-ch
	if ok {
		t.Fatal("expected channel to be closed after context cancel")
	}
}

func TestElectionFirstCandidateWins(t *testing.T) {
	b := newTestBackend()
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

func TestElectionBlocksSecondCandidate(t *testing.T) {
	b := newTestBackend()
	ctx := context.Background()
	e := b.NewElection("test-election")

	e.Campaign(ctx, "node-1")

	doneCh := make(chan error, 1)
	go func() {
		doneCh <- e.Campaign(ctx, "node-2")
	}()

	select {
	case <-doneCh:
		t.Fatal("second campaign should block")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestElectionResignTransfersLeadership(t *testing.T) {
	b := newTestBackend()
	ctx := context.Background()
	e := b.NewElection("test-election")

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
	case <-time.After(time.Second):
		t.Fatal("second campaign should complete after resign")
	}

	leader, _ := e.Leader(ctx)
	if leader != "node-2" {
		t.Fatalf("expected node-2 as leader after resign, got %s", leader)
	}
}

func TestElectionNoLeader(t *testing.T) {
	b := newTestBackend()
	ctx := context.Background()
	e := b.NewElection("empty-election")

	leader, err := e.Leader(ctx)
	if err != nil {
		t.Fatalf("Leader: %v", err)
	}
	if leader != "" {
		t.Fatalf("expected empty string, got %s", leader)
	}
}

func TestBackendGetInfo(t *testing.T) {
	b := newTestBackend()
	info := b.GetInfo()
	if info.Type != "ram" {
		t.Fatalf("expected type ram, got %s", info.Type)
	}
	if info.Version != "0.0.0" {
		t.Fatalf("expected version 0.0.0, got %s", info.Version)
	}
}

func TestNewElectionSameNameShares(t *testing.T) {
	b := newTestBackend()
	ctx := context.Background()

	e1 := b.NewElection("shared")
	e2 := b.NewElection("shared")

	e1.Campaign(ctx, "node-1")

	leader, _ := e2.Leader(ctx)
	if leader != "node-1" {
		t.Fatalf("expected e2 to see node-1 as leader, got %s", leader)
	}
}
