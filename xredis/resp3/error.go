//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-16

package resp3

import (
	"bytes"
	"errors"
	"strconv"
)

var (
	ErrNil          = errors.New("redis null reply") // redis 中无此 key
	ErrInvalidReply = errors.New("invalid reply")    // 错误的响应
)

// RespError redis server 返回的错误类型（Element）
type RespError interface {
	RespError() bool
}

func IsRespError(err error) bool {
	if err == nil {
		return false
	}
	var re RespError
	return errors.As(err, &re)
}

var _ Element = SimpleError("")

// SimpleError RESP has specific data types for errors.
// Simple errors, or simply just errors, are similar to simple strings,
// but their first character is the minus (-) character.
// The difference between simple strings and errors in RESP is that clients should treat errors as exceptions,
// whereas the string encoded in the error type is the error message itself.
type SimpleError string

func (s SimpleError) Bytes(bf *bytes.Buffer) []byte {
	// -Error message\r\n
	bf.Reset()
	bf.WriteByte(DataTypeSimpleError.Byte())
	bf.WriteString(string(s))
	bf.Write(CRLF)
	return bf.Bytes()
}

func (s SimpleError) Error() string {
	return string(s)
}

var _ RespError = SimpleError("")

func (s SimpleError) RespError() bool {
	return true
}

func (s SimpleError) DataType() DataType {
	return DataTypeSimpleError
}

var _ Element = BulkError("")

type BulkError string

func (be BulkError) Bytes(bf *bytes.Buffer) []byte {
	// !21\r\n
	// SYNTAX invalid syntax\r\n
	bf.Reset()
	bf.WriteByte(DataTypeBulkError.Byte())
	bf.WriteString(strconv.Itoa(len(string(be))))
	bf.Write(CRLF)
	bf.WriteString(string(be))
	bf.Write(CRLF)
	return bf.Bytes()
}

func (be BulkError) DataType() DataType {
	return DataTypeBulkError
}

func (be BulkError) Error() string {
	return string(be)
}

var _ RespError = SimpleError("")

func (be BulkError) RespError() bool {
	return true
}
