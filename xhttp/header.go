//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-08

package xhttp

import (
	"maps"
	"net/http"

	"github.com/xanygo/anygo/ds/xslice"
)

// HeaderDiffMore 查找到 new 相比 old 增量的部分，总是返回一个全新的 Header
func HeaderDiffMore(old, new http.Header) http.Header {
	if len(old) == 0 {
		return maps.Clone(new)
	}
	result := make(http.Header)
	for key, values := range new {
		if diff := xslice.DiffMore(values, old[key]); len(diff) > 0 {
			result[key] = diff
		}
	}
	return result
}

func WriteHeader(w http.ResponseWriter, header http.Header) {
	for key, vs := range header {
		for _, value := range vs {
			w.Header().Add(key, value)
		}
	}
}
