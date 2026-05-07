package main

import (
	"context"
	"fmt"
	"time"

	ramDisc "github.com/sora-soft/sora-go-framework.git/pkg/discovery/store/ram"
	jsoncodec "github.com/sora-soft/sora-go-framework.git/pkg/rpc/codec/json"
	"github.com/sora-soft/sora-go-framework.git/pkg/rpc/provider"
	"github.com/sora-soft/sora-go-framework.git/pkg/rpc/router"
	"github.com/sora-soft/sora-go-framework.git/pkg/rpc/transport/tcp"

	"github.com/sora-soft/sora-go-framework.git/pkg/discovery"
	"github.com/sora-soft/sora-go-framework.git/pkg/rpc"
	"github.com/sora-soft/sora-go-framework.git/pkg/utility"
)

type EchoRequest struct {
	Message string `json:"message"`
}

type EchoResponse struct {
	Message string `json:"message"`
	Time    string `json:"time"`
}

func main() {
	ctx := context.Background()

	jsonCodec := &jsoncodec.JSONBufferCodec{}
	rpc.RegisterCodec(jsonCodec)

	backend := ramDisc.NewRamBackend()
	disco := backend.Discovery()
	registry := backend.Registry()

	tl, err := tcp.NewTCPListener(tcp.TCPListenerOptions{
		Host:      "127.0.0.1",
		PortRange: []int{19000, 19999},
	})
	if err != nil {
		fmt.Println("create tcp listener failed:", err)
		return
	}

	r := router.NewRouter()

	router.Method(r, "echo", func(conn *rpc.Connection, req EchoRequest) (EchoResponse, error) {
		fmt.Printf("[Listener] received: %s\n", req.Message)
		return EchoResponse{
			Message: "echo: " + req.Message,
			Time:    time.Now().Format(time.RFC3339),
		}, nil
	})

	router.Notify(r, "ping", func(conn *rpc.Connection, msg EchoRequest) error {
		fmt.Printf("[Listener] notify received: %s\n", msg.Message)
		return nil
	})

	listener := rpc.NewListener(tl, utility.Labels{}, []rpc.Codec{jsonCodec}, rpc.ListenerCallbacks{
		OnRequest: r.OnRequestCB(),
		OnNotify:  r.OnNotifyCB(),
		OnSessionOpen: func(conn *rpc.Connection, sessionId string) {
			fmt.Println("[Listener] session opened:", sessionId)
		},
		OnSessionClose: func(conn *rpc.Connection, sessionId string) {
			fmt.Println("[Listener] session closed:", sessionId)
		},
	})

	if err := listener.Start(ctx); err != nil {
		fmt.Println("listener start failed:", err)
		return
	}

	meta := listener.GetMetaInfo()
	fmt.Printf("[Listener] listening on %s\n", meta.Endpoint)

	endpoint := discovery.EndpointMeta{
		ID:         meta.Id,
		Protocol:   meta.Protocol,
		Endpoint:   meta.Endpoint,
		Codecs:     meta.Codecs,
		Weight:     100,
		TargetID:   "svc-echo-001",
		TargetName: "echo-service",
	}
	if err := registry.RegisterEndpoint(ctx, endpoint); err != nil {
		fmt.Println("register endpoint failed:", err)
		return
	}
	fmt.Println("[Discovery] endpoint registered:", endpoint.ID)

	p := provider.NewProvider("echo-service", disco, provider.ProviderOptions{})
	if err := p.Start(ctx); err != nil {
		fmt.Println("provider start failed:", err)
		return
	}
	fmt.Println("[Provider] started, waiting for sender to connect...")

	time.Sleep(2 * time.Second)

	for i := 0; i < 3; i++ {
		fmt.Printf("\n--- Call %d ---\n", i+1)

		req := EchoRequest{Message: fmt.Sprintf("hello world %d", i+1)}
		resp, err := provider.CallRpc[EchoResponse](p, ctx, "echo", req)
		if err != nil {
			fmt.Println("CallRpc failed:", err)
			continue
		}

		fmt.Printf("[Provider] response: message=%s time=%s\n", resp.Message, resp.Time)

		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("\n--- Test: unregistered method ---")
	{
		_, err := provider.CallRpc[any](p, ctx, "nonexistent", EchoRequest{Message: "test"})
		if err != nil {
			fmt.Printf("[Provider] unregistered method error: %v (expected ERR_METHOD_NOT_FOUND)\n", err)
		}
	}

	_ = p.Stop()
	listener.Stop()

	fmt.Println("\ndone")
}
