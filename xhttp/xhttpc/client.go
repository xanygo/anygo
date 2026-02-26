//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-10

package xhttpc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"time"

	"github.com/xanygo/anygo/ds/xctx"
	"github.com/xanygo/anygo/ds/xhash"
	"github.com/xanygo/anygo/safely"
	"github.com/xanygo/anygo/store/xcache"
	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xnet/xrpc"
	"github.com/xanygo/anygo/xnet/xservice"
)

type Invoker interface {
	Invoke(ctx context.Context, service any, req *http.Request, handler HandlerFunc, opts ...xrpc.Option) error
}

type InvokeFunc func(ctx context.Context, service any, req *http.Request, handler HandlerFunc, opts ...xrpc.Option) error

func (in InvokeFunc) Invoke(ctx context.Context, service any, req *http.Request, handler HandlerFunc, opts ...xrpc.Option) error {
	return in(ctx, service, req, handler, opts...)
}

func Invoke(ctx context.Context, service any, req *http.Request, handler HandlerFunc, opts ...xrpc.Option) error {
	hr := &NativeRequest{
		Request: req,
	}
	resp := &Response{
		Handler: handler,
	}
	return xrpc.Invoke(ctx, service, hr, resp, opts...)
}

func Get(ctx context.Context, service any, url string, handler HandlerFunc, opts ...xrpc.Option) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	return Invoke(ctx, service, req, handler, opts...)
}

// GetAsJSON 使用 GET 读取 url 的内容（响应码必须为 200 ），并解析为 json 格式
func GetAsJSON[T any](ctx context.Context, service any, url string, opts ...xrpc.Option) (obj *T, err error) {
	obj = new(T)
	handler := Combine(StatusIn(200), JSONBody(obj))
	err = Get(ctx, service, url, handler, opts...)
	return obj, err
}

func InvokeWithCodec(ctx context.Context, service any, method string, url string, body any, ec xcodec.Encoder, handler HandlerFunc, opts ...xrpc.Option) error {
	contentType, err := xcodec.ContentType(ec)
	if err != nil {
		return err
	}

	bf, err := ec.Encode(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if len(bf) > 0 {
		// 使用 GetBody 赋值，这样当有重试的时候,才能保证每次都读取到 body
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewBuffer(bf)), nil
		}
	}
	req.Header.Set("Content-Type", contentType)
	return Invoke(ctx, service, req, handler, opts...)
}

func Post(ctx context.Context, service any, url string, ct string, body io.Reader, handler HandlerFunc, opts ...xrpc.Option) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", ct)
	return Invoke(ctx, service, req, handler, opts...)
}

func PostForm(ctx context.Context, service any, url string, body url.Values, handler HandlerFunc, opts ...xrpc.Option) error {
	return InvokeWithCodec(ctx, service, http.MethodPost, url, body, xcodec.Form, handler, opts...)
}

func PostJSON(ctx context.Context, service any, url string, body any, handler HandlerFunc, opts ...xrpc.Option) error {
	return InvokeWithCodec(ctx, service, http.MethodPost, url, body, xcodec.JSON, handler, opts...)
}

var ctxKeySkipCache = xctx.NewKey()

// SkipCache 在 context 里设置让 CachedClient 是否强制跳过缓存
func SkipCache(ctx context.Context, skip bool) context.Context {
	return context.WithValue(ctx, ctxKeySkipCache, skip)
}

func isSkipCache(ctx context.Context) bool {
	val, _ := ctx.Value(ctxKeySkipCache).(bool)
	return val
}

var _ Executor = (*http.Client)(nil)

type Executor interface {
	Do(req *http.Request) (*http.Response, error)
}

// ExecutorToInvoker 将 Executor 转换为 Invoker
//
// 使用场景：
// 默认的 Invoker 是不支持 302 跳转的，在使用 CachedClient 的时候，会有诸多限制，所以若需要跳转，则可以这样：
//
//	 hc := &http.Client{
//	    Transport: &xhttpc.Client{},
//	 }
//	 cc := xhttpc.CachedClient{
//	    Invoker: xhttpc.ExecutorToInvoker(hc),
//		// ... 其他参数...
//	 }
func ExecutorToInvoker(e Executor) Invoker {
	return InvokeFunc(func(ctx context.Context, service any, req *http.Request, handler HandlerFunc, opts ...xrpc.Option) error {
		req = req.WithContext(ctx)
		resp, err := e.Do(req)
		if err != nil {
			return err
		}
		return handler(ctx, resp)
	})
}

var _ http.RoundTripper = (*Client)(nil)

// Client 实现了 RoundTripper 的 HTTP Client
//
// 该 Client 所有方法都只会发送一次请求，不会处理更高层协议细节，例如重定向、认证或 Cookie。
// 若需要处理重定向，可以结合 http.Client 使用
type Client struct {
	Service any           // 可选，当为空时，会使用 Dummy
	Opts    []xrpc.Option // 可选，额外的 RPC Client 参数
}

func (c *Client) getService() any {
	if c.Service == nil {
		return xservice.DummyService()
	}
	return c.Service
}

func (c *Client) RoundTrip(req *http.Request) (*http.Response, error) {
	r := &http.Response{}
	handler := FetchResponse(r)
	err := Invoke(req.Context(), c.getService(), req, handler, c.Opts...)
	return r, err
}

func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	r := &http.Response{}
	handler := FetchResponse(r)
	err := Get(ctx, c.getService(), url, handler, c.Opts...)
	return r, err
}

func (c *Client) Post(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	r := &http.Response{}
	handler := FetchResponse(r)
	err := Post(ctx, c.getService(), url, contentType, body, handler, c.Opts...)
	return r, err
}

func (c *Client) PostForm(ctx context.Context, url string, data url.Values) (*http.Response, error) {
	r := &http.Response{}
	handler := FetchResponse(r)
	err := PostForm(ctx, c.getService(), url, data, handler, c.Opts...)
	return r, err
}

func (c *Client) PostJSON(ctx context.Context, url string, data any) (*http.Response, error) {
	r := &http.Response{}
	handler := FetchResponse(r)
	err := PostJSON(ctx, c.getService(), url, data, handler, c.Opts...)
	return r, err
}

// CachedClient 带有缓存的 HTTP Client,只会成功获取 response 的才会被缓存
type CachedClient struct {
	Cache   xcache.Cache[string, *StoredResponse] // 必填，缓存对象
	Request *http.Request                         // 必填，请求
	Key     string                                // 可选，缓存的 key，默认为读取 Request.URL 作为 key
	Invoker Invoker                               // 可选，用于发送请求的实体

	// 可选，缓存时间，默认 1 小时
	TTL time.Duration

	// 在cache 设置前，重新调整缓存有效期，可选
	AdjustTTL func(rs StoredResponse, ttl time.Duration) time.Duration

	// 可选，当缓存数据达到此有效期后，提前异步加载数据，默认为 0.8 * TTL
	PreFlush time.Duration

	// 可选，默认为 JSON，用于将 []byte 转换为传给 Invoke 的 result any 类型
	Decoder xcodec.Decoder

	// 可选，若不配置则默认为验证 statusCode==200
	HandlerFunc HandlerFunc

	Service string        // 可选，默认为 dummy
	Opts    []xrpc.Option // 可选
}

func (ci CachedClient) getTTL() time.Duration {
	if ci.TTL > 0 {
		return ci.TTL
	}
	return time.Hour
}

func (ci CachedClient) getAdjustTTL(rd *StoredResponse, ttl time.Duration) time.Duration {
	if ci.AdjustTTL == nil {
		return ttl
	}
	return ci.AdjustTTL(*rd, ttl)
}

func (ci CachedClient) getService() string {
	if ci.Service == "" {
		return xservice.Dummy
	}
	return ci.Service
}

func (ci CachedClient) getDecoder() xcodec.Decoder {
	if ci.Decoder == nil {
		return xcodec.JSON
	}
	return ci.Decoder
}

func (ci CachedClient) needPreFlush(cacheCreate time.Time) bool {
	preFlush := ci.PreFlush
	if preFlush < 0 {
		return false
	}
	if preFlush == 0 {
		preFlush = ci.getTTL() * 4 / 5
	}
	return time.Since(cacheCreate) > preFlush
}

func (ci CachedClient) getCacheKey() string {
	if ci.Key != "" {
		return ci.getService() + "|" + ci.Request.Method + "|" + ci.Key
	}
	return xhash.Md5(ci.getService() + "|" + ci.Request.Method + "|" + ci.Request.URL.String())
}

// DeleteCache 删除此请求对应的缓存
func (ci CachedClient) DeleteCache(ctx context.Context) error {
	return ci.Cache.Delete(ctx, ci.getCacheKey())
}

func (ci CachedClient) Invoke(ctx context.Context, result any) error {
	var cachedResponse *StoredResponse
	var err error
	key := ci.getCacheKey()

	if !isSkipCache(ctx) {
		cachedResponse, err = ci.Cache.Get(ctx, key)
	}

	decoder := ci.getDecoder()

	sr, ok1 := result.(*StoredResponse)

	if cachedResponse != nil && err == nil {
		npr := ci.needPreFlush(cachedResponse.CreateTime())
		if ok1 {
			*sr = *cachedResponse
			if npr {
				ci.doPreFlush(ctx, decoder, result, key)
			}
			return nil
		}
		err = xcodec.Decode(decoder, cachedResponse.Body, result)
		if err == nil {
			if npr {
				ci.doPreFlush(ctx, decoder, result, key)
			}
			return nil
		}
	}

	return ci.direct(ctx, decoder, result, key)
}

var preFlushDB sync.Map

func (ci CachedClient) doPreFlush(ctx context.Context, decoder xcodec.Decoder, result any, cacheKey string) {
	t := reflect.TypeOf(result)
	if t.Kind() != reflect.Pointer {
		return
	}

	// 只需要有一个 PreFlush 的操作，避免同时发起多个相同请求
	if _, loaded := preFlushDB.LoadOrStore(cacheKey, struct{}{}); loaded {
		return
	}

	go safely.RunVoid(func() {
		defer preFlushDB.Delete(cacheKey)
		// 创建一个新的对象（和 result 指向的类型一样）
		newObj := reflect.New(t.Elem()).Interface()
		ctx = context.WithoutCancel(ctx)
		_ = ci.direct(ctx, decoder, newObj, cacheKey)
	})
}

func (ci CachedClient) direct(ctx context.Context, decoder xcodec.Decoder, result any, cacheKey string) error {
	var hs handlerCombine
	rd := &StoredResponse{
		CreateAt: time.Now().Unix(),
	}
	// 在其他 Handler 之前，把状态信息读取出来
	readHeaderHandler := func(ctx context.Context, resp *http.Response) error {
		rd.StatusCode = resp.StatusCode
		rd.Header = resp.Header.Clone()
		if resp.Request != nil {
			rd.URL = resp.Request.URL.String()
		}
		return nil
	}
	hs = append(hs, readHeaderHandler)

	if ci.HandlerFunc == nil {
		hs = append(hs, StatusIn(http.StatusOK))
	} else {
		hs = append(hs, ci.HandlerFunc)
	}

	hs = append(hs, TeeReader(rd))

	sr, ok1 := result.(*StoredResponse)
	if !ok1 {
		hs = append(hs, DecodeBody(decoder, result))
	}
	inv := ci.Invoker
	if inv == nil {
		inv = InvokeFunc(Invoke)
	}
	start := time.Now()
	err := inv.Invoke(ctx, ci.getService(), ci.Request, Combine(hs...), ci.Opts...)
	rd.Cost = time.Since(start)
	if ok1 {
		*sr = *rd
	}
	if err != nil {
		return err
	}
	rd.FromCache = true
	ttl := ci.getAdjustTTL(rd, ci.getTTL())
	if ttl > 0 {
		_ = ci.Cache.Set(ctx, cacheKey, rd, ttl)
	}
	return nil
}
