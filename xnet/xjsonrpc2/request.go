//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-03-28

package xjsonrpc2

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xio"
)

const Version = "2.0"
const Protocol = "JSON-RPC2"

func NewRequest(id ID, method string, params any) (*Request, error) {
	req := &Request{
		ID:     id,
		Method: method,
	}
	err := req.WithParams(params)
	return req, err
}

type Request struct {
	// ID 客户端的唯一标识id，值必须包含一个字符串、数值或 NULL 空值
	ID ID

	// Method 方法名
	Method string

	// Params 请求参数
	Params json.RawMessage
}

func (req *Request) envelope() envelope {
	return envelope{
		Version: Version,
		ID:      idBytes(req.ID),
		Method:  req.Method,
		Params:  req.Params,
	}
}

type noReply interface {
	NoReply() bool
}

var _ noReply = (*Request)(nil)

// NoReply 是否是通知/不需要回复
func (req *Request) NoReply() bool {
	return req.ID == nil
}

var _ io.WriterTo = (*Request)(nil)

func (req *Request) WriteTo(w io.Writer) (int64, error) {
	bf, err := xcodec.JSON.Encode(req.envelope())
	if err != nil {
		return 0, err
	}
	bf = append(bf, '\n')
	num, err := w.Write(bf)
	if err != nil {
		return int64(num), err
	}
	return int64(num), xio.TryFlush(w)
}

func (req *Request) Write(w io.Writer) error {
	_, err := req.WriteTo(w)
	return err
}

func (req *Request) DecodeParams(obj any) error {
	err := xcodec.JSON.Decode(req.Params, obj)
	if err == nil {
		return nil
	}
	return errors.Join(ErrInvalidParams, err)
}

func (req *Request) WithParams(obj any) error {
	if obj == nil {
		return nil
	}
	bf, err := xcodec.JSON.Encode(obj)
	req.Params = bf
	return err
}

// envelope 包含 Request 和 Response 所有字段 用于序列化
type envelope struct {
	Version string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Error   *Error          `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

func ReadRequest(rd xio.SliceReader) (*Request, error) {
	req, err := readRequest(rd)
	if err != nil {
		return nil, errors.Join(ErrParse, err)
	}
	return req, nil
}

func readRequest(rd xio.SliceReader) (*Request, error) {
	bf, err := rd.ReadSlice('\n')
	if err != nil {
		return nil, err
	}
	return parserRequest(bf)
}

func parserRequest(bf []byte) (*Request, error) {
	el := &envelope{}
	err := xcodec.JSON.Decode(bf, &el)
	if err != nil {
		return nil, err
	}
	if el.Version != Version {
		return nil, fmt.Errorf("invalid version %q", el.Version)
	}
	id, err := parserID(el.ID)
	if err != nil {
		return nil, err
	}
	return &Request{
		ID:     id,
		Method: el.Method,
		Params: el.Params,
	}, nil
}

// ReadRequests 读取请求信息
//
// 返回值：请求列表，是否批量，错误
//
// 若不是批量请求，则返回的 []*Request 个数总是 1
func ReadRequests(rd *bufio.Reader) ([]*Request, bool, error) {
	head, err := rd.Peek(1)
	if err != nil {
		return nil, false, err
	}

	if head[0] != '[' {
		req, err := ReadRequest(rd)
		if err != nil {
			return nil, false, err
		}
		return []*Request{req}, false, nil
	}
	var bf bytes.Buffer
	for {
		line, err := rd.ReadSlice('\n')
		if err != nil {
			return nil, false, err
		}
		bf.Write(line)
		line = bytes.TrimSpace(line)
		if bytes.HasSuffix(line, []byte("]")) {
			break
		}
	}
	var batch []json.RawMessage
	err = xcodec.JSON.Decode(bf.Bytes(), &batch)
	if err != nil {
		return nil, true, err
	}

	result := make([]*Request, len(batch))

	for i, b := range batch {
		req, err := parserRequest(b)
		if err != nil {
			return nil, true, errors.Join(err, ErrParse)
		}
		result[i] = req
	}
	return result, true, nil
}
