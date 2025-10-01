//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-30

package resp3

import "bytes"

var _ Request = HelloRequest{}

type HelloRequest struct {
	Username   string
	Password   string
	ClientName string
}

func (ch HelloRequest) ResponseType() DataType {
	return DataTypeMap
}

func (ch HelloRequest) Name() string {
	return "hello"
}

func (ch HelloRequest) Args() []any {
	return nil
}

func (ch HelloRequest) Bytes(bf *bytes.Buffer) []byte {
	// HELLO [protover [AUTH username password] [SETNAME clientname]]
	args := []any{"hello", "3", "AUTH", ch.getUsername(), ch.Password}
	if ch.ClientName != "" {
		args = append(args, "SETNAME", ch.ClientName)
	}
	return NewRequest(ch.ResponseType(), args...).Bytes(bf)
}

func (ch HelloRequest) getUsername() string {
	if ch.Username == "" {
		return "default"
	}
	return ch.Username
}
