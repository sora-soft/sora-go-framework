package rpc

import (
	"sync"
)

var (
	codecMu sync.RWMutex
	codecs  = make(map[string]Codec)
)

func RegisterCodec(c Codec) {
	codecMu.Lock()
	defer codecMu.Unlock()
	codecs[c.GetCode()] = c
}

func GetCodec(code string) (Codec, bool) {
	codecMu.RLock()
	defer codecMu.RUnlock()
	c, ok := codecs[code]
	return c, ok
}
