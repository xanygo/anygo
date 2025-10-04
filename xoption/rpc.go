//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-06

package xoption

import (
	"fmt"
	"time"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/xbus"
)

var (
	KeyConnectTimeout   = NewKey("ConnectTimeout")
	KeyConnectRetry     = NewKey("ConnectRetry")
	KeyWriteTimeout     = NewKey("WriteTimeout")
	KeyReadTimeout      = NewKey("ReadTimeout")
	KeyHandshakeTimeout = NewKey("HandshakeTimeout")
	KeyRetry            = NewKey("Retry")
	KeyBalancer         = NewKey("Balancer") // 负载均衡策略名称
	KeyMaxResponseSize  = NewKey("MaxResponseSize")
	KeyUseProxy         = NewKey("UseProxy")
	KeyProtocol         = NewKey("Protocol")

	keyExtraPrefix = "Extra:"
)

func SetConnectTimeout(opt Writer, timeout time.Duration) {
	if timeout > 0 {
		opt.Set(KeyConnectTimeout, timeout)
	}
}

func ConnectTimeout(opt Reader) time.Duration {
	return Duration(opt, KeyConnectTimeout, 3*time.Second)
}

func SetConnectRetry(opt Writer, retry int) {
	opt.Set(KeyConnectRetry, retry)
}

func ConnectRetry(opt Reader) int {
	return Int(opt, KeyConnectRetry, 0)
}

func SetWriteTimeout(opt Writer, timeout time.Duration) {
	if timeout > 0 {
		opt.Set(KeyWriteTimeout, timeout)
	}
}

func WriteTimeout(opt Reader) time.Duration {
	return Duration(opt, KeyWriteTimeout, 3*time.Second)
}

func SetReadTimeout(opt Writer, timeout time.Duration) {
	if timeout > 0 {
		opt.Set(KeyReadTimeout, timeout)
	}
}

func ReadTimeout(opt Reader) time.Duration {
	return Duration(opt, KeyReadTimeout, 3*time.Second)
}

func SetHandshakeTimeout(opt Writer, timeout time.Duration) {
	if timeout > 0 {
		opt.Set(KeyHandshakeTimeout, timeout)
	}
}

func HandshakeTimeout(opt Reader) time.Duration {
	return Duration(opt, KeyHandshakeTimeout, 3*time.Second)
}

func SetRetry(opt Writer, retry int) {
	retry = max(0, retry)
	opt.Set(KeyRetry, retry)
}

func Retry(opt Reader) int {
	return Int(opt, KeyRetry, 0)
}

func SetMaxResponseSize(opt Writer, maxSize int64) {
	if maxSize > 0 {
		opt.Set(KeyMaxResponseSize, maxSize)
	}
}

const (
	mb = 1 << 20 // 1 MB
)

func MaxResponseSize(opt Reader) int64 {
	return Int64(opt, KeyMaxResponseSize, 10*mb)
}

func WriteReadTimeout(opt Reader) time.Duration {
	return WriteTimeout(opt) + ReadTimeout(opt)
}

func TotalTimeout(opt Reader) time.Duration {
	return ConnectTimeout(opt) + WriteReadTimeout(opt)
}

func SetBalancer(opt Writer, name string) {
	opt.Set(KeyBalancer, name)
}

func Balancer(opt Reader) string {
	return String(opt, KeyBalancer, "RoundRobin")
}

func SetUseProxy(opt Writer, name string) {
	opt.Set(KeyUseProxy, name)
}

func UseProxy(opt Reader) string {
	return String(opt, KeyUseProxy, "")
}

func SetProtocol(opt Writer, name string) {
	opt.Set(KeyProtocol, name)
}

func Protocol(opt Reader) string {
	return String(opt, KeyProtocol, "")
}

var extraKeys = &xmap.Cached[string, Key]{
	New: func(k string) Key {
		return NewKey(keyExtraPrefix + k)
	},
}

func SetExtra(opt Writer, key string, value any) {
	ek := extraKeys.Get(key)
	opt.Set(ek, value)
}

func Extra(opt Reader, key string) any {
	val, _ := Get(opt, extraKeys.Get(key))
	return val
}

func ConsumeRPCConfig(d Writer, msg xbus.Message) error {
	if msg.Topic != Topic || msg.Key == nil {
		return nil
	}

	switch msg.Key {
	case KeyConnectTimeout, KeyConnectTimeout.Name():
		return convertDoSet[time.Duration](d, msg.Payload, SetConnectTimeout)
	case KeyConnectRetry, KeyConnectRetry.Name():
		return convertDoSet[int](d, msg.Payload, SetConnectRetry)
	case KeyWriteTimeout, KeyWriteTimeout.Name():
		return convertDoSet[time.Duration](d, msg.Payload, SetWriteTimeout)
	case KeyReadTimeout, KeyReadTimeout.Name():
		return convertDoSet[time.Duration](d, msg.Payload, SetReadTimeout)
	case KeyRetry, KeyRetry.Name():
		return convertDoSet[int](d, msg.Payload, SetRetry)
	case KeyBalancer, KeyBalancer.Name():
		return convertDoSet[string](d, msg.Payload, SetBalancer)
	}

	return nil
}

func convertDoSet[T any](opt Writer, value any, fn func(opt Writer, val T)) error {
	cv, ok := value.(T)
	if !ok {
		return fmt.Errorf("invalid value type: %T", value)
	}
	fn(opt, cv)
	return nil
}
