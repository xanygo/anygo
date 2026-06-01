//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-05-24

package xws

import (
	"context"
	"encoding/json"

	"github.com/xanygo/anygo/xcodec"
)

type MessageType uint8

const (
	TextMessage MessageType = 1

	BinaryMessage MessageType = 2
)

// Message 传递的消息
//
// 若是传递给对端，此时消息类型只应该是 TextMessage。
type Message struct {
	Type MessageType

	//  Bin 后面是否紧接着附加二进制消息,当消息类型是 TextMessage 时，同时 More=true，则紧接着的 BinaryMessage 消息，
	// 会继续交给前一个 TextMessage 对应的 Method 的 Handler 此处
	// 取一个更好的名字
	Bin bool

	// Method 请求方法
	Method string

	// Payload 请求的消息文本
	Payload json.RawMessage
}

// Decode 解析 Payload
func (m *Message) Decode(obj any) error {
	return xcodec.Decode(xcodec.JSON, m.Payload, obj)
}

// DecodeString 解析 Payload 为字符串
func (m *Message) DecodeString() (string, error) {
	var str string
	err := m.Decode(&str)
	return str, err
}

func (m *Message) WithPayload(obj any) error {
	bf, err := xcodec.Encode(xcodec.JSON, obj)
	m.Payload = bf
	return err
}

type Request struct {
	Conn    ConnMeta
	Message *Message
}

type ResponseWriter interface {
	Write(context.Context, *Message) error
	Conn() Conn
}

type Handler interface {
	ServeWS(ctx context.Context, w ResponseWriter, req *Request)
}

type HandlerFunc func(ctx context.Context, w ResponseWriter, r *Request)

func (hf HandlerFunc) ServeWS(ctx context.Context, w ResponseWriter, r *Request) {
	hf(ctx, w, r)
}

type Middleware func(next HandlerFunc) HandlerFunc

type EventHandler func(Conn)

var _ Handler = (*Router)(nil)

// Router 路由
type Router struct {
	textHandlers map[string]Handler

	binaryHandler Handler

	middlewares []Middleware

	notFoundHandler Handler
}

func NewRouter() *Router {
	r := &Router{
		textHandlers: make(map[string]Handler),
	}
	return r
}

func (r *Router) doNotFound(ctx context.Context, w ResponseWriter, req *Request) {
	if r.notFoundHandler != nil {
		r.notFoundHandler.ServeWS(ctx, w, req)
		return
	}
	msg := &Message{
		Type:   TextMessage,
		Method: "error.notFound",
	}
	w.Write(ctx, msg)
}

func (r *Router) ServeWS(ctx context.Context, w ResponseWriter, req *Request) {
	msg := req.Message
	if msg.Type == BinaryMessage && msg.Method == "" {
		if r.binaryHandler != nil {
			r.binaryHandler.ServeWS(ctx, w, req)
		} else {
			r.doNotFound(ctx, w, req)
		}
		return
	}
	h, ok := r.textHandlers[msg.Method]
	if !ok {
		r.doNotFound(ctx, w, req)
		return
	}
	h.ServeWS(ctx, w, req)
}

func (r *Router) Use(mw Middleware) {
	r.middlewares = append(r.middlewares, mw)
}

func (r *Router) chain(h HandlerFunc) HandlerFunc {
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		h = r.middlewares[i](h)
	}
	return h
}

func (r *Router) HandleText(method string, h Handler) {
	r.HandleTextFunc(method, h.ServeWS)
}

func (r *Router) HandleTextFunc(method string, h HandlerFunc) {
	if r.textHandlers == nil {
		r.textHandlers = make(map[string]Handler)
	}
	r.textHandlers[method] = r.chain(h)
}

func (r *Router) HandleBinaryFunc(h HandlerFunc) {
	r.binaryHandler = h
}

func (r *Router) HandleBinary(h Handler) {
	r.binaryHandler = h
}

// HandleNotFound 设置 NotFound Handler
func (r *Router) HandleNotFound(h Handler) {
	r.notFoundHandler = h
}

func (r *Router) HandleNotFoundFunc(h HandlerFunc) {
	r.notFoundHandler = h
}
