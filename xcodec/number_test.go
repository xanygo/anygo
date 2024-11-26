//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-26

package xcodec

import (
	"fmt"
	"testing"

	"github.com/fsgo/fst"
)

func TestInt64Cipher_Encode(t *testing.T) {
	ec := &Int64Cipher{
		Cipher: &AesOFB{
			Key: "demo",
		},
	}
	nums := []int64{0, 1, 100, 1000, 99999, 99999999}
	for _, num := range nums {
		t.Run(fmt.Sprintf("n_%d", num), func(t *testing.T) {
			str, err := ec.Encode(num)
			t.Logf("Encode(%d) = %q %v", num, str, err)
			fst.NoError(t, err)
			n, err := ec.Decode(str)
			fst.NoError(t, err)
			fst.Equal(t, num, n)
		})
	}
}
