//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-12

package xservice

import (
	"testing"

	"github.com/fsgo/fst"
)

func TestDummyService(t *testing.T) {
	ser := DummyService()
	fst.NotEmpty(t, ser)
}
