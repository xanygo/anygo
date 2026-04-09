//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-03-28

package xjsonrpc2

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"maps"
	"net/http"
	"sync"

	"github.com/xanygo/anygo/ds/xctx"
	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xio"
)

type ResponseWriter interface {
	Write(resp *Response) error
}

type Handler interface {
	Handle(ctx context.Context, w ResponseWriter, req *Request) error
}

type HandlerFunc func(ctx context.Context, w ResponseWriter, req *Request) error

func (f HandlerFunc) Handle(ctx context.Context, w ResponseWriter, req *Request) error {
	return f(ctx, w, req)
}

type UnaryHandler interface {
	HandleUnary(ctx context.Context, req *Request) (result any, err error)
}

type UnaryHandlerFunc func(ctx context.Context, req *Request) (result any, err error)

func (h UnaryHandlerFunc) HandleUnary(ctx context.Context, req *Request) (result any, err error) {
	return h(ctx, req)
}

func (h UnaryHandlerFunc) Handle(ctx context.Context, w ResponseWriter, req *Request) error {
	data, err := h(ctx, req)
	if req.NoReply() {
		return err
	}
	resp, err1 := NewResponse(req.ID, data, err)
	if err1 != nil {
		return err1
	}
	return w.Write(resp)
}

var _ ResponseWriter = (*responseWriterImpl)(nil)

type responseWriterImpl struct {
	w io.Writer
}

func (rw *responseWriterImpl) Write(resp *Response) error {
	_, err := resp.WriteTo(rw.w)
	return err
}

func NotFound(ctx context.Context, w ResponseWriter, req *Request) error {
	if req.NoReply() {
		return nil
	}
	resp, err := NewResponse(req.ID, nil, ErrMethodNotFound)
	if err != nil {
		return err
	}
	return w.Write(resp)
}

var _ http.Handler = (*Router)(nil)
var _ Handler = (*Router)(nil)

func NewRouter() *Router {
	return &Router{
		handlers: make(map[string]Handler),
		notFound: HandlerFunc(NotFound),
	}
}

type Router struct {
	handlers map[string]Handler
	notFound Handler
}

func (r *Router) Register(method string, h Handler) {
	if r.handlers == nil {
		r.handlers = make(map[string]Handler)
	}
	r.handlers[method] = h
}

func (r *Router) Clone() *Router {
	return &Router{
		handlers: maps.Clone(r.handlers),
		notFound: r.notFound,
	}
}

func (r *Router) RegisterUnary(method string, fn func(ctx context.Context, req *Request) (result any, err error)) {
	r.Register(method, UnaryHandlerFunc(fn))
}

func (r *Router) Handle(ctx context.Context, w ResponseWriter, req *Request) error {
	if len(r.handlers) == 0 {
		return r.handleNotFound(ctx, w, req)
	}
	h, ok := r.handlers[req.Method]
	if !ok {
		return r.handleNotFound(ctx, w, req)
	}
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()
	return h.Handle(ctx, w, req)
}

func (r *Router) handleNotFound(ctx context.Context, w ResponseWriter, req *Request) error {
	if r.notFound != nil {
		return r.notFound.Handle(ctx, w, req)
	}
	return NotFound(ctx, w, req)
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
	ctx := xctx.WithClientConn(req.Context(), conn)

	r.Serve(ctx, rw.Reader, rw.Writer)
}

func (r *Router) Serve(ctx context.Context, rd io.Reader, w io.Writer) error {
	br := bufio.NewReader(rd)
	lbw := &xio.LockedWriter[*bufio.Writer]{
		Writer: bufio.NewWriter(w),
	}
	var wg xsync.WaitGroup
	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		default:
		}
		reqs, batch, err := ReadRequests(br)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			err1 := r.sendError(lbw, err)
			return errors.Join(err, err1)
		}
		wg.GoCtx(ctx, func(ctx context.Context) {
			if batch {
				_ = r.serveBatch(ctx, lbw, reqs)
			} else {
				_ = r.serveOne(ctx, lbw, reqs[0])
			}
		})
	}
	return wg.Wait()
}

type lockedBW = xio.LockedWriter[*bufio.Writer]

func (r *Router) sendError(w *lockedBW, err error) error {
	resp, e := NewResponse(nil, nil, err)
	if e != nil {
		return e
	}
	err2 := w.WithLock(func(w *bufio.Writer) error {
		if err1 := resp.Write(w); err1 != nil {
			return err1
		}
		return w.Flush()
	})
	return err2
}

func (r *Router) serveOne(ctx context.Context, w *lockedBW, req *Request) error {
	ww := &responseWriterImpl{
		w: w,
	}
	err := r.Handle(ctx, ww, req)
	if err != nil {
		return err
	}
	return w.WithLock(func(w *bufio.Writer) error {
		return w.Flush()
	})
}

func (r *Router) serveBatch(ctx context.Context, w *lockedBW, reqs []*Request) error {
	resps := make([]json.RawMessage, 0, len(reqs))
	var mux sync.Mutex
	var wg xsync.WaitGroup
	for _, req := range reqs {
		wg.GoCtx(ctx, func(ctx context.Context) {
			bf := bytes.NewBuffer(nil)
			ww := &responseWriterImpl{
				w: bf,
			}
			r.Handle(ctx, ww, req)
			if bf.Len() > 0 {
				mux.Lock()
				resps = append(resps, bf.Bytes())
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
	bf = append(bf, '\n')
	return w.WithLock(func(w *bufio.Writer) error {
		w.Write(bf)
		return w.Flush()
	})
}
