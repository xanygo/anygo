//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-11

package xnaming

import (
	"context"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestLookup(t *testing.T) {
	nodes1, err1 := Lookup(context.Background(), "", "bj", "example.com:80", nil)
	xt.NoError(t, err1)
	testNodesEqual(t, nodes1, []string{"example.com:80"})
}
