//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-01

package session

import (
	"testing"

	"github.com/fsgo/fst"
)

func TestNewID(t *testing.T) {
	got := NewID()
	fst.Len(t, got, 32)
}
