//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-26

package xcodec_test

import (
	"fmt"

	"github.com/xanygo/anygo/xcodec"
)

func ExampleAesBlock_Encrypt() {
	ac := &xcodec.AesBlock{
		Key: "demo",
	}
	str1, _ := ac.Encrypt([]byte("hello"))
	fmt.Printf("Encrypt= %q\n", str1)

	str2, _ := ac.Decrypt(str1)
	fmt.Printf("Decrypt= %q\n", str2)

	// Output:
	// Encrypt= "\xc8\xc4?\xa2\xf3\x00Ͳ\xc1~\xb1\xb7\x96\xe3\xe4\x82"
	// Decrypt= "hello"
}

func ExampleAesOFB_Encrypt() {
	ac := &xcodec.AesOFB{
		Key: "demo",
	}
	str1, _ := ac.Encrypt([]byte("hello"))
	fmt.Printf("Encrypt= %q\n", str1)

	str2, _ := ac.Decrypt(str1)
	fmt.Printf("Decrypt= %q\n", str2)

	// Output:
	// Encrypt= "2\xa0\x1c\x90\xb4"
	// Decrypt= "hello"
}

func ExampleInt64Cipher_Encode() {
	ac := &xcodec.Int64Cipher{
		Cipher: &xcodec.AesOFB{
			Key: "demo",
		},
	}
	nums := []int64{0, 1, 1000, 10000, 99999999}
	for _, num := range nums {
		str1, _ := ac.Encode(num)
		fmt.Printf("Encode(%d) = %q\n", num, str1)

		num1, _ := ac.Decode(str1)
		fmt.Printf("Decode(%q) = %d\n\n", str1, num1)
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
