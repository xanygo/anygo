//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package xattr

import (
	"os"
	"testing"

	"github.com/fsgo/fst"
)

func Test_osEnvDefault(t *testing.T) {
	key := "fsenv_k1"
	fst.NoError(t, os.Unsetenv(key))
	defer fst.NoError(t, os.Unsetenv(key))

	fst.Equal(t, "v1", osEnvDefault(key, "v1"))
	fst.NoError(t, os.Setenv(key, "v2"))
	fst.Equal(t, "v2", osEnvDefault(key, "v1"))
}
