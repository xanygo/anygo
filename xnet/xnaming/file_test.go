//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-11

package xnaming

import (
	"context"
	"testing"

	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xt"
)

func TestFileStore_Lookup(t *testing.T) {
	f := &FileStore{}
	nodes1, err1 := f.Lookup(context.Background(), "bj", "testdata/file/server_list_0.txt", nil)
	xt.NoError(t, err1)
	testNodesEqual(t, nodes1, []string{"127.0.0.1:8000", "127.0.0.2:8000", "10.0.0.1:9000"})
}

func testNodesEqual(t *testing.T, nodes []xnet.AddrNode, want []string) {
	xt.Len(t, nodes, len(want))
	var addrs []string
	for _, node := range nodes {
		addrs = append(addrs, node.Addr.String())
	}
	xt.Equal(t, want, addrs)
}
