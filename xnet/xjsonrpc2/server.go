//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-03-28

package xjsonrpc2

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net/http"
	"sync"

	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xio"
)

type Handler interface {
	Handle(ctx context.Context, req *Request) (result any, err error)
}

type HandlerFunc func(ctx context.Context, req *Request) (result any, err error)

func (h HandlerFunc) Handle(ctx context.Context, req *Request) (result any, err error) {
	return h(ctx, req)
}

var _ http.Handler = (*Router)(nil)
var _ Handler = (*Router)(nil)

func NewRouter() *Router {
	return &Router{
		handlers: make(map[string]Handler),
	}
}

type Router struct {
	handlers map[string]Handler
}

func (r *Router) Register(method string, h Handler) {
	if r.handlers == nil {
		r.handlers = make(map[string]Handler)
	}
	r.handlers[method] = h
}

func (r *Router) RegisterFunc(method string, fn func(ctx context.Context, req *Request) (result any, err error)) {
	r.Register(method, HandlerFunc(fn))
}

func (r *Router) Handle(ctx context.Context, req *Request) (result any, err error) {
	if len(r.handlers) == 0 {
		return nil, ErrMethodNotFound
	}
	h, ok := r.handlers[req.Method]
	if !ok {
		return nil, ErrMethodNotFound
	}
	result, err = h.Handle(ctx, req)
	if req.NoReply() {
		return nil, err
	}
	return result, err
}

const upgradeResp = "HTTP/1.1 101 Switching Protocols\r\n" +
	"Upgrade: " + Protocol + "\r\n" +
	"Connection: Upgrade\r\n\r\n"

// ServeHTTP 处理 HTTP 请求，从 request.Body 读取数据流
// 客户端需要发送 HTTP Upgrade 请求
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if up := req.Header.Get("Upgrade"); up == "" {
		http.Error(w, "not Upgrade request", http.StatusBadRequest)
		return
	}
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijack not supported", http.StatusInternalServerError)
		return
	}
	conn, rw, err := hj.Hijack()
	if err != nil {
		return
	}
	defer conn.Close()
	_, _ = rw.WriteString(upgradeResp)
	_ = rw.Flush()

	r.Serve(req.Context(), rw.Reader, rw.Writer)
}

func (r *Router) Serve(ctx context.Context, rd io.Reader, w io.Writer) error {
	br := bufio.NewReader(rd)
	bw := bufio.NewWriter(w)
	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		default:
		}
		reqs, batch, err := ReadRequests(br)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			err1 := r.sendError(bw, err)
			return errors.Join(err, err1)
		}
		if batch {
			err = r.serveBatch(ctx, bw, reqs)
		} else {
			err = r.serveOne(ctx, bw, reqs[0])
		}
		if err == nil {
			err = xio.TryFlush(bw, w)
		}
		if err != nil {
			return err
		}
	}
}

func (r *Router) sendError(w *bufio.Writer, err error) error {
	el := envelope{}
	if ee, ok := err.(*Error); ok {
		el.Error = ee
	} else {
		el.Error = &Error{
			Code:    xerror.ErrCode(err, -32000),
			Message: err.Error(),
		}
	}
	bf, _ := xcodec.JSON.Encode(el)
	w.Write(bf)
	w.Write([]byte("\n"))
	return w.Flush()
}

func (r *Router) oneResponse(ctx context.Context, req *Request) *envelope {
	result, err := r.Handle(ctx, req)
	if req.NoReply() {
		return nil
	}
	el := &envelope{
		Version: Version,
		ID:      idBytes(req.ID),
	}
	if err == nil {
		el.Result, err = xcodec.JSON.Encode(result)
	}

	if err != nil {
		if ee, ok := err.(*Error); ok {
			el.Error = ee
		} else {
			el.Error = &Error{
				Code:    xerror.ErrCode(err, -32000),
				Message: err.Error(),
			}
		}
	}
	return el
}

func (r *Router) serveOne(ctx context.Context, w *bufio.Writer, req *Request) error {
	el := r.oneResponse(ctx, req)
	if el == nil {
		return nil
	}
	bf, _ := xcodec.JSON.Encode(el)
	w.Write(bf)
	w.WriteByte('\n')
	return w.Flush()
}

func (r *Router) serveBatch(ctx context.Context, w *bufio.Writer, reqs []*Request) error {
	resps := make([]*envelope, 0, len(reqs))
	var mux sync.Mutex
	var wg xsync.WaitGroup
	for _, req := range reqs {
		wg.GoCtx(ctx, func(ctx context.Context) {
			if resp := r.oneResponse(ctx, req); resp != nil {
				mux.Lock()
				resps = append(resps, resp)
				mux.Unlock()
			}
		})
	}
	wg.Wait()

	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	default:
	}

	if len(resps) == 0 {
		return nil
	}

	bf, _ := xcodec.JSON.Encode(resps)
	w.Write(bf)
	w.WriteByte('\n')
	return w.Flush()
}
