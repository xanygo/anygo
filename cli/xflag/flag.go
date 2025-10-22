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

func (f *float32Value) Get() any {
	return float64(*f)
}

func (f *float32Value) String() string {
	return strconv.FormatFloat(float64(*f), 'g', -1, 64)
}

func Float32Var(p *float32, name string, value float32, usage string) {
	flag.CommandLine.Var(newFloat32Value(value, p), name, usage)
}

func Float32(name string, value float32, usage string) *float32 {
	p := new(float32)
	Float32Var(p, name, value, usage)
	return p
}

// ----------------------------------------------------------------------------

type int8Value int8

func newInt8Value(val int8, p *int8) *int8Value {
	*p = val
	return (*int8Value)(p)
}

func (f *int8Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 10, 8)
	*f = int8Value(v)
	return err
}

func (f *int8Value) Get() any {
	return int8(*f)
}

func (f *int8Value) String() string {
	return strconv.FormatUint(uint64(*f), 10)
}

func Int8Var(p *int8, name string, value int8, usage string) {
	flag.CommandLine.Var(newInt8Value(value, p), name, usage)
}

func Int8(name string, value int8, usage string) *int8 {
	p := new(int8)
	Int8Var(p, name, value, usage)
	return p
}

// ----------------------------------------------------------------------------

type uint8Value uint8

func newUint8Value(val uint8, p *uint8) *uint8Value {
	*p = val
	return (*uint8Value)(p)
}

func (f *uint8Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 10, 8)
	*f = uint8Value(v)
	return err
}

func (f *uint8Value) Get() any {
	return uint8(*f)
}

func (f *uint8Value) String() string {
	return strconv.FormatUint(uint64(*f), 10)
}

func Uint8Var(p *uint8, name string, value uint8, usage string) {
	flag.CommandLine.Var(newUint8Value(value, p), name, usage)
}

func Uint8(name string, value uint8, usage string) *uint8 {
	p := new(uint8)
	Uint8Var(p, name, value, usage)
	return p
}

// ----------------------------------------------------------------------------

type int16Value int16

func newInt16Value(val int16, p *int16) *int16Value {
	*p = val
	return (*int16Value)(p)
}

func (f *int16Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 10, 16)
	*f = int16Value(v)
	return err
}

func (f *int16Value) Get() any {
	return int16(*f)
}

func (f *int16Value) String() string {
	return strconv.FormatUint(uint64(*f), 10)
}

func Int16Var(p *int16, name string, value int16, usage string) {
	flag.CommandLine.Var(newInt16Value(value, p), name, usage)
}

func Int16(name string, value int16, usage string) *int16 {
	p := new(int16)
	Int16Var(p, name, value, usage)
	return p
}

// ----------------------------------------------------------------------------

type uint16Value uint16

func newUint16Value(val uint16, p *uint16) *uint16Value {
	*p = val
	return (*uint16Value)(p)
}

func (f *uint16Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 10, 16)
	*f = uint16Value(v)
	return err
}

func (f *uint16Value) Get() any {
	return uint16(*f)
}

func (f *uint16Value) String() string {
	return strconv.FormatUint(uint64(*f), 10)
}

func Uint16Var(p *uint16, name string, value uint16, usage string) {
	flag.CommandLine.Var(newUint16Value(value, p), name, usage)
}

func Uint16(name string, value uint16, usage string) *uint16 {
	p := new(uint16)
	Uint16Var(p, name, value, usage)
	return p
}

// ----------------------------------------------------------------------------

type int32Value int32

func newInt32Value(val int32, p *int32) *int32Value {
	*p = val
	return (*int32Value)(p)
}

func (f *int32Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 10, 32)
	*f = int32Value(v)
	return err
}

func (f *int32Value) Get() any {
	return int32(*f)
}

func (f *int32Value) String() string {
	return strconv.FormatUint(uint64(*f), 10)
}

func Int32Var(p *int32, name string, value int32, usage string) {
	flag.CommandLine.Var(newInt32Value(value, p), name, usage)
}

func Int32(name string, value int32, usage string) *int32 {
	p := new(int32)
	Int32Var(p, name, value, usage)
	return p
}

// ----------------------------------------------------------------------------

type uint32Value uint32

func newUint32Value(val uint32, p *uint32) *uint32Value {
	*p = val
	return (*uint32Value)(p)
}

func (f *uint32Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 10, 32)
	*f = uint32Value(v)
	return err
}

func (f *uint32Value) Get() any {
	return uint32(*f)
}

func (f *uint32Value) String() string {
	return strconv.FormatUint(uint64(*f), 10)
}

func Uint32Var(p *uint32, name string, value uint32, usage string) {
	flag.CommandLine.Var(newUint32Value(value, p), name, usage)
}

func Uint32(name string, value uint32, usage string) *uint32 {
	p := new(uint32)
	Uint32Var(p, name, value, usage)
	return p
}
