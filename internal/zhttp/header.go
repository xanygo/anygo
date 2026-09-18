package zhttp

import (
	"maps"
	"net/http"

	"github.com/xanygo/anygo/xslice"
)

// HeaderDifference 查找到 a 相比 b 增量的部分，总是返回一个全新的 Header
func HeaderDifference(a, b http.Header) http.Header {
	if len(a) == 0 {
		return nil
	}
	if len(b) == 0 {
		return maps.Clone(a)
	}
	result := make(http.Header)
	for key, values := range a {
		if diff := xslice.Difference(values, b[key]); len(diff) > 0 {
			result[key] = diff
		}
	}
	return result
}
