package node

import (
	"context"
	"os"

	"github.com/sora-soft/sora-go-framework.git/pkg/component"
	"github.com/sora-soft/sora-go-framework.git/pkg/rpc"
	"github.com/sora-soft/sora-go-framework.git/pkg/runtime"
	"github.com/sora-soft/sora-go-framework.git/pkg/runner/types"
)

type NodeVersions struct {
	Framework string `json:"framework,omitempty"`
	App       string `json:"app,omitempty"`
}

type NodeMetaData struct {
	Id        string                  `json:"id,omitempty"`
	Alias     *string                 `json:"alias,omitempty"`
	Host      string                  `json:"host,omitempty"`
	Pid       int                     `json:"pid,omitempty"`
	Endpoints []rpc.ListenerMetaInfo  `json:"endpoints,omitempty"`
	State     types.WorkerState       `json:"state,omitempty"`
	StartTime int64                   `json:"startTime,omitempty"`
	Versions  NodeVersions            `json:"versions,omitempty"`
}

type NodeRunData struct {
	Node       NodeMetaData                  `json:"node"`
	Components []component.ComponentMetadata `json:"components"`
	Services   []types.WorkerMetaData        `json:"services"`
	Workers    []types.WorkerMetaData        `json:"workers"`
}

type NodeOptions struct {
	types.WorkerOptions
	Version string `json:"version,omitempty"`
}

type NodeRunner struct {
	options   NodeOptions
	listeners []*rpc.Listener
	svc       types.Service
	host      string
	pid       int
}

func NewNodeRunner(opts NodeOptions, listeners []*rpc.Listener) *NodeRunner {
	return &NodeRunner{
		options:   opts,
		listeners: listeners,
		host:      mustHostname(),
		pid:       os.Getpid(),
	}
}

func mustHostname() string {
	name, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return name
}

func (n *NodeRunner) Startup(ctx context.Context) error {
	for _, l := range n.listeners {
		if err := n.svc.InstallListener(ctx, l); err != nil {
			return err
		}
	}

	return nil
}

func (n *NodeRunner) Shutdown() error {
	return nil
}

func (n *NodeRunner) SetService(svc types.Service) {
	n.svc = svc
}

func (n *NodeRunner) StateData() NodeMetaData {
	meta := NodeMetaData{
		Alias: n.options.Alias,
		Host:  n.host,
		Pid:   n.pid,
		Versions: NodeVersions{
			Framework: "0.0.0",
			App:       n.options.Version,
		},
	}

	if n.svc != nil {
		svcMeta := n.svc.GetMetadata()
		meta.Id = svcMeta.Id
		meta.State = svcMeta.State
		meta.StartTime = svcMeta.StartTime
		listenerInfos := make([]rpc.ListenerMetaInfo, 0, len(n.listeners))
		for _, l := range n.listeners {
			listenerInfos = append(listenerInfos, l.GetMetaInfo())
		}
		meta.Endpoints = listenerInfos
	}

	return meta
}

func (n *NodeRunner) RunData() NodeRunData {
	rt := runtime.RT

	components := rt.GetAllComponents()
	compMeta := make([]component.ComponentMetadata, 0, len(components))
	for _, c := range components {
		compMeta = append(compMeta, c.GetMetaInfo())
	}

	services := rt.GetAllServices()
	svcMeta := make([]types.WorkerMetaData, 0, len(services))
	for _, s := range services {
		svcMeta = append(svcMeta, s.GetMetadata())
	}

	workers := rt.GetAllWorkers()
	workerMeta := make([]types.WorkerMetaData, 0, len(workers))
	for _, w := range workers {
		workerMeta = append(workerMeta, w.GetMetadata())
	}

	return NodeRunData{
		Node:       n.StateData(),
		Components: compMeta,
		Services:   svcMeta,
		Workers:    workerMeta,
	}
}
