//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-20

package zos

import "sync"

var globalMux = &sync.Mutex{}

func GlobalLock() {
	globalMux.Lock()
}

func GlobalUnlock() {
	globalMux.Unlock()
}
