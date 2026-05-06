package main

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/sora-soft/sora-go-framework.git/pkg/component"
	etcdDisc "github.com/sora-soft/sora-go-framework.git/pkg/discovery/store/etcd"
	"github.com/sora-soft/sora-go-framework.git/pkg/logger"
	nodePkg "github.com/sora-soft/sora-go-framework.git/pkg/node"
	rtPkg "github.com/sora-soft/sora-go-framework.git/pkg/runtime"
	"github.com/sora-soft/sora-go-framework.git/pkg/runner"
	"github.com/sora-soft/sora-go-framework.git/pkg/runner/types"
)

func main() {
	rt := rtPkg.RT
	ctx := context.Background()

	consoleOutput := logger.NewConsoleOutput(
		logger.LogLevelDebug,
		logger.LogLevelInfo,
		logger.LogLevelSuccess,
		logger.LogLevelWarn,
		logger.LogLevelError,
		logger.LogLevelFatal,
	)
	rt.FrameLogger.AddOutput(consoleOutput)
	rt.RpcLogger.AddOutput(consoleOutput)

	log := logger.NewLogger("sora-test").AddOutput(consoleOutput)

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Panic: %v\n", err)
			fmt.Printf("Stack Trace:\n%s\n", debug.Stack())
		}
	}()

	etcdComp := component.NewBaseEtcdComponent("etcd")
	etcdComp.LoadOptions(&component.EtcdOptions{
		Endpoints:   []string{"http://chitanda.xyyaya.com:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err := etcdComp.Start(ctx); err != nil {
		log.Error("main", err, map[string]any{"event": "etcd_start_failed"})
		return
	}
	rt.RegisterComponent("etcd", etcdComp)
	log.Info("main", map[string]string{"event": "etcd_component_registered"})

	backend := etcdDisc.NewEtcdBackend(etcdDisc.EtcdBackendOptions{
		ComponentName: "etcd",
		Prefix:        "sora-test",
		TTL:           20,
	})

	nodeRunner := nodePkg.NewNodeRunner(nodePkg.NodeOptions{Version: "1.0.0"}, nil)
	nodeSvc := runner.NewService("node", nodeRunner, types.ServiceOptions{})
	nodeRunner.SetService(nodeSvc)

	if err := rt.Startup(ctx, nodeSvc, backend); err != nil {
		log.Error("main", err, map[string]any{"event": "startup_failed"})
		return
	}
	log.Info("main", map[string]string{"event": "runtime_started"})

	time.Sleep(500 * time.Millisecond)

	disc := rt.GetDiscovery()
	services, err := disc.ListServices(ctx)
	if err != nil {
		log.Error("main", err, map[string]any{"event": "list_services_failed"})
	} else {
		fmt.Println("=== Service List ===")
		for _, svc := range services {
			fmt.Printf("  Name: %s | ID: %s | NodeID: %s | State: %d\n", svc.Name, svc.ID, svc.NodeID, svc.State)
		}
		if len(services) == 0 {
			fmt.Println("  (empty)")
		}
		fmt.Println("====================")
	}

	if err := rt.Shutdown(); err != nil {
		log.Error("main", err, map[string]any{"event": "shutdown_failed"})
	}
	etcdComp.Stop()
	log.Info("main", "Shutdown complete")
	log.Close()
}
