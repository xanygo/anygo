//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-06

package xsession

import (
	"testing"

	"github.com/fsgo/fst"
)

func TestSet(t *testing.T) {
	s := &Session{}
	fst.NoError(t, Set(s, "k1", "v1"))
	got1, err1 := Load[string](s, "k1")
	fst.NoError(t, err1)
	fst.Equal(t, "v1", got1)
}
