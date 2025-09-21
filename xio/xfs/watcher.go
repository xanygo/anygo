//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-19

package xfs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/safely"
	"github.com/xanygo/anygo/xpp"
)

type WatcherEventType string

const (
	// WatcherEventUpdate contains create and update ent
	WatcherEventUpdate WatcherEventType = "update"

	// WatcherEventDelete  delete event
	WatcherEventDelete WatcherEventType = "delete"
)

// WatcherEvent event for watcher
type WatcherEvent struct {
	FileName  string
	EventType WatcherEventType
}

// String event desc
func (we *WatcherEvent) String() string {
	return we.FileName + " " + string(we.EventType)
}

// Watcher 通过定时器定期检查实现的文件、目录监听功能
type Watcher struct {
	// Interval 检查的时间间隔，可选，默认为 1 秒
	Interval time.Duration

	// Delay 文件变化后多久可以被检查出，可选，默认为 1 秒
	Delay time.Duration

	timer   *xpp.Interval
	rules   xslice.Sync[*watchRule]
	mux     sync.RWMutex
	started bool
	errors  chan error
}

// Watch 对一个文件路径添加监听
//
//	pattern: 文件名或者包含*的文件路径，规则同 filepath.Glob 的参数
//	callback: 回调函数
func (w *Watcher) Watch(pattern string, callback func(event WatcherEvent)) {
	if pattern == "" || callback == nil {
		return
	}
	wd := &watchRule{
		Pattern:  pattern,
		CallBack: callback,
		delay:    time.Second,
	}
	if w.Delay > 0 {
		wd.delay = w.Delay
	}
	w.rules.Append(wd)
}

func (w *Watcher) getInterval() time.Duration {
	if w.Interval > 0 {
		return w.Interval
	}
	return time.Second
}

// Start ticker start async
func (w *Watcher) Start() error {
	w.mux.Lock()
	defer w.mux.Unlock()
	if w.started {
		return errors.New("already started")
	}
	w.timer = &xpp.Interval{}
	w.timer.Add(w.scan)
	w.timer.Start(w.getInterval())
	w.started = true
	w.errors = make(chan error)
	return nil
}

func (w *Watcher) scan() {
	for _, rule := range w.rules.Load() {
		safely.RunVoid(func() {
			rule.scan(w.errors)
		})
	}
}

func (w *Watcher) Errors() <-chan error {
	return w.errors
}

// Stop ticker stop
func (w *Watcher) Stop() {
	w.mux.Lock()
	defer w.mux.Unlock()
	if !w.started {
		return
	}
	w.timer.Stop()
	w.started = false
	close(w.errors)
}

type watchRule struct {
	CallBack func(event WatcherEvent)
	last     map[string]time.Time
	Pattern  string
	delay    time.Duration
}

func (wr *watchRule) checkDelay(modTime time.Time) bool {
	return modTime.Before(time.Now().Add(-1 * wr.delay))
}

func (wr *watchRule) scan(ec chan<- error) {
	matches, err := filepath.Glob(wr.Pattern)
	if err != nil {
		select {
		case ec <- fmt.Errorf("glob(%q): %w", wr.Pattern, err):
		default:
		}
		return
	}
	if wr.last == nil {
		wr.last = map[string]time.Time{}
	}
	nowData := map[string]time.Time{}
	for _, name := range matches {
		info, err := os.Stat(name)
		if err != nil {
			select {
			case ec <- fmt.Errorf("os.Stat(%q): %w", name, err):
			default:
			}
			continue
		}
		lastMod, has := wr.last[name]
		// 新增 或者有变更的情况
		if !has || !info.ModTime().Equal(lastMod) {
			if wr.checkDelay(info.ModTime()) {
				nowData[name] = info.ModTime()
				event := WatcherEvent{
					FileName:  name,
					EventType: WatcherEventUpdate,
				}
				wr.CallBack(event)
			} else {
				nowData[name] = info.ModTime().Add(-1)
			}
		} else {
			// 没有变化的情况
			nowData[name] = lastMod
		}
	}
	for name := range wr.last {
		// 针对已删除的场景
		if _, has := nowData[name]; !has {
			event := WatcherEvent{
				FileName:  name,
				EventType: WatcherEventDelete,
			}
			wr.CallBack(event)
		}
	}
	wr.last = nowData
}
