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

type Response struct {
	ID     ID
	Error  *Error
	Result json.RawMessage
}

func (res *Response) envelope() envelope {
	return envelope{
		Version: Version,
		ID:      idBytes(res.ID),
		Error:   res.Error,
		Result:  res.Result,
	}
}

func (res *Response) DecodeResult(obj any) error {
	return xcodec.JSON.Decode(res.Result, obj)
}

func (res *Response) WIthResult(obj any) error {
	bf, err := xcodec.JSON.Encode(obj)
	res.Result = bf
	return err
}

var _ io.WriterTo = (*Response)(nil)

func (res *Response) WriteTo(w io.Writer) (int64, error) {
	bf, err := xcodec.JSON.Encode(res.envelope())
	if err != nil {
		return 0, err
	}
	bf = append(bf, '\n')
	num, err := w.Write(bf)
	return int64(num), err
}

func ReadResponse(rd xio.SliceReader) (*Response, error) {
	return readResponse(rd)
}

func readResponse(rd xio.SliceReader) (*Response, error) {
	bf, err := rd.ReadSlice('\n')
	if err != nil {
		return nil, err
	}
	return parserResponse(bf)
}

func parserResponse(bf []byte) (*Response, error) {
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
	return &Response{
		ID:     id,
		Error:  el.Error,
		Result: el.Result,
	}, nil
}

// ReadResponses 读取响应信息
//
// 返回值：响应列表，是否批量，错误
//
// 若不是批量请求，则返回的 []*Response 个数总是 1
func ReadResponses(rd *bufio.Reader) ([]*Response, bool, error) {
	head, err := rd.Peek(1)
	if err != nil {
		return nil, false, err
	}
	if head[0] != '[' {
		res, err := ReadResponse(rd)
		if err != nil {
			return nil, false, err
		}
		return []*Response{res}, false, nil
	}
	var bf bytes.Buffer
	for {
		line, err := rd.ReadSlice('\n')
		if err != nil {
			return nil, false, err
		}
		line = bytes.TrimSpace(line)
		bf.Write(line)
		if bytes.HasSuffix(line, []byte("]")) {
			break
		}
	}
	var batch []json.RawMessage
	err = xcodec.JSON.Decode(bf.Bytes(), &batch)
	if err != nil {
		return nil, true, err
	}

	result := make([]*Response, len(batch))

	for i, b := range batch {
		req, err := parserResponse(b)
		if err != nil {
			return nil, true, errors.Join(err, ErrParse)
		}
		result[i] = req
	}
	return result, true, nil
}
