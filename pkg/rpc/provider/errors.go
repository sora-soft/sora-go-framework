package provider

import "github.com/sora-soft/sora-go-framework.git/pkg/utility/errorx"

var (
	ErrNoAvailableEndpoint = errorx.New("ERR_NO_AVAILABLE_ENDPOINT", errorx.LevelExpected, "RpcError", "no available endpoint", nil)
	ErrServiceNotFound     = errorx.New("ERR_SERVICE_NOT_FOUND", errorx.LevelExpected, "RpcError", "service not found", nil)
	ErrCallTimeout         = errorx.New("ERR_CALL_TIMEOUT", errorx.LevelExpected, "RpcError", "call timeout", nil)
	ErrConnectionLost      = errorx.New("ERR_CONNECTION_LOST", errorx.LevelUnexpected, "RpcError", "connection lost", nil)
	ErrSenderStopped       = errorx.New("ERR_SENDER_STOPPED", errorx.LevelExpected, "RpcError", "sender stopped", nil)
	ErrProviderStopped     = errorx.New("ERR_PROVIDER_STOPPED", errorx.LevelExpected, "RpcError", "provider stopped", nil)
	ErrNoAvailableCodec    = errorx.New("ERR_NO_AVAILABLE_CODEC", errorx.LevelExpected, "RpcError", "no available codec", nil)
)
