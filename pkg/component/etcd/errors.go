package etcd

import (
	"github.com/sora-soft/sora-go-framework.git/pkg/utility/errorx"
)

const (
	ErrCodeOptionsNotSet   = "ERR_ETCD_OPTIONS_NOT_SET"
	ErrCodeInvalidOptions  = "ERR_ETCD_INVALID_OPTIONS"
	ErrCodeHostsEmpty      = "ERR_ETCD_HOSTS_EMPTY"
	ErrCodeDialTimeoutZero = "ERR_ETCD_DIAL_TIMEOUT_ZERO"
	ErrCodeTTLInvalid      = "ERR_ETCD_TTL_INVALID"
	ErrCodePrefixEmpty     = "ERR_ETCD_PREFIX_EMPTY"
	ErrCodeNotConnected    = "ERR_ETCD_NOT_CONNECTED"
	ErrCodeLeaseNotFound   = "ERR_ETCD_LEASE_NOT_FOUND"
)

func newOptionsNotSetError() *errorx.Error {
	return errorx.New(ErrCodeOptionsNotSet, errorx.LevelUnexpected, "EtcdComponentError", "etcd component options not set", nil)
}

func newInvalidOptionsError() *errorx.Error {
	return errorx.New(ErrCodeInvalidOptions, errorx.LevelUnexpected, "EtcdComponentError", "invalid options type, expected *EtcdComponentOptions", nil)
}

func newHostsEmptyError() *errorx.Error {
	return errorx.New(ErrCodeHostsEmpty, errorx.LevelExpected, "EtcdComponentError", "etcd hosts must not be empty", nil)
}

func newDialTimeoutZeroError() *errorx.Error {
	return errorx.New(ErrCodeDialTimeoutZero, errorx.LevelExpected, "EtcdComponentError", "etcd dialTimeout must be greater than zero", nil)
}

func newTTLInvalidError(ttl int64) *errorx.Error {
	return errorx.New(ErrCodeTTLInvalid, errorx.LevelExpected, "EtcdComponentError", "etcd ttl must be greater than zero", map[string]any{"ttl": ttl})
}

func newPrefixEmptyError() *errorx.Error {
	return errorx.New(ErrCodePrefixEmpty, errorx.LevelExpected, "EtcdComponentError", "etcd prefix must not be empty", nil)
}

func newNotConnectedError() *errorx.Error {
	return errorx.New(ErrCodeNotConnected, errorx.LevelUnexpected, "EtcdComponentError", "etcd component not connected", nil)
}

func newLeaseNotFoundError() *errorx.Error {
	return errorx.New(ErrCodeLeaseNotFound, errorx.LevelUnexpected, "EtcdComponentError", "etcd lease not found", nil)
}
