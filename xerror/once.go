//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-19

package xerror

import "sync"

var _ error = (*OnceSet)(nil)

// OnceSet 只允许设置一次的用于存储 error 信息的对象
type OnceSet struct {
	err error
	mux sync.Mutex
}

func (o *OnceSet) Error() string {
	if o == nil {
		return "<nil>"
	}
	o.mux.Lock()
	err := o.err
	o.mux.Unlock()
	if err != nil {
		return err.Error()
	}
	return "<nil>"
}

// SetOnce 存储 error，当没有存储过则成功，否则失败
func (o *OnceSet) SetOnce(err error) bool {
	if err == nil {
		return false
	}
	o.mux.Lock()
	if o.err == nil {
		o.err = err
	}
	o.mux.Unlock()
	return true
}

// Replace 强制替换值，不判断是否存储过
func (o *OnceSet) Replace(err error) {
	if err == nil {
		return
	}
	o.mux.Lock()
	o.err = err
	o.mux.Unlock()
}

// Clear 清除存储的值和状态，调用后再次调用 SetOnce 可成功存储
func (o *OnceSet) Clear() {
	o.mux.Lock()
	o.err = nil
	o.mux.Unlock()
}

// Unwrap 返回底层存储的 error
func (o *OnceSet) Unwrap() error {
	if o == nil {
		return nil
	}
	o.mux.Lock()
	err := o.err
	o.mux.Unlock()
	return err
}
