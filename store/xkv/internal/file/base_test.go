package file

import (
	"path/filepath"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestBase_WriteKVDataFile2(t *testing.T) {
	fb := Base{
		Key: "hello",
		Dir: filepath.Join(t.TempDir(), "fb1"),
	}

	added, err := fb.WriteKVDataFile2("k1", "hello")
	xt.NoError(t, err)
	xt.True(t, added)

	added, err = fb.WriteKVDataFile2("k1", "hello")
	xt.NoError(t, err)
	xt.False(t, added) // 内容一样

	added, err = fb.WriteKVDataFile2("k1", "world")
	xt.NoError(t, err)
	xt.True(t, added) // 内容有变化
}
