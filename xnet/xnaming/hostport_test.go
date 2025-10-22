//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-11

package xnaming

import (
	"context"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestHostPort_Lookup(t *testing.T) {
	hp := &HostPort{}
	nodes1, err1 := hp.Lookup(context.Background(), "bj", "example.com:80", nil)
	xt.NoError(t, err1)
	xt.Len(t, nodes1, 1)
	testNodesEqual(t, nodes1, []string{"example.com:80"})

	nodes2, err2 := hp.Lookup(context.Background(), "bj", "example.com", nil)
	xt.Error(t, err2)
	xt.Empty(t, nodes2)
}
