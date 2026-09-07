//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-26

package xcipher_test

import (
	"fmt"
	"testing"

	"github.com/xanygo/anygo/xenc/xcipher"
	"github.com/xanygo/anygo/xt"
)

func TestInt64Cipher_Encode(t *testing.T) {
	ec := &xcipher.Int64{
		Cipher: &xcipher.AesOFB{
			Key: "demo",
		},
	}
	nums := []int64{0, 1, 100, 1000, 99999, 99999999}
	for _, num := range nums {
		t.Run(fmt.Sprintf("n_%d", num), func(t *testing.T) {
			str, err := ec.EncodeInt(num)
			t.Logf("Encode(%d) = %q %v", num, str, err)
			xt.NoError(t, err)

			n, err := ec.DecodeInt(str)
			xt.NoError(t, err)
			xt.Equal(t, n, num)
		})
	}
}
