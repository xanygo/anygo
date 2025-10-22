//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-22

package xsmtp

import (
	"context"
	"errors"
	"net"
	"slices"
	"strconv"

	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/xnet/xrpc"
	"github.com/xanygo/anygo/xnet/xservice"
	"github.com/xanygo/anygo/xoption"
)

type Config struct {
	Host       string        // smtp 服务器地址，必填
	Port       int           // smtp 服务器端口，必填
	Username   string        // 登录账号/发件人，必填
	Password   string        // 登录的密码
	NoStartTLS bool          // 是否不使用 StartTLS 功能，可选
	Options    []xrpc.Option // 其他的配置项目，可选

	once xsync.OnceDoValue[[]xrpc.Option]
}

func (c *Config) initOption() []xrpc.Option {
	opt1 := xrpc.OptHostPort(net.JoinHostPort(c.Host, strconv.Itoa(c.Port)))
	data := map[string]any{
		"Username": c.Username,
		"Password": c.Password,
		"StartTLS": !c.NoStartTLS,
	}
	opt2 := xrpc.OptOptionSetter(func(o xoption.Option) {
		xoption.SetExtra(o, Protocol, data)
		xoption.SetProtocol(o, Protocol)
	})
	opts := slices.Clone(c.Options)
	opts = append(opts, opt1, opt2)
	return opts
}

func (c *Config) check() error {
	if c.Host == "" {
		return errors.New("host is required")
	}
	if c.Port == 0 {
		return errors.New("port is required")
	}
	if c.Username == "" {
		return errors.New("username is required")
	}
	return nil
}

func (c *Config) Send(ctx context.Context, m *Mail) error {
	if err := c.check(); err != nil {
		return err
	}
	opts := c.once.Do(c.initOption)
	return Send(ctx, xservice.Dummy, m, opts...)
}

func Send(ctx context.Context, service any, m *Mail, opts ...xrpc.Option) error {
	return xrpc.Invoke(ctx, service, m, xrpc.DiscardResponse(), opts...)
}
