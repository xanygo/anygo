//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-04-02

package dsession

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/xanygo/anygo/xmap"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xoption"
)

// HTTPUpgrade 构建一个能执行 HTTP Upgrade 逻辑的会话创建逻辑
func HTTPUpgrade(method string, uri string, protocol string) Starter {
	hd := bytes.NewBuffer(nil)
	fmt.Fprintf(hd, "%s %s HTTP/1.1\r\n", method, uri)
	fmt.Fprintf(hd, "Upgrade: %s\r\n", protocol)
	fmt.Fprint(hd, "Connection: Upgrade\r\n")
	return StartFunc(func(ctx context.Context, rw io.ReadWriter, opt xoption.Reader) (Reply, error) {
		conn, ok := rw.(*xnet.ConnNode)
		if !ok {
			return nil, errors.New("conn is not a net.ConnNode")
		}
		host := fmt.Sprintf("Host: %s\r\n\r\n", conn.Addr.HostPort)
		bf := bytes.NewBuffer(nil)
		bf.Grow(hd.Len() + len(host))
		bf.Write(hd.Bytes())
		bf.WriteString(host)
		_, err := rw.Write(bf.Bytes())
		if err != nil {
			return nil, err
		}
		reader := bufio.NewReader(conn)
		resp, err := http.ReadResponse(reader, nil)
		if err != nil {
			return nil, err
		}

		// 校验状态码是否为 101 Switching Protocols
		if resp.StatusCode != http.StatusSwitchingProtocols {
			return nil, fmt.Errorf("upgrade failed with status=%s, expect statusCode=101", resp.Status)
		}
		return nil, nil
	})
}

func httpUpgradeFactory(param map[string]any) (Starter, error) {
	if len(param) == 0 {
		return nil, errors.New("cannot create Starter by httpUpgradeFactory with empty param")
	}
	method, err := xmap.GetString(param, "Method")
	if err != nil {
		return nil, err
	}
	if method == "" {
		return nil, fmt.Errorf("invalid httpUpgradeFactory method %q", method)
	}
	uri, err := xmap.GetString(param, "URI")
	if err != nil {
		return nil, err
	}
	if uri == "" {
		return nil, fmt.Errorf("invalid httpUpgradeFactory URI %q", uri)
	}
	protocol, err := xmap.GetString(param, "Protocol")
	if err != nil {
		return nil, err
	}
	if protocol == "" {
		return nil, fmt.Errorf("invalid httpUpgradeFactory protocol %q", protocol)
	}
	return HTTPUpgrade(method, uri, protocol), nil
}

func init() {
	RegisterFactory("HTTP-Upgrade", httpUpgradeFactory)
	RegisterFactory("HTTPUpgrade", httpUpgradeFactory)
}
