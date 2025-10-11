//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-27

package xnet

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/internal/zslice"
	"github.com/xanygo/anygo/store/xcache"
	"github.com/xanygo/anygo/xattr"
	"github.com/xanygo/anygo/xlog"
	"github.com/xanygo/anygo/xnet/internal"
	"github.com/xanygo/anygo/xpp"
)

// Resolver 名字解析的接口定义
type Resolver interface {
	// LookupIP 根据传入的地址，查询返回其所有 IP 地址列表
	//
	// 当返回的 error == nil 时，[]net.IP 总是不为空
	LookupIP(ctx context.Context, network string, host string) ([]net.IP, error)
}

// LookupIPFunc 名字解析的方法定义
type LookupIPFunc func(ctx context.Context, network string, host string) ([]net.IP, error)

func (lf LookupIPFunc) LookupIP(ctx context.Context, network string, host string) ([]net.IP, error) {
	return lf(ctx, network, host)
}

var defaultResolver = &xsync.OnceInit[Resolver]{
	New: func() Resolver {
		return &ResolverImpl{}
	},
}

func DefaultResolver() Resolver {
	return defaultResolver.Load()
}

func SetDefaultResolver(r Resolver) {
	defaultResolver.Store(r)
}

var globalResolverITs resolverInterceptors

// 在 interceptor.go 里统一用 RegisterIntercotor 注册
func registerResolverIT(its ...*ResolverInterceptor) {
	globalResolverITs = append(globalResolverITs, its...)
}

// LookupIP 将域名解析为 IP 地址列表
func LookupIP(ctx context.Context, network string, host string) ([]net.IP, error) {
	return LookupIPWith(ctx, DefaultResolver(), network, host)
}

func LookupIPWith(ctx context.Context, re Resolver, network string, host string) ([]net.IP, error) {
	if re == nil {
		re = DefaultResolver()
	}
	its := zslice.SafeMerge(globalResolverITs, ITsFromContext[*ResolverInterceptor](ctx))
	return its.Execute(ctx, re.LookupIP, network, host)
}

// ResolverImpl 默认的 Resolver 实现，带有缓存
type ResolverImpl struct {
	// Invoker 可选，实际查询名字的组件，当为 nil 时，会使用标准库的 net.DefaultResolver
	Invoker Resolver

	// Interceptors 可选，拦截器，先注册的先执行
	Interceptors []*ResolverInterceptor

	// CacheTTL 结果缓存时间,当 > 0 时缓存生效
	//  若此无有效值，会尝试读取环境变量 AnyGo_Resolver_CaChe_TTL 的值，如  "3s" 表示缓存有效期 3 秒。
	//  若上述两者均无有效值，最终会使用默认值 30 秒。
	//  若值为 -1，则不缓存
	CacheTTL time.Duration

	// Cache 缓存对象，可选，当为 nil 时，会使用 LRU 缓存
	Cache xcache.Cache[string, xcache.ValueError[[]net.IP]]

	// 最后被访问的列表
	lastVisitLRU xsync.OnceDoValue[*xmap.LRU[string, time.Time]]

	flushTask xpp.SoloTask

	cacheOnce xsync.OnceDoValue[*xcache.Reader[string, []net.IP]]
}

// LookupIP Lookup IP
func (r *ResolverImpl) LookupIP(ctx context.Context, network string, host string) (ips []net.IP, err error) {
	if ip, _ := internal.ParseIPZone(host); ip != nil {
		return []net.IP{ip}, nil
	}
	if ttl := r.getTTL(); ttl > 0 {
		key := network + "@" + host
		r.getVisitLRU().Set(key, time.Now())
		r.flushTask.Run(r.doFlush, resolverFlushLife, max(ttl/3, time.Second))
	}
	its := resolverInterceptors(r.Interceptors)
	return its.Execute(ctx, r.lookupIP, network, host)
}

const resolverFlushLife = 5 * time.Minute

func (r *ResolverImpl) lookupIP(ctx context.Context, network string, host string) ([]net.IP, error) {
	cache := r.getCacheOnce()
	if cache == nil {
		return r.getStdResolver().LookupIP(ctx, network, host)
	}
	cacheKey := network + "@" + host
	return cache.Get(ctx, cacheKey)
}

func (r *ResolverImpl) getStdResolver() Resolver {
	if r.Invoker != nil {
		return r.Invoker
	}
	return net.DefaultResolver
}

// WithInterceptor 注册拦截器
func (r *ResolverImpl) WithInterceptor(its ...*ResolverInterceptor) {
	r.Interceptors = append(r.Interceptors, its...)
}

func (r *ResolverImpl) getVisitLRU() *xmap.LRU[string, time.Time] {
	return r.lastVisitLRU.Do(r.initVisitLRU)
}

func (r *ResolverImpl) doFlush() {
	list := r.getVisitLRU().Map()
	if len(list) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), resolverFlushLife)
	defer cancel()

	now := time.Now()
	expiredTime := now.Add(-5 * time.Minute)

	for k, v := range list {
		// 对近一段时间内有访问的，持续刷新缓存
		if v.Before(expiredTime) {
			continue
		}
		start := time.Now()
		result, err1 := r.getCacheOnce().Flush(ctx, k)
		if xattr.RunMode() == xattr.ModeDebug {
			xlog.Debug(ctx, "ResolverImpl.doFlush",
				xlog.String("key", k),
				xlog.Any("result", result),
				xlog.Any("error", err1),
				xlog.String("cost", time.Since(start).String()),
			)
		}
	}
}

func (r *ResolverImpl) initVisitLRU() *xmap.LRU[string, time.Time] {
	return xmap.NewLRU[string, time.Time](256)
}

func (r *ResolverImpl) getCacheOnce() *xcache.Reader[string, []net.IP] {
	return r.cacheOnce.Do(r.createCacheReader)
}

// 创建缓存对象：只会被调用一次
func (r *ResolverImpl) createCacheReader() *xcache.Reader[string, []net.IP] {
	ttl := r.getTTL()
	if ttl < 0 {
		return nil
	}
	cc := r.Cache
	if cc == nil {
		cc = xcache.NewLRU[string, xcache.ValueError[[]net.IP]](10000)
		xcache.Registry().TryRegister("sys:ResolverLRU", cc)
	}
	cache := &xcache.Reader[string, []net.IP]{
		Cache:   cc,
		TTL:     ttl,
		FailTTL: min(max(ttl/10, 100*time.Millisecond), time.Second),
		New: func(ctx context.Context, key string) ([]net.IP, error) {
			network, host, _ := strings.Cut(key, "@")
			return r.getStdResolver().LookupIP(ctx, network, host)
		},
	}
	return cache
}

func (r *ResolverImpl) getTTL() time.Duration {
	if r.CacheTTL != 0 {
		return r.CacheTTL
	}
	return r.getTTLFromEnv()
}

func (r *ResolverImpl) getTTLFromEnv() time.Duration {
	return envResolverCacheTTL.Load()
}

var envResolverCacheTTL = &xsync.OnceInit[time.Duration]{
	New: func() time.Duration {
		val := os.Getenv("AnyGo_Resolver_CaChe_TTL")
		ts, _ := time.ParseDuration(val)
		if ts >= time.Millisecond {
			return ts
		}
		return 30 * time.Second
	},
}

type AfterLookupIPFunc func(ctx context.Context, network string, host string, ips []net.IP, err error) ([]net.IP, error)

func (a AfterLookupIPFunc) IT() *ResolverInterceptor {
	return &ResolverInterceptor{
		AfterLookupIP: a,
	}
}

type ResolverInterceptor struct {
	LookupIP func(ctx context.Context, network string, host string, invoker LookupIPFunc) ([]net.IP, error)

	// BeforeLookupIP 解析前的回调，可以对 ctx、network、 host 更新
	BeforeLookupIP func(ctx context.Context, network string, host string) (context.Context, string, string)

	// AfterLookupIP 解析完成后的回调，可以对 ips 和 err 更新
	AfterLookupIP func(ctx context.Context, network string, host string, ips []net.IP, err error) ([]net.IP, error)
}

var _ Interceptor = (*ResolverInterceptor)(nil)

func (r *ResolverInterceptor) Interceptor() {}

type resolverInterceptors []*ResolverInterceptor

func (rhs resolverInterceptors) Execute(ctx context.Context, invoker LookupIPFunc, network string, host string) (ips []net.IP, err error) {
	lookIdx := -1
	afterIdx := -1
	for i := 0; i < len(rhs); i++ {
		item := rhs[i]
		if item == nil {
			continue
		}
		if item.BeforeLookupIP != nil {
			ctx, network, host = item.BeforeLookupIP(ctx, network, host)
		}
		if lookIdx == -1 && item.LookupIP != nil {
			lookIdx = i
		}
		if afterIdx == -1 && item.AfterLookupIP != nil {
			afterIdx = i
		}
	}
	if lookIdx == -1 {
		ips, err = invoker(ctx, network, host)
	} else {
		ips, err = rhs.CallLookupIP(ctx, network, host, invoker, lookIdx)
	}
	if afterIdx != -1 {
		for ; afterIdx < len(rhs); afterIdx++ {
			item := rhs[afterIdx]
			if item != nil && item.AfterLookupIP != nil {
				ips, err = item.AfterLookupIP(ctx, network, host, ips, err)
			}
		}
	}
	return ips, err
}

func (rhs resolverInterceptors) CallLookupIP(ctx context.Context, network string, host string, invoker LookupIPFunc,
	idx int) (ips []net.IP, err error) {
	for ; idx < len(rhs); idx++ {
		if rhs[idx].LookupIP != nil {
			break
		}
	}
	if len(rhs) == 0 || idx >= len(rhs) {
		return invoker(ctx, network, host)
	}

	return rhs[idx].LookupIP(ctx, network, host, func(ctx context.Context, network string, host string) ([]net.IP, error) {
		return rhs.CallLookupIP(ctx, network, host, invoker, idx+1)
	})
}

// BlockPrivateResolution 禁止解析出私有和环回地址
func BlockPrivateResolution(ctx context.Context, network string, host string, ips []net.IP, err error) ([]net.IP, error) {
	for i := 0; i < len(ips); i++ {
		ip := ips[i]
		if ip.IsPrivate() || ip.IsLoopback() {
			return ips, fmt.Errorf("blocked private ip: %s", ip.String())
		}
	}
	return ips, err
}
