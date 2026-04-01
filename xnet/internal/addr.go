//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-04-01

package internal

import (
	"encoding"
	"encoding/json"
	"net"
)

func NewAddr(network, address string) *Addr {
	return &Addr{
		network: network,
		address: address,
	}
}

var _ net.Addr = (*Addr)(nil)
var _ encoding.TextMarshaler = (*Addr)(nil)
var _ json.Marshaler = (*Addr)(nil)

type Addr struct {
	network string
	address string
}

func (a *Addr) MarshalJSON() ([]byte, error) {
	data := map[string]any{
		"Network": a.network,
		"Address": a.address,
	}
	return json.Marshal(data)
}

func (a *Addr) MarshalText() ([]byte, error) {
	return a.MarshalJSON()
}

func (a *Addr) Network() string {
	return a.network
}

func (a *Addr) String() string {
	return a.address
}
