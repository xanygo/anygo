//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-03-30

package xjsonrpc2_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/xanygo/anygo/xnet/xjsonrpc2"
	"github.com/xanygo/anygo/xt"
)

func TestRouter_Handle(t *testing.T) {
	router := &xjsonrpc2.Router{}
	router.Register("hello", xjsonrpc2.HandlerFunc(func(ctx context.Context, req *xjsonrpc2.Request) (result any, err error) {
		return "world", nil
	}))
	t.Run("case 1", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			input := `{"jsonrpc": "2.0", "method": "hello", "params": [1,2,4], "id": "1"}
{"jsonrpc": "2.0", "method": "hello", "params": [1,2,4], "id": "2"}
`
			bf := strings.NewReader(input)
			w := &bytes.Buffer{}
			err := router.Serve(context.Background(), bf, w)
			xt.NoError(t, err)
			want := `{"jsonrpc":"2.0","id":"1","result":"world"}
{"jsonrpc":"2.0","id":"2","result":"world"}
`
			xt.Equal(t, want, w.String())
		}
	})
}
