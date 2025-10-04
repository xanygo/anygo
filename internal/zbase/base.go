//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-04

package zbase

import "strconv"

func BaseTypeToString(va any) (string, bool) {
	switch vv := va.(type) {
	case string:
		return vv, true
	case int:
		return strconv.Itoa(vv), true
	case int8:
		return strconv.FormatInt(int64(vv), 10), true
	case int16:
		return strconv.FormatInt(int64(vv), 10), true
	case int32:
		return strconv.FormatInt(int64(vv), 10), true
	case int64:
		return strconv.FormatInt(vv, 10), true

	case uint:
		return strconv.FormatUint(uint64(vv), 10), true
	case uint8:
		return strconv.FormatUint(uint64(vv), 10), true
	case uint16:
		return strconv.FormatUint(uint64(vv), 10), true
	case uint32:
		return strconv.FormatUint(uint64(vv), 10), true
	case uint64:
		return strconv.FormatUint(vv, 10), true

	case float32:
		return strconv.FormatFloat(float64(vv), 'f', -1, 32), true
	case float64:
		return strconv.FormatFloat(vv, 'f', -1, 64), true

	case bool:
		return strconv.FormatBool(vv), true
	default:
		return "", false
	}
}
