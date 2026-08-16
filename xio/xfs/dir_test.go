package xfs

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xanygo/anygo/xt"
)

func TestRemoveEmptyDir(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		dir1 := filepath.Join(t.TempDir(), "case1")
		dir2 := filepath.Join(dir1, "data")
		err := KeepDirExists(dir2)
		xt.NoError(t, err)

		fp := filepath.Join(dir2, "fp.txt")
		err = os.WriteFile(fp, []byte("hello"), 0777)
		xt.NoError(t, err)
		num, err := RemoveEmptyDir(dir1, time.Time{})
		xt.NoError(t, err)
		xt.Equal(t, num, 0)

		err = os.Remove(fp)
		xt.NoError(t, err)

		num, err = RemoveEmptyDir(dir1, time.Time{})
		xt.NoError(t, err)
		xt.Equal(t, num, 1)

		ok, err := Exists(dir2)
		xt.NoError(t, err)
		xt.False(t, ok)
	})
	t.Run("case 2", func(t *testing.T) {
		dir1 := filepath.Join(t.TempDir(), "case1")
		dir2 := filepath.Join(dir1, "data")
		err := KeepDirExists(dir2)
		xt.NoError(t, err)

		fp := filepath.Join(dir2, "fp.txt")
		err = os.WriteFile(fp, []byte("hello"), 0777)
		xt.NoError(t, err)

		expire := time.Now().Add(-1 * time.Hour)

		num, err := RemoveEmptyDir(dir1, expire)
		xt.NoError(t, err)
		xt.Equal(t, num, 0)

		err = os.Remove(fp)
		xt.NoError(t, err)

		num, err = RemoveEmptyDir(dir1, expire)
		xt.NoError(t, err)
		xt.Equal(t, num, 0)

		ok, err := Exists(dir2)
		xt.NoError(t, err)
		xt.True(t, ok)
	})
}
