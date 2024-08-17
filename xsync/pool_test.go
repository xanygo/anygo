//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-17

package xsync

import (
	"bytes"
	"testing"

	"github.com/fsgo/fst"
)

func TestPool(t *testing.T) {
	p := &Pool[*bytes.Buffer]{
		New: func() *bytes.Buffer {
			return &bytes.Buffer{}
		},
	}
	g1 := p.Get()
	fst.NotNil(t, g1)
	p.Put(g1)
}
