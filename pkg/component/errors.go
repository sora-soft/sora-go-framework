package component

import "errors"

var (
	ErrEtcdOptionsNotSet  = errors.New("etcd: options not set")
	ErrEtcdInvalidOptions = errors.New("etcd: invalid options type, expected *EtcdOptions")
)
