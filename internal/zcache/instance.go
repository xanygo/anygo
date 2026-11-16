package zcache

import (
	"regexp"
)

type RegexpValue struct {
	Regexp *regexp.Regexp
	Err    error
}

var Regexp = &MapCache[string, *RegexpValue]{
	New: func(key string) *RegexpValue {
		r, err := regexp.Compile(key)
		value := &RegexpValue{
			Regexp: r,
			Err:    err,
		}
		return value
	},
}
