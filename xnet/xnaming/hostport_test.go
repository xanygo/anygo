//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-11

package xnaming

import (
	"context"
	"testing"

	"github.com/fsgo/fst"
)

func TestHostPort_Lookup(t *testing.T) {
	hp := &HostPort{}
	nodes1, err1 := hp.Lookup(context.Background(), "bj", "example.com:80", nil)
	fst.NoError(t, err1)
	fst.Len(t, nodes1, 1)
	testNodesEqual(t, nodes1, []string{"example.com:80"})

	nodes2, err2 := hp.Lookup(context.Background(), "bj", "example.com", nil)
	fst.Error(t, err2)
	fst.Empty(t, nodes2)
}
