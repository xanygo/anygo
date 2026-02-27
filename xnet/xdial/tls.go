//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-02-26

package xdial

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/xanygo/anygo/ds/xmetric"
	"github.com/xanygo/anygo/ds/xoption"
	"github.com/xanygo/anygo/xnet"
)

func tlsUpgrade(ctx context.Context, conn *xnet.ConnNode, opt xoption.Reader, target xnet.AddrNode) (nc *xnet.ConnNode, err error) {
	tc := xoption.GetTLSConfig(opt)

	its := allITs(ctx)
	for _, it := range its {
		if it.BeforeTlsHandshake == nil {
			continue
		}
		ctx, conn, opt, target, tc = it.BeforeTlsHandshake(ctx, conn, opt, target, tc)
	}
	defer func() {
		for _, it := range its {
			if it.AfterTlsHandshake == nil {
				continue
			}
			nc, err = it.AfterTlsHandshake(ctx, conn, opt, target, tc, nc, err)
		}
	}()

	if tc == nil {
		return conn, nil
	}

	ctx, span := xmetric.Start(ctx, "TLSHandshake")
	defer func() {
		span.RecordError(err)
		span.End()
	}()
	tc = tc.Clone()
	if tc.ServerName == "" {
		tc.ServerName = target.Host()
	}
	if tc.MinVersion == 0 {
		tc.MinVersion = tls.VersionTLS12
	}
	span.SetAttributes(
		xmetric.AnyAttr("ServerName", tc.ServerName),
		xmetric.AnyAttr("SkipVerify", tc.InsecureSkipVerify),
	)
	tlsConn := tls.Client(conn.Outer(), tc)

	timeout := xoption.HandshakeTimeout(opt)
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	if err = tlsConn.HandshakeContext(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("%w, ServerName=%q", err, tc.ServerName)
	}
	conn.AddWrap(tlsConn)
	return conn, nil
}

// TlsHandshake Tls 握手的逻辑
func TlsHandshake(ctx context.Context, conn *xnet.ConnNode, opt xoption.Reader, target xnet.AddrNode) (nc *xnet.ConnNode, err error) {
	return tlsUpgrade(ctx, conn, opt, target)
}
