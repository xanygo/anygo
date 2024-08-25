//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-19

package xfs

import (
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fsgo/fst"
)

func TestKeepFile(t *testing.T) {
	fp := filepath.Join("testdata", "tmp", "keep.txt")
	defer os.Remove(fp)
	const ci = 50 * time.Millisecond
	kp := &Keeper{
		FilePath: func() string {
			return fp
		},
		CheckInterval: ci,
	}
	var changeNum int32
	kp.AfterChange(func(f *os.File) {
		atomic.AddInt32(&changeNum, 1)
	})

	fst.NoError(t, kp.Start())

	t.Run("after start", func(t *testing.T) {
		fst.Equal(t, int32(1), atomic.LoadInt32(&changeNum))
		fst.NotNil(t, kp.File())
	})

	defer kp.Stop()

	checkExists := func(t *testing.T) {
		info, err := os.Stat(fp)
		fst.NoError(t, err)
		fst.NotEmpty(t, info.Name())
	}

	t.Run("same file not change", func(t *testing.T) {
		stat1, err := kp.File().Stat()
		fst.NoError(t, err)
		time.Sleep(ci * 2)

		stat2, err := kp.File().Stat()
		fst.NoError(t, err)

		fst.True(t, os.SameFile(stat1, stat2))
	})

	t.Run("rm and create it auto", func(t *testing.T) {
		// on Windows
		//  remove testdata\tmp\keep.txt: The process cannot access the file because it is being used by another process.
		if runtime.GOOS == "windows" {
			t.SkipNow()
		}
		checkExists(t)
		fst.NoError(t, os.Remove(fp))
		time.Sleep(ci * 2)
		checkExists(t)
		fst.Equal(t, int32(2), atomic.LoadInt32(&changeNum))
	})

	t.Run("stopped", func(t *testing.T) {
		kp.Stop()
		time.Sleep(ci * 2)
		fst.NoError(t, os.Remove(fp))

		// check not exists
		_, err := os.Stat(fp)
		fst.Error(t, err)
	})
}
