//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-25

package anygo

import (
	"testing"

	"github.com/fsgo/fst"
)

func TestTernary(t *testing.T) {
	fst.Equal(t, 1, Ternary(true, 1, 2))
	fst.Equal(t, 2, Ternary(false, 1, 2))
}
