//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-26

package xbase_test

import (
	"fmt"

	"github.com/xanygo/anygo/xenc/xcipher"
)

func ExampleAesBlock_Encrypt() {
	ac := &xcipher.AesBlock{
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
	ac := &xcipher.AesOFB{
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
