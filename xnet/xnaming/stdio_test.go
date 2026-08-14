//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-04-01

package xnaming

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestCommand_Lookup(t *testing.T) {
	si := &Stdio{}
	items, err := si.Lookup(t.Context(), "test", `{"Path":"echo"}`)
	xt.NoError(t, err)
	xt.Len(t, items, 1)

	items, err = si.Lookup(t.Context(), "test", `{"Path":"echo","Args":["a"]}`)
	xt.NoError(t, err)
	xt.Len(t, items, 1)

	items, err = si.Lookup(t.Context(), "test", `{"Args":["a"]}`)
	xt.Error(t, err)
	xt.Len(t, items, 0)
}
