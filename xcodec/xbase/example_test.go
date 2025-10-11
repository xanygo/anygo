//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-27

package xbase_test

import (
	"fmt"

	"github.com/xanygo/anygo/xcodec/xbase"
)

func ExampleEncoding_EncodeInt64() {
	nums := []int64{0, 1, 100, 1000, 99999, 99999999}
	for _, n := range nums {
		fmt.Printf("Base62.EncodeInt64(%d)= %s\n", n, xbase.Base62.EncodeInt64(n))
		fmt.Printf("Base36.EncodeInt64(%d)= %s\n\n", n, xbase.Base36.EncodeInt64(n))
	}

	// Output:
	// Base62.EncodeInt64(0)= 0
	// Base36.EncodeInt64(0)= 0
	//
	// Base62.EncodeInt64(1)= 1
	// Base36.EncodeInt64(1)= 1
	//
	// Base62.EncodeInt64(100)= c1
	// Base36.EncodeInt64(100)= s2
	//
	// Base62.EncodeInt64(1000)= 8G
	// Base36.EncodeInt64(1000)= sr
	//
	// Base62.EncodeInt64(99999)= t0Q
	// Base36.EncodeInt64(99999)= r552
	//
	// Base62.EncodeInt64(99999999)= DZal6
	// Base36.EncodeInt64(99999999)= rhcjn1
}
