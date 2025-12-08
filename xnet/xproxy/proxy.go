//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-15

package xproxy

import (
	"context"
	"fmt"

	"github.com/xanygo/anygo/ds/xoption"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xvalidator"
)

var OptKeyProxy = xoption.NewKey("Proxy") // proxy 类型，支持的值： HTTP

type Config struct {
	Protocol string `json:"Protocol" yaml:"Protocol"` // 代理类型，必填，可选值： HTTP、HTTPS、SOCKS5（未支持）
	AuthType string `json:"AuthType" yaml:"AuthType"` // 认证类型，可选，可选值为： Basic(默认)
	Username string `json:"Username" yaml:"Username"` // 认证账号，可选，有值时才会发送认证信息
	Password string `json:"Password" yaml:"Password"` // 认证密码，可选
	Host     string // 在 ParserProxyURL 解析的时候才会用到
	Port     string // 在 ParserProxyURL 解析的时候才会用到
	TLS      *xoption.TLSConfig
}

var _ xvalidator.AutoChecker = (*Config)(nil)

func (pc *Config) AutoCheck() (err error) {
	if pc.Protocol != "" {
		_, err = Find(pc.Protocol)
	}
	return err
}

func (pc *Config) Proxy(ctx context.Context, proxyConn *xnet.ConnNode, target string) (*xnet.ConnNode, error) {
	if pc == nil || pc.Protocol == "" {
		return proxyConn, nil
	}
	d, err := Find(pc.Protocol)
	if err != nil {
		return nil, err
	}
	return d.Proxy(ctx, proxyConn, pc, target)
}

func SetOptConfig(opt xoption.Writer, proxy *Config) {
	opt.Set(OptKeyProxy, proxy)
}

func OptConfig(opt xoption.Reader) *Config {
	return xoption.GetAsDefault[*Config](opt, OptKeyProxy, nil)
}

type Driver interface {
	Protocol() string

	// Proxy 创建代理连接
	// target: 被代理的目标地址(host:port)
	Proxy(ctx context.Context, proxyConn *xnet.ConnNode, c *Config, target string) (*xnet.ConnNode, error)
}

var drivers = map[string]Driver{}

func Register(driver Driver) error {
	protocol := driver.Protocol()
	if protocol == "" {
		return fmt.Errorf("invalid proxy protocol %q", protocol)
	}
	if _, ok := drivers[protocol]; ok {
		return fmt.Errorf("proxy driver %q already registered", driver.Protocol())
	}
	drivers[protocol] = driver
	return nil
}

func Find(protocol string) (Driver, error) {
	d, ok := drivers[protocol]
	if ok {
		return d, nil
	}
	return nil, fmt.Errorf("proxy driver %q not registered", protocol)
}

// Proxy  对 target 的地址进行代理
// proxyConn: 和代理服务器的连接
// target: 被代理的地址，其中的 HostPort 是必须是正常的值
func Proxy(ctx context.Context, d Driver, proxyConn *xnet.ConnNode, c *Config, target xnet.AddrNode) (*xnet.ConnNode, error) {
	conn, err := d.Proxy(ctx, proxyConn, c, target.HostPort)
	if err != nil {
		return proxyConn, err
	}
	return &xnet.ConnNode{
		Addr: target, // 将地址替换为被代理的地址
		Conn: conn.Conn,
	}, nil
}
