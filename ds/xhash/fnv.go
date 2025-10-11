//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-12-21

package xhash

import (
	"hash/fnv"
	"unsafe"
)

func Fnv64(s string) uint64 {
	bf := unsafe.Slice(unsafe.StringData(s), len(s))
	h := fnv.New64()
	_, _ = h.Write(bf)
	return h.Sum64()
}

func ByteFnv64(bf []byte) uint64 {
	h := fnv.New64()
	_, _ = h.Write(bf)
	return h.Sum64()
}
