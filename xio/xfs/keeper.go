//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-19

package xfs

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/xpp"
)

// Keeper 保持文件存在
type Keeper struct {
	// FilePath 返回文件地址，必填
	FilePath func() string

	// OpenFile 创建文件的函数，可选
	// 默认为 os.OpenFile(fp, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	OpenFile func(fp string) (*os.File, error)

	// CheckInterval 检查间隔，可选
	// 默认为 50ms
	CheckInterval time.Duration

	file  *os.File
	info  os.FileInfo
	timer *xpp.Interval

	beforeChange fileChange
	afterChange  fileChange

	mux     sync.RWMutex
	running bool
}

func (kf *Keeper) getCheckInterval() time.Duration {
	if kf.CheckInterval > 0 {
		return kf.CheckInterval
	}
	return 50 * time.Millisecond
}

// Start 开始,非阻塞运行
//
//	与之对应的有 Stop 方法
func (kf *Keeper) Start() error {
	if kf.FilePath == nil {
		return errors.New("fn FilePath is nil")
	}

	if err := kf.checkFile(); err != nil {
		return err
	}
	kf.mux.Lock()
	defer kf.mux.Unlock()
	if kf.running {
		return errors.New("already started")
	}
	kf.running = true
	kf.timer = &xpp.Interval{}
	kf.timer.Add(kf.loop)
	kf.timer.Start(kf.getCheckInterval())
	return nil
}

func (kf *Keeper) loop() {
	if err := kf.checkFile(); err != nil {
		log.Println("[anygo][xfs][Keeper][error]", err)
	}
	kf.timer.Reset(kf.getCheckInterval())
}

// Stop 停止运行
func (kf *Keeper) Stop() {
	kf.mux.Lock()
	defer kf.mux.Unlock()
	if !kf.running {
		return
	}
	kf.running = false
	kf.timer.Stop()
	_ = kf.file.Close()
}

// File 获取文件
func (kf *Keeper) File() *os.File {
	kf.mux.RLock()
	defer kf.mux.RUnlock()
	return kf.file
}

// BeforeChange 注册当文件变化前的回调函数
func (kf *Keeper) BeforeChange(fn func(old *os.File)) {
	kf.beforeChange.register(fn)
}

// AfterChange 注册当文件变化后的回调函数
func (kf *Keeper) AfterChange(fn func(newFile *os.File)) {
	kf.afterChange.register(fn)
}

func (kf *Keeper) checkFile() error {
	fp := kf.FilePath()

	if len(fp) == 0 {
		return errors.New("empty file path")
	}

	if has, err := kf.exists(fp); has {
		return nil
	} else if err != nil {
		return err
	}

	file, err := kf.openFile(fp)
	if err != nil {
		return err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return err
	}

	kf.mux.RLock()
	old := kf.file
	kf.mux.RUnlock()

	if old != nil {
		kf.beforeChange.onChange(old)
	}

	kf.mux.Lock()
	kf.file = file
	kf.info = info
	kf.mux.Unlock()

	kf.afterChange.onChange(file)

	if old != nil {
		return old.Close()
	}
	return nil
}

func (kf *Keeper) openFile(fp string) (*os.File, error) {
	dir := filepath.Dir(fp)
	err := KeepDirExists(dir)
	if err != nil {
		return nil, err
	}
	if kf.OpenFile != nil {
		return kf.OpenFile(fp)
	}
	return os.OpenFile(fp, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
}

func (kf *Keeper) exists(fp string) (bool, error) {
	kf.mux.RLock()
	info := kf.info
	kf.mux.RUnlock()

	if info == nil {
		return false, nil
	}
	curInfo, err := os.Stat(fp)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return os.SameFile(info, curInfo), nil
}

type fileChange struct {
	callbacks xslice.Sync[func(f *os.File)]
}

func (fc *fileChange) register(fn func(f *os.File)) {
	fc.callbacks.Append(fn)
}

func (fc *fileChange) onChange(f *os.File) {
	fns := fc.callbacks.Load()
	for _, fn := range fns {
		fn(f)
	}
}
