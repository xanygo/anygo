//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-27

package xnet

import (
	"context"
	"net"
	"time"

	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/internal/zslice"
	"github.com/xanygo/anygo/xpp"
)

type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type DialContextFunc func(ctx context.Context, network, address string) (net.Conn, error)

func (df DialContextFunc) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return df(ctx, network, address)
}

var defaultDialer = &xsync.OnceInit[Dialer]{
	New: func() Dialer {
		return &HedgingDialer{}
	},
}

func SetDefaultDialer(d Dialer) {
	defaultDialer.Store(d)
}

// DefaultDialer 默认的拨号器,10 秒超时
func DefaultDialer() Dialer {
	return defaultDialer.Load()
}

// DialContext 使用 DefaultDialer 拨号，会触发 DialerInterceptor 相关逻辑
func DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return DialContextWith(ctx, DefaultDialer(), network, address)
}

// DialContextWith 使用指定的Dialer 拨号，会触发 DialerInterceptor 相关逻辑
func DialContextWith(ctx context.Context, d Dialer, network, address string) (net.Conn, error) {
	if d == nil {
		d = DefaultDialer()
	}
	its := zslice.SafeMerge(globalDialIts, ITsFromContext[*DialerInterceptor](ctx))
	return its.Execute(ctx, d.DialContext, network, address, 0)
}

var globalDialIts dialerInterceptors

// 在 interceptor.go 里统一用 WithInterceptor 注册
func registerDialerITs(its ...*DialerInterceptor) {
	globalDialIts = append(globalDialIts, its...)
}

// HedgingDialer 默认带有请求对冲功能的拨号器
type HedgingDialer struct {
	// Invoker 可选，底层拨号器,当为 nil 时，会使用 net.Dialer
	Invoker Dialer

	// Resolver 可选，dns 解析器，当为 nil 时，会使用 DefaultResolver
	Resolver Resolver

	// Timeout 可选，超时时间, 默认为 0 - 不超时
	Timeout time.Duration
}

func (d *HedgingDialer) DialContext(ctx context.Context, network, address string) (conn net.Conn, err error) {
	if d.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, d.Timeout)
		defer cancel()
	}

	conn, err = d.doDial(ctx, network, address)

	if err != nil {
		return nil, err
	}
	return NewContextConn(ctx, conn), nil
}

func splitHostPort(hostPort string) (host string, port string, err error) {
	host, port, err = net.SplitHostPort(hostPort)
	if err != nil {
		return "", "", err
	}

	if len(host) == 0 {
		return "", "", &net.AddrError{
			Err:  "empty host",
			Addr: hostPort,
		}
	}

	return host, port, nil
}

func (d *HedgingDialer) doDial(ctx context.Context, network string, address string) (net.Conn, error) {
	if address == dummyIP {
		return nil, errDialDummy
	}
	nt := Network(network).Resolver()
	if nt.IsIP() {
		host, port, err := splitHostPort(address)
		if err != nil {
			return nil, err
		}
		var ips []net.IP
		if ip, _ := ParseIPZone(host); ip != nil {
			ips = []net.IP{ip}
		} else {
			ips, err = LookupIPWith(ctx, d.Resolver, string(nt), host)
			if err != nil {
				return nil, err
			}
		}
		return d.dialHedging(ctx, network, ips, port)
	}
	return d.dialStd(ctx, network, address)
}

var zeroDialer = &net.Dialer{}

func (d *HedgingDialer) dialStd(ctx context.Context, network string, address string) (net.Conn, error) {
	conn, err := zeroDialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	return NewContextConn(ctx, conn), nil
}

func (d *HedgingDialer) dialHedging(ctx context.Context, network string, ips []net.IP, port string) (net.Conn, error) {
	select {
	case <-ctx.Done():
		return nil, context.Cause(ctx)
	default:
	}
	hr := &xpp.Hedging[net.Conn]{
		Main: func(ctx context.Context) (net.Conn, error) {
			hostPort := net.JoinHostPort(ips[0].String(), port)
			return d.dialStd(ctx, network, hostPort)
		},
		CallNext: func(ctx context.Context, value net.Conn, err error) bool {
			return err != nil
		},
	}
	maxDelay := d.Timeout
	if maxDelay == 0 {
		maxDelay = 10 * time.Second
	}
	if dl, has := ctx.Deadline(); has {
		maxDelay = min(time.Until(dl), maxDelay)
	}
	minDelay := maxDelay / 2
	part := minDelay / time.Duration(len(ips))
	// 发送 backup request 请求
	// 请求平均分布于超时时间的后半段
	for idx, ip := range ips[1:] {
		hr.Add(minDelay+time.Duration(idx)*part, func(ctx context.Context) (net.Conn, error) {
			hostPort := net.JoinHostPort(ip.String(), port)
			return d.dialStd(ctx, network, hostPort)
		})
	}
	return hr.Run(ctx)
}

type AfterDialContextFunc func(ctx context.Context, network string, address string, conn net.Conn, err error) (net.Conn, error)

func (a AfterDialContextFunc) IT() *DialerInterceptor {
	return &DialerInterceptor{
		AfterDialContext: a,
	}
}

var _ Interceptor = (*DialerInterceptor)(nil)

// DialerInterceptor  拨号器的拦截器
type DialerInterceptor struct {
	DialContext func(ctx context.Context, network string, address string, invoker DialContextFunc) (net.Conn, error)

	BeforeDialContext func(ctx context.Context, network string, address string) (c context.Context, nt string, ad string)

	AfterDialContext func(ctx context.Context, network string, address string, conn net.Conn, err error) (net.Conn, error)
}

func (d DialerInterceptor) Interceptor() {}

type dialerInterceptors []*DialerInterceptor

func (dhs dialerInterceptors) Execute(ctx context.Context, invoker DialContextFunc, network, address string, idx int) (conn net.Conn, err error) {
	dialIdx := -1
	afterIdx := -1
	for i := range dhs {
		item := dhs[i]
		if item == nil {
			continue
		}
		if item.BeforeDialContext != nil {
			ctx, network, address = item.BeforeDialContext(ctx, network, address)
		}
		if dialIdx == -1 && item.DialContext != nil {
			dialIdx = i
		}
		if afterIdx == -1 && item.AfterDialContext != nil {
			afterIdx = i
		}
	}
	if dialIdx == -1 {
		conn, err = invoker(ctx, network, address)
	} else {
		conn, err = dhs.CallDialContext(ctx, network, address, invoker, dialIdx)
	}

	if afterIdx != -1 {
		for i := afterIdx; i < len(dhs); i++ {
			item := dhs[i]
			if item != nil && item.AfterDialContext != nil {
				conn, err = item.AfterDialContext(ctx, network, address, conn, err)
			}
		}
	}
	return conn, err
}

func (dhs dialerInterceptors) CallDialContext(ctx context.Context, network, address string, invoker DialContextFunc, idx int) (conn net.Conn, err error) {
	for ; idx < len(dhs); idx++ {
		if dhs[idx].DialContext != nil {
			break
		}
	}
	if len(dhs) == 0 || idx >= len(dhs) {
		return invoker(ctx, network, address)
	}
	return dhs[idx].DialContext(ctx, network, address, func(ctx context.Context, network, address string) (net.Conn, error) {
		return dhs.CallDialContext(ctx, network, address, invoker, idx+1)
	})
}
