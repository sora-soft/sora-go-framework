package etcd

import (
	"github.com/sora-soft/sora-go-framework/pkg/utility/errorx"
)

const (
	ErrEtcdNotConnected = "ERR_ETCD_NOT_CONNECTED"
	ErrNodeNotFound     = "ERR_NODE_NOT_FOUND"
	ErrServiceNotFound  = "ERR_SERVICE_NOT_FOUND"
	ErrWorkerNotFound   = "ERR_WORKER_NOT_FOUND"
	ErrEndpointNotFound = "ERR_ENDPOINT_NOT_FOUND"
)

func newNotConnectedError() *errorx.Error {
	return errorx.New(ErrEtcdNotConnected, errorx.LevelUnexpected, "ETCDDiscoveryError", "etcd is not connected", nil)
}

func newNodeNotFoundError(id string) *errorx.Error {
	return errorx.New(ErrNodeNotFound, errorx.LevelExpected, "ETCDDiscoveryError", "node not found", map[string]any{"id": id})
}

func newServiceNotFoundError(id string) *errorx.Error {
	return errorx.New(ErrServiceNotFound, errorx.LevelExpected, "ETCDDiscoveryError", "service not found", map[string]any{"id": id})
}

func newWorkerNotFoundError(id string) *errorx.Error {
	return errorx.New(ErrWorkerNotFound, errorx.LevelExpected, "ETCDDiscoveryError", "worker not found", map[string]any{"id": id})
}

func newEndpointNotFoundError(id string) *errorx.Error {
	return errorx.New(ErrEndpointNotFound, errorx.LevelExpected, "ETCDDiscoveryError", "endpoint not found", map[string]any{"id": id})
}

func newComponentNotFoundError(name string) *errorx.Error {
	return errorx.New("ERR_COMPONENT_NOT_FOUND", errorx.LevelUnexpected, "ETCDDiscoveryError", "etcd component not found in runtime", map[string]any{"componentName": name})
}

func newComponentTypeError(name string) *errorx.Error {
	return errorx.New("ERR_COMPONENT_TYPE", errorx.LevelUnexpected, "ETCDDiscoveryError", "component is not an etcd component", map[string]any{"componentName": name})
}
