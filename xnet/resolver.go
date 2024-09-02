//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-27

package xnet

import (
	"context"
	"net"
	"os"
	"strings"
	"time"

	"github.com/xanygo/anygo/xcache"
	"github.com/xanygo/anygo/xnet/internal"
	"github.com/xanygo/anygo/xsync"
)

type (
	// Resolver 名字解析的接口定义
	Resolver interface {
		// LookupIP 根据传入的地址，查询返回其所有 IP 地址列表
		//
		// 当返回的 error == nil 时，[]net.IP 总是不为空
		LookupIP(ctx context.Context, network string, host string) ([]net.IP, error)
	}

	// LookupIPFunc 名字解析的方法定义
	LookupIPFunc func(ctx context.Context, network string, host string) ([]net.IP, error)
)

var DefaultResolver Resolver = &ResolverImpl{}

func LookupIP(ctx context.Context, network string, host string) ([]net.IP, error) {
	return DefaultResolver.LookupIP(ctx, network, host)
}

// ResolverImpl 默认的 Resolver 实现，带有缓存
type ResolverImpl struct {
	// Invoker 可选，实际查询名字的组件，当为 nil 时，会使用标准库的 net.DefaultResolver
	Invoker Resolver

	// Interceptors 可选，拦截器，先注册的先执行
	Interceptors []*ResolverInterceptor

	// CacheTTL 结果缓存时间,当 > 0 时缓存生效
	//  若此无有效值，会尝试读取环境变量 AnyGo_Resolver_CaChe_TTL 的值，如  "3s" 表示缓存有效期 3 秒。
	//  若上述两者均无有效值，最终会使用默认值 1分钟。
	CacheTTL time.Duration

	cacheOnce xsync.OnceDoValue[*xcache.Reader[string, []net.IP]]
}

func (r *ResolverImpl) getInterceptors(ctx context.Context) resolverInterceptors {
	return Interceptors[*ResolverInterceptor](ctx, r.Interceptors)
}

// LookupIP Lookup IP
func (r *ResolverImpl) LookupIP(ctx context.Context, network string, host string) (ips []net.IP, err error) {
	its := r.getInterceptors(ctx)
	lookIdx := -1
	afterIdx := -1
	for i := 0; i < len(its); i++ {
		if its[i].BeforeLookupIP != nil {
			ctx, network, host = its[i].BeforeLookupIP(ctx, network, host)
		}
		if lookIdx == -1 && its[i].LookupIP != nil {
			lookIdx = i
		}
		if afterIdx == -1 && its[i].AfterLookupIP != nil {
			afterIdx = i
		}
	}
	if lookIdx == -1 {
		ips, err = r.lookupIP(ctx, network, host)
	} else {
		ips, err = its.CallLookupIP(ctx, network, host, r.lookupIP, lookIdx)
	}
	if afterIdx != -1 {
		for ; afterIdx < len(its); afterIdx++ {
			if its[afterIdx].AfterLookupIP != nil {
				ips, err = its[afterIdx].AfterLookupIP(ctx, network, host, ips, err)
			}
		}
	}
	return ips, err
}

func (r *ResolverImpl) lookupIP(ctx context.Context, network string, host string) ([]net.IP, error) {
	if ip, _ := internal.ParseIPZone(host); ip != nil {
		return []net.IP{ip}, nil
	}
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

func (r *ResolverImpl) getCacheOnce() *xcache.Reader[string, []net.IP] {
	return r.cacheOnce.Do(r.getCache)
}

func (r *ResolverImpl) getCache() *xcache.Reader[string, []net.IP] {
	ttl := r.getTTL()
	if ttl <= 0 {
		return nil
	}
	lru := xcache.NewLRU[string, xcache.ValueError[[]net.IP]](10000)
	cache := &xcache.Reader[string, []net.IP]{
		Cache:   lru,
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
	if r.CacheTTL > 0 {
		return r.CacheTTL
	}
	return defaultResolverCacheTTL()
}

func defaultResolverCacheTTL() time.Duration {
	val := os.Getenv("AnyGo_Resolver_CaChe_TTL")
	ts, _ := time.ParseDuration(val)
	if ts > time.Second {
		return ts
	}
	return time.Minute
}

type ResolverInterceptor struct {
	LookupIP func(ctx context.Context, network string, host string, invoker LookupIPFunc) ([]net.IP, error)

	BeforeLookupIP func(ctx context.Context, network string, host string) (context.Context, string, string)
	AfterLookupIP  func(ctx context.Context, network string, host string, ips []net.IP, err error) ([]net.IP, error)
}

var _ Interceptor = (*ResolverInterceptor)(nil)

func (r *ResolverInterceptor) Interceptor() {}

type resolverInterceptors []*ResolverInterceptor

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
