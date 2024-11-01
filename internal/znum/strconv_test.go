//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-01

package znum

import (
	"math"
	"strconv"
	"testing"

	"github.com/fsgo/fst"
)

func TestFormatIntB62(t *testing.T) {
	numbers := []int64{-1, 0, -999, 999, 9999, 100000, math.MaxInt64}
	for _, num := range numbers {
		t.Run(strconv.FormatInt(num, 10), func(t *testing.T) {
			str := FormatIntB62(num)
			t.Logf("num=%d b62=%q", num, str)
			got, err := ParserIntB62(str)
			fst.NoError(t, err)
			fst.Equal(t, num, got)
		})
	}
	fst.Equal(t, "7m85Y0n8LzA", FormatIntB62(math.MaxInt64))
}
