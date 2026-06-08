//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-06-01

package xws

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/safely"
	"github.com/xanygo/anygo/xerror"
)

type ConnID int64

type ConnMeta interface {
	ID() ConnID

	HTTPRequest() *http.Request
}

type Conn interface {
	ConnMeta

	Write(context.Context, *Message) error

	Close() error
}

func newConn(id ConnID, sendBuf int, hr *http.Request) *wsConn {
	return &wsConn{
		id:   id,
		hr:   hr,
		send: make(chan *Message, sendBuf),
		quit: make(chan struct{}),
	}
}

var _ Conn = (*wsConn)(nil)

type wsConn struct {
	id     ConnID
	hr     *http.Request
	send   chan *Message
	quit   chan struct{}
	closed atomic.Bool
}

func (c *wsConn) HTTPRequest() *http.Request {
	return c.hr
}

func (c *wsConn) ID() ConnID {
	return c.id
}

func (c *wsConn) Write(ctx context.Context, msg *Message) error {
	if c.closed.Load() {
		return xerror.Closed
	}

	select {
	case c.send <- msg:
		return nil
	case <-ctx.Done():
		return context.Cause(ctx)
	case <-c.quit:
		return xerror.Closed
	}
}

func (c *wsConn) loopWrite(ctx context.Context, writeFn func(*Message) error, retryFn func(*Message)) {
	for {
		select {
		case msg := <-c.send:
			err := writeFn(msg)
			if err != nil {
				if retryFn != nil {
					retryFn(msg)
				}
			}

		case <-c.quit:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (c *wsConn) Close() error {
	if c.closed.CompareAndSwap(false, true) {
		close(c.quit)
		close(c.send)
	}
	return nil
}

type Filter func(Conn) bool

func FilterAnd(filters ...Filter) Filter {
	return func(c Conn) bool {
		if len(filters) == 0 {
			return false
		}
		for _, f := range filters {
			if !f(c) {
				return false
			}
		}
		return true
	}
}

func FilterOr(filters ...Filter) Filter {
	return func(c Conn) bool {
		if len(filters) == 0 {
			return false
		}
		for _, f := range filters {
			if f(c) {
				return true
			}
		}
		return false
	}
}

var _ ResponseWriter = (*responseWriter)(nil)

type responseWriter struct {
	conn Conn
}

func (w *responseWriter) Conn() Conn {
	return w.conn
}

func (w *responseWriter) Write(ctx context.Context, msg *Message) error {
	return w.conn.Write(ctx, msg)
}

// Hub 管理 websocket 连接
type Hub struct {
	conns xmap.Sync[ConnID, Conn]

	onConnect    []func(context.Context, Conn) error
	onDisconnect []func(context.Context, Conn)
}

func (h *Hub) Add(conn Conn) {
	h.conns.Store(conn.ID(), conn)
}

func (h *Hub) Remove(id ConnID) {
	h.conns.Delete(id)
}

func (h *Hub) Get(id ConnID) (Conn, bool) {
	return h.conns.Load(id)
}

func (h *Hub) Exists(id ConnID) bool {
	return h.conns.Exists(id)
}

func (h *Hub) Count() int {
	return h.conns.Len()
}

func (h *Hub) Range(fn func(id ConnID, conn Conn) bool) {
	h.conns.Range(fn)
}

var ErrConnNotFound = errors.New("connection not found")

func (h *Hub) Send(ctx context.Context, id ConnID, msg *Message) error {
	conn, ok := h.Get(id)
	if !ok {
		return ErrConnNotFound
	}
	return conn.Write(ctx, msg)
}

// Broadcast 广播，返回值为满足 filter 筛选条件，并且已进入连接的异步发送队列的数量
func (h *Hub) Broadcast(ctx context.Context, filter Filter, msg *Message) int64 {
	var num int64
	h.Range(func(id ConnID, conn Conn) bool {
		if filter(conn) {
			return true
		}

		if err := conn.Write(ctx, msg); err != nil {
			num++
		}
		return true
	})
	return num
}

// OnConnect 注册连接创建时候的回调函数，若期望阻塞后续逻辑，回调函数应返回 error!=nil，
//
//	如注册了 fn1，fn2，fn3, fn4, 若 fn2 返回 io.EOF ,则后续逻辑（包括 fn3,fn4）不再执行
func (h *Hub) OnConnect(fn func(context.Context, Conn) error) {
	h.onConnect = append(h.onConnect, fn)
}

// OnDisconnect 注册连接断开后的回调函数
func (h *Hub) OnDisconnect(fn func(context.Context, Conn)) {
	h.onDisconnect = append(h.onDisconnect, fn)
}

func (h *Hub) fireConnect(ctx context.Context, w Conn) error {
	for _, fn := range h.onConnect {
		if err := fn(ctx, w); err != nil {
			return err
		}
	}
	return nil
}

func (h *Hub) fireDisconnect(ctx context.Context, w Conn) {
	for _, fn := range h.onDisconnect {
		fn(ctx, w)
	}
}

func (h *Hub) CloseAll() {
	h.Range(func(id ConnID, conn Conn) bool {
		_ = conn.Close()
		h.conns.Delete(id)
		return true
	})
}

var nextConnID atomic.Int64

// ServeWSUpgradeRaw 使用 Handler/Router 接收处理 ws 连接请求
//
// import "github.com/gorilla/websocket"
//
//	var hub = &xws.Hub{}
//	var wsRouter = xws.NewRouter()
//	var upgrader = websocket.Upgrader{}
//
//	func wsHandler(w http.ResponseWriter, r *http.Request) {
//		c, err := upgrader.Upgrade(w, r, nil)
//		if err == nil {
//			err = hub.ServeWSUpgradeRaw(r.Context(), wsRouter, r, c)
//		}
//		log.Println("wsHandler, err=", err)
//	}
//
// 更多内容详见  ServeWSUpgrade
func (h *Hub) ServeWSUpgradeRaw(ctx context.Context, handler Handler, hr *http.Request, rw RawReadWriter) error {
	rw1 := &readWriter{
		raw: rw,
	}
	return h.ServeWSUpgrade(ctx, handler, hr, rw1)
}

// ServeWSUpgrade  使用 Handler/Router 接收处理 ws 连接业务数据请求（MessageType 必须是 TextMessage 或 BinaryMessage）
//
// Ping-Pong 等其他类型的数据不予处理(若传入，会报错)
// 此方法是同步的，可以处理一个 http upgrade 之后的 websocket 连接，在一开始，会先触发 Connect 回调，在读写异常等情况后，会触发 Disconnect 回调
func (h *Hub) ServeWSUpgrade(ctx context.Context, handler Handler, hr *http.Request, rw ReadWriter) error {
	id := ConnID(nextConnID.Add(1))
	conn := newConn(id, 8, hr)
	defer conn.Close()
	go safely.Run(func() {
		conn.loopWrite(ctx, func(m *Message) error {
			return m.WriteTo(rw)
		}, nil)

		if wc, ok := rw.(io.Closer); ok {
			wc.Close()
		}
	})
	w := &responseWriter{conn: conn}
	if err := h.fireConnect(ctx, conn); err != nil {
		return err
	}
	h.Add(conn)
	defer func() {
		h.Remove(id)
		h.fireDisconnect(ctx, conn)
	}()
	var moreBin bool
	var method string
	for {
		mt, payload, err := rw.ReadMessage()
		if err != nil {
			return err
		}
		var msg *Message
		// 当第一个请求时 Text 类型的时候，可以在请求里设置 More=true，并紧接着发送 Binary 消息
		// 这时候，此二进制消息会继续发送给之前的 Handler 去处理
		if mt == BinaryMessage {
			if moreBin {
				msg = &Message{
					Type:    BinaryMessage,
					Payload: payload,
					Bin:     true,
					Method:  method,
				}
			} else {
				msg = &Message{
					Type:    BinaryMessage,
					Payload: payload,
				}
				method = ""
			}
		} else {
			msg, err = decode(mt, payload)
			if err != nil {
				return err
			}
			method = msg.Method
			moreBin = msg.Bin
		}

		req := &Request{
			Conn:    conn,
			Message: msg,
		}
		handler.ServeWS(ctx, w, req)
	}
}

type ReadWriter interface {
	Reader
	Writer
}

type Reader interface {
	ReadMessage() (MessageType, []byte, error)
}

type Writer interface {
	WriteMessage(MessageType, []byte) error
}

type RawReadWriter interface {
	ReadMessage() (int, []byte, error)
	WriteMessage(int, []byte) error
}

var _ ReadWriter = (*readWriter)(nil)

type readWriter struct {
	raw RawReadWriter
}

func (r *readWriter) ReadMessage() (MessageType, []byte, error) {
	mt, p, err := r.raw.ReadMessage()
	if err != nil {
		return MessageType(mt), p, err
	}
	tp := MessageType(mt)
	switch tp {
	case TextMessage, BinaryMessage:
		return tp, p, err
	default:
		return tp, p, fmt.Errorf("invalid message type %d", mt)
	}
}

func (r *readWriter) WriteMessage(mt MessageType, bytes []byte) error {
	return r.raw.WriteMessage(int(mt), bytes)
}
