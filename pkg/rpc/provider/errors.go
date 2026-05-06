package provider

import "github.com/sora-soft/sora-go-framework.git/pkg/utility/errorx"

var (
	ErrNoAvailableEndpoint = errorx.New("ERR_NO_AVAILABLE_ENDPOINT", errorx.LevelExpected, "ProviderError", "no available endpoint", nil)
	ErrServiceNotFound     = errorx.New("ERR_SERVICE_NOT_FOUND", errorx.LevelExpected, "ProviderError", "service not found", nil)
	ErrCallTimeout         = errorx.New("ERR_CALL_TIMEOUT", errorx.LevelExpected, "ProviderError", "call timeout", nil)
	ErrConnectionLost      = errorx.New("ERR_CONNECTION_LOST", errorx.LevelUnexpected, "ProviderError", "connection lost", nil)
	ErrSenderStopped       = errorx.New("ERR_SENDER_STOPPED", errorx.LevelExpected, "ProviderError", "sender stopped", nil)
	ErrProviderStopped     = errorx.New("ERR_PROVIDER_STOPPED", errorx.LevelExpected, "ProviderError", "provider stopped", nil)
	ErrNoAvailableCodec    = errorx.New("ERR_NO_AVAILABLE_CODEC", errorx.LevelExpected, "ProviderError", "no available codec", nil)
)
