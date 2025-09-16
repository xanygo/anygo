//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-10

package xhttpc

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/xanygo/anygo/xlog"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xbalance"
	"github.com/xanygo/anygo/xnet/xproxy"
	"github.com/xanygo/anygo/xnet/xrpc"
	"github.com/xanygo/anygo/xnet/xservice"
	"github.com/xanygo/anygo/xoption"
)

var _ xrpc.Request = (*Request)(nil)

type Request struct {
	API    string // APIName
	Method string
	Path   string
	HTTPS  bool
	Query  url.Values
	Header http.Header
	Body   io.Reader
}

func (r *Request) Protocol() string {
	return "HTTP"
}

func (r *Request) String() string {
	return "Request:" + r.APIName()
}

func (r *Request) APIName() string {
	if r.API != "" {
		return r.API
	}
	return r.Path
}

func (r *Request) WriteTo(ctx context.Context, node *xnet.ConnNode, opt xoption.Reader) error {
	api, err := r.getURL(opt, node.Addr.HostPort)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, r.getMethod(), api, r.Body)
	if err != nil {
		return err
	}
	if req.Host == xservice.Dummy {
		req.Host = ""
	}
	setHeader(ctx, req, opt)
	return req.Write(node.Conn)
}

func (r *Request) getURL(so xoption.Reader, address string) (string, error) {
	opt := xservice.OptHTTP(so)
	u, err := url.Parse(r.Path)
	if err != nil {
		return "", err
	}
	if u.Host == xservice.DummyAddress {
		u.Host = ""
	}

	u.Scheme = "http" // 不需要区分是 HTTP 还是 HTTPS
	if opt.Host != "" {
		u.Host = opt.Host
	}
	if u.Host == "" {
		u.Host = address
	}
	return u.String(), nil
}

func (r *Request) getMethod() string {
	if r.Method == "" {
		return http.MethodGet
	}
	return r.Method
}

func (r *Request) balancer(opt xoption.Reader, u *url.URL) xbalance.Reader {
	if u.Host == "" || u.Hostname() == xservice.Dummy {
		return nil
	}
	proxy := xproxy.OptConfig(opt)
	if proxy != nil {
		return nil
	}
	hostPort := getHostPort(u)
	if hostPort != "" {
		return xbalance.NewStaticByAddr(xnet.NewAddr("tcp", hostPort))
	}
	return nil
}

func (r *Request) OptionReader(ctx context.Context, rd xoption.Reader) xoption.Reader {
	u, err := url.Parse(r.Path)
	if err != nil {
		return nil
	}
	opt := xoption.NewSimple()
	if b := r.balancer(rd, u); b != nil {
		xbalance.OptSetDownstream(opt, b)
	}
	if proxy := xproxy.OptConfig(opt); proxy != nil {
		if hostPort := getHostPort(u); hostPort != "" {
			xbalance.OptSetTarget(opt, xbalance.NewStaticByAddr(xnet.NewAddr("tcp", hostPort)))
		}
	}
	return opt.Value()
}

var _ xrpc.Request = (*NativeRequest)(nil)

type NativeRequest struct {
	API     string
	Request *http.Request
}

func (h *NativeRequest) Protocol() string {
	return "HTTP"
}

func (h *NativeRequest) String() string {
	return "HTTPRequestNative:" + h.APIName()
}

func (h *NativeRequest) APIName() string {
	if h.API != "" {
		return h.API
	}
	return h.Request.URL.Path
}

func (h *NativeRequest) WriteTo(ctx context.Context, node *xnet.ConnNode, opt xoption.Reader) error {
	req := h.Request.WithContext(ctx)
	if req.Host == xservice.Dummy {
		req.Host = ""
	}
	if req.Host == "" {
		req.Host = node.Addr.Host()
	}
	setHeader(ctx, req, opt)
	return req.Write(node.Conn)
}

func (h *NativeRequest) balancer(opt xoption.Reader) xbalance.Reader {
	host := h.Request.URL.Hostname()
	if host == xservice.Dummy {
		return nil
	}
	proxy := xproxy.OptConfig(opt)
	if proxy != nil {
		// 有代理的情况下，应该连接代理
		return nil
	}
	if hostPort := getHostPort(h.Request.URL); hostPort != "" {
		return xbalance.NewStaticByAddr(xnet.NewAddr("tcp", hostPort))
	}
	return nil
}

func (h *NativeRequest) OptionReader(ctx context.Context, opt xoption.Reader) xoption.Reader {
	mp := xoption.NewSimple()
	if ap := h.balancer(opt); ap != nil {
		xbalance.OptSetDownstream(mp, ap)
	}

	if proxy := xproxy.OptConfig(opt); proxy != nil {
		if hostPort := getHostPort(h.Request.URL); hostPort != "" {
			xbalance.OptSetTarget(mp, xbalance.NewStaticByAddr(xnet.NewAddr("tcp", hostPort)))
		}
	}
	return mp.Value()
}

func setHeader(ctx context.Context, req *http.Request, opt xoption.Reader) {
	hc := xservice.OptHTTP(opt)
	if hc.Host != "" {
		req.Host = hc.Host
	}
	for k, v := range hc.Header {
		req.Header[k] = v
	}
	if req.UserAgent() == "" {
		req.Header.Set("User-Agent", xnet.UserAgent)
	}
	logID := xlog.FindLogID(ctx)
	if logID != "" {
		req.Header.Set("X-Log-ID", logID)
	}
}

func getHostPort(u *url.URL) string {
	// 此处不需要考虑 Hostname 为 dummy 的情况
	serverName := u.Hostname()
	if serverName == "" {
		return ""
	}
	port := u.Port()
	if port != "" {
		return net.JoinHostPort(serverName, port)
	}
	if u.Scheme == "https" {
		port = "443"
	} else {
		port = "80"
	}
	return net.JoinHostPort(serverName, port)
}
