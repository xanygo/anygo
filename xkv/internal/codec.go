//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-08

package internal

import "github.com/xanygo/anygo/xenc/xcodec"

func EncodeToStrings[V any](codec xcodec.Marshaler, members []V) ([]string, []error) {
	ms := make([]string, 0, len(members))
	var errs []error
	for _, member := range members {
		str, err := xcodec.MarshalToString(codec, member)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		ms = append(ms, str)
	}
	return ms, errs
}

func EncodeMapValueToStrings[K comparable, V any](codec xcodec.Marshaler, values map[K]V) (map[K]string, []error) {
	result := make(map[K]string, len(values))
	var errs []error
	for key, value := range values {
		str, err := xcodec.MarshalToString(codec, value)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		result[key] = str
	}
	return result, errs
}
