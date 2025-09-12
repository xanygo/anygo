//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-06

package xoption

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/xanygo/anygo/xbus"
	"github.com/xanygo/anygo/xvalidator"
)

var (
	KeyConnectTimeout  = NewKey("ConnectTimeout")
	KeyConnectRetry    = NewKey("ConnectRetry")
	KeyWriteTimeout    = NewKey("WriteTimeout")
	KeyReadTimeout     = NewKey("ReadTimeout")
	KeyRetry           = NewKey("Retry")
	KeyBalancer        = NewKey("Balancer") // 负载均衡策略名称
	KeyMaxResponseSize = NewKey("MaxResponseSize")

	KeyTLSConfig = NewKey("tls.Config")
	KeyProxy     = NewKey("Proxy") // proxy 类型，支持的值： HTTP
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
	case KeyProxy, KeyProxy.Name():
		return convertDoSet[*ProxyConfig](d, msg.Payload, SetProxy)
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

func SetTLSConfig(opt Writer, c *tls.Config) {
	opt.Set(KeyTLSConfig, c)
}

func TLSConfig(opt Reader) *tls.Config {
	return GetAsDefault[*tls.Config](opt, KeyTLSConfig, nil)
}

type ProxyConfig struct {
	Type     string `json:"Type" yaml:"Type"`         // 代理类型，必填，可选值： HTTP、SOCKS5（未支持）
	AuthType string `json:"AuthType" yaml:"AuthType"` // 认证类型，可选，可选值为： Basic(默认)
	Username string `json:"Username" yaml:"Username"` // 认证账号，可选，有值时才会发送认证信息
	Password string `json:"Password" yaml:"Password"` // 认证密码，可选
}

var _ xvalidator.AutoChecker = (*ProxyConfig)(nil)

func (pc *ProxyConfig) AutoCheck() error {
	switch pc.Type {
	case "HTTP", "SOCKS5":
	default:
		return fmt.Errorf("invalid proxy type: %q", pc.Type)
	}
	return nil
}

func SetProxy(opt Writer, proxy *ProxyConfig) {
	opt.Set(KeyProxy, proxy)
}

func Proxy(opt Reader) *ProxyConfig {
	return GetAsDefault[*ProxyConfig](opt, KeyProxy, nil)
}
