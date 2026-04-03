//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package xattr_test

import (
	"fmt"

	"github.com/xanygo/anygo/xattr"
)

func ExampleIDC() {
	fmt.Println("idc=", xattr.IDC()) // idc= online

	// Output:
	// idc= dev
}

func ExampleRunMode() {
	fmt.Println("runMode=", xattr.RunMode()) // runMode= product

	// Output:
	// runMode= debug
}

func ExampleAppName() {
	fmt.Println("appName=", xattr.AppName())

	// Output:
	// appName= xattr
}
