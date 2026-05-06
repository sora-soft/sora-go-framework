package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	ramDisc "github.com/sora-soft/sora-go-framework.git/pkg/discovery/store/ram"
	jsoncodec "github.com/sora-soft/sora-go-framework.git/pkg/rpc/codec/json"
	"github.com/sora-soft/sora-go-framework.git/pkg/rpc/packet"
	"github.com/sora-soft/sora-go-framework.git/pkg/rpc/provider"
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
		fmt.Println("create listener failed:", err)
		return
	}

	listener := rpc.NewListener(tl, utility.Labels{}, []rpc.Codec{jsonCodec}, rpc.ListenerCallbacks{
		OnRequest: func(conn *rpc.Connection, pkt packet.Packet) {
			req, err := packet.Decode[EchoRequest](pkt)
			if err != nil {
				fmt.Println("decode failed:", err)
				return
			}

			fmt.Printf("[Listener] received: %s\n", req.Message)

			resp := EchoResponse{
				Message: "echo: " + req.Message,
				Time:    time.Now().Format(time.RFC3339),
			}

			respPayload, _ := json.Marshal(resp)
			respPkt := packet.NewDecodedPacket(packet.PacketOpcodeResponse, "", "", pkt.Headers, respPayload, jsonCodec)
			if err := conn.SendResponse(ctx, respPkt); err != nil {
				fmt.Println("send response failed:", err)
			}
		},
		OnNotify: func(conn *rpc.Connection, pkt packet.Packet) {
		},
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
	p.Start(ctx)
	fmt.Println("[Provider] started, waiting for sender to connect...")

	time.Sleep(2 * time.Second)

	for i := 0; i < 3; i++ {
		fmt.Printf("\n--- Call %d ---\n", i+1)

		req := EchoRequest{Message: fmt.Sprintf("hello world %d", i+1)}
		respPkt, err := p.CallRpc(ctx, "echo", req)
		if err != nil {
			fmt.Println("CallRpc failed:", err)
			continue
		}

		resp, err := packet.Decode[EchoResponse](respPkt)
		if err != nil {
			fmt.Println("decode response failed:", err)
			continue
		}

		fmt.Printf("[Provider] response: message=%s time=%s\n", resp.Message, resp.Time)

		time.Sleep(500 * time.Millisecond)
	}

	p.Stop()
	listener.Stop()

	fmt.Println("\ndone")
}
