//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-03

package xrpc

import (
	"context"
	"io"
	"net"

	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xnet/xservice"
)

type Client interface {
	Invoke(ctx context.Context, service string, req Request, resp Response, opts ...ClientOption) error
}

type Request interface {
	String() string
	APIName() string
	WriteTo(ctx context.Context, w net.Conn, opt *xservice.Option) error
}

type Response interface {
	String() string
	io.ReaderFrom
	xerror.HasErrCode
	xerror.HasErrMsg
}

type ClientOption interface {
	withOption(o *xservice.Option)
}
