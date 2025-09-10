//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-10

package xhttpc

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xio"
)

type HandlerFunc func(ctx context.Context, resp *http.Response) error

type handlerFuncs []HandlerFunc

func (hfs handlerFuncs) Handle(ctx context.Context, resp *http.Response) error {
	for _, f := range hfs {
		if err := f(ctx, resp); err != nil {
			return err
		}
	}
	return nil
}

func HandlerFuncs(hfs ...HandlerFunc) HandlerFunc {
	return handlerFuncs(hfs).Handle
}

func HandlerLimitBody(size int64) HandlerFunc {
	return func(ctx context.Context, resp *http.Response) error {
		resp.Body = xio.LimitReaderCloser(resp.Body, size)
		return nil
	}
}

func HandlerStatusIn(codes ...int) HandlerFunc {
	mp := make(map[int]bool, len(codes))
	for _, code := range codes {
		mp[code] = true
	}
	return func(ctx context.Context, resp *http.Response) error {
		if !mp[resp.StatusCode] {
			return fmt.Errorf("invalid status code %d", resp.StatusCode)
		}
		return nil
	}
}

func HandlerStatusRange(begin int, end int) HandlerFunc {
	return func(ctx context.Context, resp *http.Response) error {
		if resp.StatusCode < begin || resp.StatusCode > end {
			return fmt.Errorf("invalid status code %d", resp.StatusCode)
		}
		return nil
	}
}

func HandlerDecodeBody(dc xcodec.Decoder, a any) HandlerFunc {
	return func(ctx context.Context, resp *http.Response) error {
		status := resp.StatusCode
		if status < 200 || status >= 500 {
			return fmt.Errorf("invalid status code %d", status)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return dc.Decode(body, a)
	}
}

func HandlerJSONBody(a any) HandlerFunc {
	return HandlerDecodeBody(xcodec.JSON, a)
}
