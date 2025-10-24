//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-23

package xcast

import (
	"math"
	"net"
	"strconv"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestInteger(t *testing.T) {
	testInteger[int](t, "123", 123, true)

	testInteger[int](t, 123, 123, true)
	testInteger[int](t, int(123), 123, true)
	testInteger[int](t, int8(123), 123, true)
	testInteger[int](t, int16(123), 123, true)
	testInteger[int](t, int32(123), 123, true)
	testInteger[int](t, int64(123), 123, true)

	testInteger[int](t, uint(123), 123, true)
	testInteger[int](t, uint8(123), 123, true)
	testInteger[int](t, uint16(123), 123, true)
	testInteger[int](t, uint32(123), 123, true)
	testInteger[int](t, uint64(123), 123, true)

	testInteger[int](t, math.Inf(-1), 0, false)
	testInteger[int](t, math.Inf(1), 0, false)

	testInteger[int](t, float64(123.0), 123, true)
	testInteger[int](t, float32(123.0), 123, true)

	testInteger[int](t, true, 1, true)
	testInteger[int](t, false, 0, true)

	testInteger[int](t, float32(123), 123, true)
	testInteger[int](t, float32(math.Inf(-1)), 0, false)
	testInteger[int](t, math.Inf(-1), 0, false)
	testInteger[int](t, math.NaN(), 0, false)

	testInteger[int](t, uint(1), 1, true)
	testInteger[int8](t, uint8(1), 1, true)
	testInteger[int16](t, uint16(1), 1, true)
	testInteger[int32](t, uint32(1), 1, true)
	testInteger[int64](t, uint64(1), 1, true)

	testInteger[uint](t, int(1), 1, true)
	testInteger[uint8](t, int8(1), 1, true)
	testInteger[uint16](t, int16(1), 1, true)
	testInteger[uint32](t, int32(1), 1, true)
	testInteger[uint64](t, int64(1), 1, true)

	testInteger[uint8](t, 12345, 0, false)
	testInteger[uint16](t, math.MaxUint16+1, 0, false)
	testInteger[uint32](t, math.MaxUint32+1, 0, false)
	testInteger[uint64](t, math.MaxFloat64, 0, false)

	testInteger[int8](t, math.MaxInt8+1, 0, false)
	testInteger[int16](t, math.MaxInt16+1, 0, false)
	testInteger[int32](t, math.MaxInt32+1, 0, false)
	testInteger[int64](t, uint64(math.MaxUint64), 0, false)

	testInteger[int](t, "-INF", 0, false)

	testInteger[int](t, nil, 0, false)
	testInteger[int](t, &net.AddrError{}, 0, false)

	testInteger[uint8](t, strconv.FormatUint(math.MaxUint64, 10), 0, false)
	testInteger[uint16](t, strconv.FormatUint(math.MaxUint64, 10), 0, false)
	testInteger[uint32](t, strconv.FormatUint(math.MaxUint64, 10), 0, false)

	testInteger[uint](t, "255", 255, true)
	testInteger[uint](t, "-1", 0, false)
	testInteger[uint8](t, "255", 255, true)
	testInteger[uint16](t, "65535", 65535, true)
	testInteger[uint32](t, strconv.FormatUint(math.MaxUint32, 10), math.MaxUint32, true)
	testInteger[uint64](t, strconv.FormatUint(math.MaxUint64, 10), math.MaxUint64, true)

	testInteger[uint](t, strconv.FormatUint(math.MaxUint, 10), math.MaxUint, true)

	testInteger[int8](t, strconv.FormatUint(math.MaxUint64, 10), 0, false)
	testInteger[int16](t, strconv.FormatUint(math.MaxUint64, 10), 0, false)
	testInteger[int32](t, strconv.FormatUint(math.MaxUint64, 10), 0, false)

	testInteger[int](t, "123", 123, true)
	testInteger[int8](t, strconv.FormatUint(math.MaxInt8, 10), math.MaxInt8, true)
	testInteger[int16](t, strconv.FormatUint(math.MaxInt16, 10), math.MaxInt16, true)
	testInteger[int32](t, strconv.FormatUint(math.MaxInt32, 10), math.MaxInt32, true)
	testInteger[int64](t, strconv.FormatUint(math.MaxInt64, 10), math.MaxInt64, true)

	xt.Equal(t, 123, ToInteger[int]("123"))
}

func testInteger[T IntegerTypes](t *testing.T, v any, val T, ok bool) {
	t.Helper()
	num, status := Integer[T](v)
	xt.Equal[T](t, val, num)
	xt.Equal(t, ok, status)
}

func TestFloat(t *testing.T) {
	testFloat[float32](t, "", 0, false)
	testFloat[float32](t, nil, 0, false)
	testFloat[float64](t, "", 0, false)
	testFloat[float32](t, "123", 123, true)
	testFloat[float64](t, "123", 123, true)

	testFloat[float32](t, true, 1, true)
	testFloat[float32](t, false, 0, true)
	testFloat[float64](t, true, 1, true)
	testFloat[float64](t, false, 0, true)

	testFloat[float32](t, float32(123), 123, true)
	testFloat[float64](t, float64(123), 123, true)

	testFloat[float32](t, uint(8), 8, true)
	testFloat[float32](t, uint8(8), 8, true)
	testFloat[float32](t, uint16(8), 8, true)
	testFloat[float32](t, uint32(8), 8, true)
	testFloat[float32](t, uint64(8), 8, true)

	testFloat[float32](t, int(8), 8, true)
	testFloat[float32](t, int8(8), 8, true)
	testFloat[float32](t, int16(8), 8, true)
	testFloat[float32](t, int32(8), 8, true)
	testFloat[float32](t, int64(8), 8, true)

	testFloat[float64](t, 123, 123, true)

	testFloat[float64](t, uint(8), 8, true)
	testFloat[float64](t, uint8(8), 8, true)
	testFloat[float64](t, uint16(8), 8, true)
	testFloat[float64](t, uint32(8), 8, true)
	testFloat[float64](t, uint64(8), 8, true)

	testFloat[float64](t, int(8), 8, true)
	testFloat[float64](t, int8(8), 8, true)
	testFloat[float64](t, int16(8), 8, true)
	testFloat[float64](t, int32(8), 8, true)
	testFloat[float64](t, int64(8), 8, true)

	xt.Equal(t, 123.0, ToFloat[float64]("123"))
}

func testFloat[T FloatTypes](t *testing.T, v any, val T, ok bool) {
	t.Helper()
	num, status := Float[T](v)
	xt.Equal[T](t, val, num)
	xt.Equal(t, ok, status)
}
