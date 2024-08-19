//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-19

package xerrors

import "sync"

var _ error = (*OnceSet)(nil)

type OnceSet struct {
	err error
	mux sync.RWMutex
	has bool
}

func (o *OnceSet) Error() string {
	if o == nil {
		return "<nil>"
	}
	o.mux.RLock()
	defer o.mux.RUnlock()
	if o.err != nil {
		return o.err.Error()
	}
	return "<nil>"
}

func (o *OnceSet) SetOnce(err error) {
	o.mux.RLock()
	has := o.has
	o.mux.RUnlock()
	if has {
		return
	}
	o.mux.Lock()
	if !o.has {
		o.err = err
	}
	o.mux.Unlock()
}

func (o *OnceSet) Replace(err error) {
	o.mux.Lock()
	o.err = err
	o.mux.Unlock()
}

func (o *OnceSet) Clear() {
	o.mux.Lock()
	o.has = false
	o.err = nil
	o.mux.Unlock()
}

func (o *OnceSet) Unwrap() error {
	if o == nil {
		return nil
	}
	o.mux.RLock()
	defer o.mux.RUnlock()
	return o.err
}
