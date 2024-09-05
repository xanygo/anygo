//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-05

package xflag

import (
	"flag"
	"strconv"
)

type float32Value float32

func newFloat32Value(val float32, p *float32) *float32Value {
	*p = val
	return (*float32Value)(p)
}

func (f *float32Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	*f = float32Value(v)
	return err
}

func (f *float32Value) Get() any { return float64(*f) }

func (f *float32Value) String() string { return strconv.FormatFloat(float64(*f), 'g', -1, 64) }

func Float32Var(p *float32, name string, value float32, usage string) {
	flag.CommandLine.Var(newFloat32Value(value, p), name, usage)
}

func Float32(name string, value float32, usage string) *float32 {
	p := new(float32)
	Float32Var(p, name, value, usage)
	return p
}
