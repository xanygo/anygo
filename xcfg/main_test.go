//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package xcfg

import (
	"os"

	"github.com/xanygo/anygo/xattr"
)

func init() {
	_ = os.Setenv("Port1", "8080")
	_ = os.Setenv("Port2", "8081")
	_ = os.Setenv("APP", "demo.fenji")
	testReset()
}

func testReset() {
	xattr.Init("test", "./testdata")
}
