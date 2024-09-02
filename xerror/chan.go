//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-02

package xerror

import "sync"

type Chan struct {
	errs chan error
	once sync.Once
}

func (e *Chan) init() {
	e.errs = make(chan error, 1)
}

func (e *Chan) Send(err error) {
	e.once.Do(e.init)
	select {
	case e.errs <- err:
	default:
	}
}

func (e *Chan) Watch() <-chan error {
	e.once.Do(e.init)
	return e.errs
}
