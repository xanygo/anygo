//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package xattr

import (
	"os"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func Test_osEnvDefault(t *testing.T) {
	key := "fsenv_k1"
	xt.NoError(t, os.Unsetenv(key))
	defer xt.NoError(t, os.Unsetenv(key))

	xt.Equal(t, "v1", osEnvDefault(key, "v1"))
	xt.NoError(t, os.Setenv(key, "v2"))
	xt.Equal(t, "v2", osEnvDefault(key, "v1"))
}
