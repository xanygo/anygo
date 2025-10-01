//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-01

package xredis

import "github.com/xanygo/anygo/store/xredis/resp3"

type Response struct {
	base *rpcResponse
}

func (rs Response) Result() (Result, error) {
	return rs.base.result, rs.base.err
}

func (rs Response) Err() error {
	return rs.base.err
}

func (rs Response) Int() (int, error) {
	return resp3.ToInt(rs.base.result, rs.base.err)
}

func (rs Response) String() (string, error) {
	return resp3.ToString(rs.base.result, rs.base.err)
}

type StringResponse struct {
	base *rpcResponse
}

func (rs StringResponse) Result() (Result, error) {
	return rs.base.result, rs.base.err
}

func (rs StringResponse) Err() error {
	return rs.base.err
}

func (rs StringResponse) Int() (int, error) {
	return resp3.ToInt(rs.base.result, rs.base.err)
}

func (rs StringResponse) String() (string, error) {
	return resp3.ToString(rs.base.result, rs.base.err)
}
