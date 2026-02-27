//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-30

package resp3

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/xnet/xdial"
)

var _ Request = HelloRequest{}

type HelloRequest struct {
	Username   string
	Password   string
	ClientName string
	DBIndex    int // 数据库编号，可选，默认为 0
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

func (ch HelloRequest) Select() Request {
	// SELECT 0
	return NewRequest(DataTypeSimpleString, "SELECT", ch.DBIndex)
}

func (ch HelloRequest) getUsername() string {
	if ch.Username == "" {
		return "default"
	}
	return ch.Username
}

var _ xdial.SessionReply = (*HelloResponse)(nil)

type HelloResponse struct {
	ID      int64                 `json:"id"`      // 12
	Mode    string                `json:"mode"`    // standalone
	Proto   int                   `json:"proto"`   // 协议版本，3
	Role    string                `json:"role"`    //   master
	Version string                `json:"version"` // 8.2.1
	Server  string                `json:"server"`  // redis
	Modules []HelloResponseModule `json:"modules,omitempty"`
}

func (hr *HelloResponse) String() string {
	bf, _ := json.Marshal(hr)
	return string(bf)
}

func (hr *HelloResponse) Summary() string {
	return fmt.Sprintf("id=%d,mode=%s,proto=%d,role=%s,version=%s", hr.ID, hr.Mode, hr.Proto, hr.Role, hr.Version)
}

func (hr *HelloResponse) FromMap(m map[any]any) error {
	if len(m) == 0 {
		return errors.New("reply data is empty")
	}
	hr.ID, _ = xmap.GetInt64(m, "id")
	hr.Mode, _ = xmap.GetString(m, "mode")
	hr.Proto, _ = xmap.GetInt(m, "proto")
	hr.Role, _ = xmap.GetString(m, "role")
	hr.Version, _ = xmap.GetString(m, "version")
	hr.Server, _ = xmap.GetString(m, "server")
	return nil
}

type HelloResponseModule struct {
	Name string   `json:"name"` // ReJSON
	Path string   `json:"path"` // /usr/lib/redis/modules/redisbloom.so
	Ver  int64    `json:"ver"`  // 80200
	Args []string `json:"args"` //
}
