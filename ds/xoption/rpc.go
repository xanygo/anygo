//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-06

package xoption

import (
	"fmt"
	"time"

	"github.com/xanygo/anygo/ds/xbus"
	"github.com/xanygo/anygo/ds/xmap"
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
	KeyWorkerCycle      = NewKey("WorkerCycle")

	keyExtraPrefix = "Extra:"

	// KeyExtra 只用于 ConsumeRPCConfig 方法,业务层可以通过此key内传递此类消息以更新 option
	KeyExtra = NewKey(keyExtraPrefix)
)

var (
	DefaultConnectRetry     = 1                // 默认网络连接次数
	DefaultConnectTimeout   = 10 * time.Second // 默认连接超时
	DefaultWriteTimeout     = 5 * time.Second  // 默认网络写超时
	DefaultReadTimeout      = 5 * time.Second  // 默认网络读超时
	DefaultHandshakeTimeout = 5 * time.Second  // 默认 rpc 协议层面握手超时
	DefaultRetry            = 1                // RPC 默认重试次数
	DefaultWorkerCycle      = 5 * time.Second  // 默认的 worker 运行周期
)

func SetConnectTimeout(opt Writer, timeout time.Duration) {
	opt.Set(KeyConnectTimeout, timeout)
}

func ConnectTimeout(opt Reader) time.Duration {
	return Duration(opt, KeyConnectTimeout, DefaultConnectTimeout)
}

func SetConnectRetry(opt Writer, retry int) {
	opt.Set(KeyConnectRetry, retry)
}

func ConnectRetry(opt Reader) int {
	return Int(opt, KeyConnectRetry, DefaultConnectRetry)
}

func SetWriteTimeout(opt Writer, timeout time.Duration) {
	opt.Set(KeyWriteTimeout, timeout)
}

func WriteTimeout(opt Reader) time.Duration {
	return Duration(opt, KeyWriteTimeout, DefaultWriteTimeout)
}

func SetReadTimeout(opt Writer, timeout time.Duration) {
	opt.Set(KeyReadTimeout, timeout)
}

func ReadTimeout(opt Reader) time.Duration {
	return Duration(opt, KeyReadTimeout, DefaultReadTimeout)
}

func SetHandshakeTimeout(opt Writer, timeout time.Duration) {
	opt.Set(KeyHandshakeTimeout, timeout)
}

func HandshakeTimeout(opt Reader) time.Duration {
	return Duration(opt, KeyHandshakeTimeout, DefaultHandshakeTimeout)
}

func SetRetry(opt Writer, retry int) {
	retry = max(0, retry)
	opt.Set(KeyRetry, retry)
}

func Retry(opt Reader) int {
	return Int(opt, KeyRetry, DefaultRetry)
}

func SetMaxResponseSize(opt Writer, maxSize int64) {
	opt.Set(KeyMaxResponseSize, maxSize)
}

const (
	mb = 1 << 20 // 1 MB
)

func MaxResponseSize(opt Reader) int64 {
	return Int64(opt, KeyMaxResponseSize, 64*mb)
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

// SetExtra 设置其他额外属性，需要注意，这类属性应该是可以枚举的，是有限的
func SetExtra(opt Writer, key string, value any) {
	ek := extraKeys.Get(key)
	opt.Set(ek, value)
}

func SetExtraByKV(opt Writer, kv KeyValue[string, any]) {
	ek := extraKeys.Get(kv.K)
	opt.Set(ek, kv.V)
}

func Extra(opt Reader, key string) any {
	val, _ := Get(opt, extraKeys.Get(key))
	return val
}

func SetWorkerCycle(opt Writer, c time.Duration) {
	opt.Set(KeyWorkerCycle, c)
}

func WorkerCycle(opt Reader) time.Duration {
	return Duration(opt, KeyWorkerCycle, DefaultWorkerCycle)
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
	case KeyProtocol, KeyProtocol.Name():
		return convertDoSet[string](d, msg.Payload, SetProtocol)
	case KeyMaxResponseSize, KeyMaxResponseSize.Name():
		return convertDoSet[int64](d, msg.Payload, SetMaxResponseSize)
	case KeyHandshakeTimeout, KeyHandshakeTimeout.Name():
		return convertDoSet[time.Duration](d, msg.Payload, SetHandshakeTimeout)
	case KeyExtra, KeyExtra.Name():
		return convertDoSet[KeyValue[string, any]](d, msg.Payload, SetExtraByKV)
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
