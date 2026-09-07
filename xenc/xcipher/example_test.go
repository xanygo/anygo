package xcipher_test

import (
	"fmt"

	"github.com/xanygo/anygo/xenc/xcipher"
)

func ExampleInt64_EncodeInt() {
	ac := &xcipher.Int64{
		Cipher: &xcipher.AesOFB{
			Key: "demo",
		},
	}
	nums := []int64{0, 1, 1000, 10000, 99999999}
	for _, num := range nums {
		str, _ := ac.EncodeInt(num)
		fmt.Printf("Encode(%d) = %q\n", num, str)

		num1, _ := ac.DecodeInt(str)
		fmt.Printf("Decode(%q) = %d\n\n", str, num1)
	}

	// Output:
	// Encode(0) = "i1"
	// Decode("i1") = 0
	//
	// Encode(1) = "j1"
	// Decode("j1") = 1
	//
	// Encode(1000) = "kY6"
	// Decode("kY6") = 1000
	//
	// Encode(10000) = "E4P5"
	// Decode("E4P5") = 10000
	//
	// Encode(99999999) = "v24ZYJ2"
	// Decode("v24ZYJ2") = 99999999
}
