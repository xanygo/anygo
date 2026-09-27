//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-15

package xproxy

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"github.com/xanygo/anygo/xnet"
)

var _ Driver = (*httpProxy)(nil)

type httpProxy struct {
}

func (h *httpProxy) Protocol() string {
	return "HTTP"
}

func (h *httpProxy) Proxy(ctx context.Context, proxyConn *xnet.ConnNode, c *Config, target string) (*xnet.ConnNode, error) {
	return doHttpProxy(ctx, proxyConn, c, target)
}

func doHttpProxy(ctx context.Context, proxyConn *xnet.ConnNode, c *Config, target string) (*xnet.ConnNode, error) {
	// todo 修改实现，authType 也采用驱动注册方式
	hello, err := getHTTPHelloRequest(c, proxyConn.Addr, target)
	if err != nil {
		return nil, err
	}
	_, err = proxyConn.Conn.Write(hello.Bytes())
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, context.Cause(ctx)
	default:
	}

	bio := bufio.NewReader(proxyConn.Conn)
	resp, err := http.ReadResponse(bio, nil)
	if err != nil {
		return nil, err
	}
	// 代理服务器应该响应：HTTP/1.1 200 Connection Established
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad proxy response status: %d  %s", resp.StatusCode, resp.Status)
	}
	return proxyConn, nil
}

type httpsProxy struct {
}

func (h *httpsProxy) Protocol() string {
	return "HTTPS"
}

func (h *httpsProxy) Proxy(ctx context.Context, proxyConn *xnet.ConnNode, c *Config, target string) (*xnet.ConnNode, error) {
	cfg, err := c.TLS.Parser()
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		cfg = &tls.Config{
			ServerName: proxyConn.Addr.Host(),
		}
	} else {
		cfg = cfg.Clone()
		cfg.ServerName = proxyConn.Addr.Host()
	}

	tc := tls.Client(proxyConn.Conn, cfg)
	if err = tc.HandshakeContext(ctx); err != nil {
		return nil, err
	}
	newConn := proxyConn.Clone()
	newConn.Conn = tc

	return doHttpProxy(ctx, newConn, c, target)
}

func getHTTPHelloRequest(c *Config, proxyAddr xnet.AddrNode, target string) (*bytes.Buffer, error) {
	if target == "" {
		return nil, errors.New("missing proxy target")
	}
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "%s %s HTTP/1.1\r\n", http.MethodConnect, target)
	fmt.Fprintf(buf, "Host: %s\r\n", proxyAddr.HostPort)
	fmt.Fprintf(buf, "User-Agent: %s\r\n", xnet.UserAgent)
	buf.WriteString("Proxy-Connection: keep-alive\r\n")

	var err error
	if c.Username != "" {
		switch c.AuthType {
		case "", "Basic":
			code := base64.StdEncoding.EncodeToString([]byte(c.Username + ":" + c.Password))
			_, err = buf.WriteString("Proxy-Authorization: Basic " + code + "\r\n")
		default:
			return nil, fmt.Errorf("not support proxy AuthType %q", c.AuthType)
		}
	}
	buf.WriteString("\r\n")
	return buf, err
}

func init() {
	Register(&httpProxy{})
	Register(&httpsProxy{})
}
