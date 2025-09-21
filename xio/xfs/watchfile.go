//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-19

package xfs

import (
	"errors"
	"sync"
	"time"

	"github.com/xanygo/anygo/ds/xslice"
)

// WatchFile 监听单个文件
type WatchFile struct {
	// FileName 文件名、文件路径，必填
	FileName string

	// Parser 解析当前文件的回调，必填
	// 在首次初始化以及文件变化的时候会调用
	Parser func(path string) error

	onStop func()

	afterChanges xslice.Sync[func()]

	mux     sync.RWMutex
	started bool
}

// Start  watch start
func (wf *WatchFile) Start() error {
	wf.mux.Lock()
	defer wf.mux.Unlock()

	if wf.started {
		return errors.New("already started")
	}
	if err := wf.Load(); err != nil {
		return err
	}

	watcher := &Watcher{
		Interval: time.Second,
		Delay:    time.Second,
	}
	watcher.Watch(wf.FileName, func(event WatcherEvent) {
		_ = wf.Load()
		all := wf.afterChanges.Load()
		for _, fn := range all {
			fn()
		}
	})

	wf.onStop = watcher.Stop
	err := watcher.Start()
	if err == nil {
		wf.started = true
	}
	return err
}

// Load 调用 Parser 回调函数解析文件内容
func (wf *WatchFile) Load() error {
	if len(wf.FileName) == 0 {
		return errors.New("fileName is empty")
	}
	if wf.Parser == nil {
		return errors.New("parser func is nil")
	}
	return wf.Parser(wf.FileName)
}

// OnChange register file change callback
func (wf *WatchFile) OnChange(fn func()) {
	wf.afterChanges.Append(fn)
}

// Stop watch stop
func (wf *WatchFile) Stop() {
	wf.mux.Lock()
	defer wf.mux.Unlock()
	if !wf.started {
		return
	}
	wf.onStop()
	wf.started = false
}
