package discovery

import (
	"github.com/sora-soft/sora-go-framework.git/pkg/rpc"
)

type NodeVersions struct {
	Framework string `json:"framework"`
	App       string `json:"app"`
}

type NodeMeta struct {
	ID        string       `json:"id"`
	Alias     *string      `json:"alias,omitempty"`
	Host      string       `json:"host"`
	Pid       int          `json:"pid"`
	State     int          `json:"state"`
	StartTime int64        `json:"startTime"`
	Versions  NodeVersions `json:"versions"`
}

type ServiceMeta struct {
	Name      string            `json:"name"`
	ID        string            `json:"id"`
	Alias     *string           `json:"alias,omitempty"`
	State     int               `json:"state"`
	NodeID    string            `json:"nodeId"`
	StartTime int64             `json:"startTime"`
	Labels    map[string]string `json:"labels"`
}

type WorkerMeta struct {
	Name      string  `json:"name"`
	ID        string  `json:"id"`
	Alias     *string `json:"alias,omitempty"`
	State     int     `json:"state"`
	NodeID    string  `json:"nodeId"`
	StartTime int64   `json:"startTime"`
}

type EndpointMeta struct {
	ID         string            `json:"id"`
	Protocol   string            `json:"protocol"`
	Endpoint   string            `json:"endpoint"`
	State      int               `json:"state"`
	Labels     map[string]string `json:"labels"`
	Codecs     []string          `json:"codecs"`
	Weight     int               `json:"weight"`
	TargetID   string            `json:"targetId"`
	TargetName string            `json:"targetName"`
}

type BackendInfo struct {
	Type    string `json:"type"`
	Version string `json:"version"`
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
