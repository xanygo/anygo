//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-21

package xsmtp

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/xanygo/anygo/ds/xcast"
	"github.com/xanygo/anygo/ds/xmap"
	xoption2 "github.com/xanygo/anygo/ds/xoption"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xdial"
)

func init() {
	handler := xdial.HandshakeFunc(handshake)
	xdial.RegisterHandshakeHandler(Protocol, handler)
}

// 创建连接后，和 smtp server 握手
func handshake(ctx context.Context, conn *xnet.ConnNode, opt xoption2.Reader) (xdial.HandshakeReply, error) {
	serverName := conn.Addr.Host()
	client, err := smtp.NewClient(conn, serverName)
	if err != nil {
		return nil, err
	}
	cfg := xoption2.Extra(opt, Protocol)
	var localName, userName, password string
	var startTLS bool = true // 默认为 true
	xmap.Range[string, any](cfg, func(k string, v any) bool {
		var ok bool
		switch k {
		case "LocalName":
			localName, ok = xcast.String(v)
		case "Username":
			userName, ok = xcast.String(v)
		case "Password":
			password, ok = xcast.String(v)
		case "StartTLS":
			startTLS, ok = xcast.Bool(v)
		default:
			ok = true
		}

		if !ok {
			err = fmt.Errorf("invalid field %s.%s=%#v", Protocol, k, v)
		}
		return ok
	})
	if err != nil {
		return nil, err
	}
	if localName != "" {
		if err = client.Hello(localName); err != nil {
			return nil, err
		}
	}

	if startTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			tc := &tls.Config{
				ServerName: serverName,
			}
			if err = client.StartTLS(tc); err != nil {
				client.Close()
				return nil, fmt.Errorf("when STARTTLS: %w", err)
			}
		}
	}

	if userName != "" {
		var auth smtp.Auth
		if ok, auths := client.Extension("AUTH"); ok {
			if strings.Contains(auths, "CRAM-MD5") {
				auth = smtp.CRAMMD5Auth(userName, password)
			} else if strings.Contains(auths, "LOGIN") &&
				!strings.Contains(auths, "PLAIN") {
				// 类似 PLAIN，但分两步发送用户名和密码,常见在 Microsoft/Exchange、Postfix
				auth = &loginAuth{
					username: userName,
					password: password,
					host:     serverName,
				}
			} else {
				auth = smtp.PlainAuth("", userName, password, serverName)
			}
		}
		if auth != nil {
			if err = client.Auth(auth); err != nil {
				client.Close()
				return nil, err
			}
		}
	}
	return &handshakeReply{
		username: userName,
		client:   client,
	}, nil
}

var _ xdial.HandshakeReply = (*handshakeReply)(nil)

type handshakeReply struct {
	username string
	client   *smtp.Client
}

func (h *handshakeReply) String() string {
	return "ok"
}

func (h *handshakeReply) Desc() string {
	return "ok"
}

var _ smtp.Auth = (*loginAuth)(nil)

type loginAuth struct {
	username string
	password string
	host     string
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if !server.TLS {
		advertised := false
		for _, mechanism := range server.Auth {
			if mechanism == "LOGIN" {
				advertised = true
				break
			}
		}
		if !advertised {
			return "", nil, errors.New("unencrypted connection")
		}
	}
	if server.Name != a.host {
		return "", nil, fmt.Errorf("invalid host name %q, expect %q", a.host, server.Name)
	}
	return "LOGIN", nil, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}

	switch {
	case bytes.Equal(fromServer, []byte("Username:")):
		return []byte(a.username), nil
	case bytes.Equal(fromServer, []byte("Password:")):
		return []byte(a.password), nil
	default:
		return nil, fmt.Errorf("unexpected server challenge: %q", fromServer)
	}
}
