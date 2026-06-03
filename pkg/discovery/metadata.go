package discovery

import (
	"github.com/sora-soft/sora-go-framework/pkg/rpc"
)

type NodeVersions struct {
	Framework string `json:"framework" yaml:"framework"`
	App       string `json:"app" yaml:"app"`
}

type NodeMeta struct {
	ID        string       `json:"id" yaml:"id"`
	Alias     *string      `json:"alias,omitempty" yaml:"alias,omitempty"`
	Host      string       `json:"host" yaml:"host"`
	Pid       int          `json:"pid" yaml:"pid"`
	State     int          `json:"state" yaml:"state"`
	StartTime int64        `json:"startTime" yaml:"startTime"`
	Versions  NodeVersions `json:"versions" yaml:"versions"`
}

type ServiceMeta struct {
	Name      string            `json:"name" yaml:"name"`
	ID        string            `json:"id" yaml:"id"`
	Alias     *string           `json:"alias,omitempty" yaml:"alias,omitempty"`
	State     int               `json:"state" yaml:"state"`
	NodeID    string            `json:"nodeId" yaml:"nodeId"`
	StartTime int64             `json:"startTime" yaml:"startTime"`
	Labels    map[string]string `json:"labels" yaml:"labels"`
}

type WorkerMeta struct {
	Name      string  `json:"name" yaml:"name"`
	ID        string  `json:"id" yaml:"id"`
	Alias     *string `json:"alias,omitempty" yaml:"alias,omitempty"`
	State     int     `json:"state" yaml:"state"`
	NodeID    string  `json:"nodeId" yaml:"nodeId"`
	StartTime int64   `json:"startTime" yaml:"startTime"`
}

type EndpointMeta struct {
	ID         string            `json:"id" yaml:"id"`
	Protocol   string            `json:"protocol" yaml:"protocol"`
	Endpoint   string            `json:"endpoint" yaml:"endpoint"`
	State      int               `json:"state" yaml:"state"`
	Labels     map[string]string `json:"labels" yaml:"labels"`
	Codecs     []string          `json:"codecs" yaml:"codecs"`
	Weight     int               `json:"weight" yaml:"weight"`
	TargetID   string            `json:"targetId" yaml:"targetId"`
	TargetName string            `json:"targetName" yaml:"targetName"`
}

type BackendInfo struct {
	Type    string `json:"type" yaml:"type"`
	Version string `json:"version" yaml:"version"`
}

func NewEndpointMetaFromListener(info rpc.ListenerMetaInfo) EndpointMeta {
	return EndpointMeta{
		ID:       info.Id,
		Protocol: info.Protocol,
		Endpoint: info.Endpoint,
		State:    int(info.State),
		Labels:   info.Labels,
		Codecs:   info.Codecs,
	}
}
