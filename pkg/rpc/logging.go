package rpc

import "github.com/sora-soft/sora-go-framework.git/pkg/logger"

var FrameLogger *logger.Logger
var RpcLogger *logger.Logger

func SetFrameLogger(l *logger.Logger) { FrameLogger = l }
func SetRpcLogger(l *logger.Logger) { RpcLogger = l }
