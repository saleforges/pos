package id

import (
	"sync/atomic"
	"time"
)

var counter atomic.Int64

func Generate() string {
	counter.Add(1)
	b := make([]byte, 8)
	now := time.Now().UnixNano()
	for i := 0; i < 8; i++ {
		b[i] = byte(now >> (i * 8))
	}
	return formatID(b)
}

func formatID(b []byte) string {
	const hex = "0123456789abcdef"
	buf := make([]byte, 16)
	for i, v := range b {
		buf[i*2] = hex[v>>4]
		buf[i*2+1] = hex[v&0x0f]
	}
	return string(buf)
}
