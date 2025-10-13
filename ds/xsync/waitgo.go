//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-13

package xsync

import (
	"errors"
	"sync"

	"github.com/xanygo/anygo/safely"
)

type WaitGo struct {
	wg   sync.WaitGroup
	mu   sync.Mutex
	errs []error
}

func (w *WaitGo) Go(f func()) {
	w.wg.Go(func() {
		err := safely.Run(f)
		if err == nil {
			return
		}
		w.mu.Lock()
		w.errs = append(w.errs, err)
		w.mu.Unlock()
	})
}

func (w *WaitGo) Go1(f func() error) {
	w.wg.Go(func() {
		err := safely.Run(f)
		if err == nil {
			return
		}
		w.mu.Lock()
		w.errs = append(w.errs, err)
		w.mu.Unlock()
	})
}

func (w *WaitGo) Wait() error {
	w.wg.Wait()
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(w.errs) == 0 {
		return nil
	}
	return errors.Join(w.errs...)
}
