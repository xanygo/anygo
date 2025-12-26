//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-12-25

package xredis

import (
	"context"
	"testing"
	"time"

	"github.com/xanygo/anygo/internal/redistest"
	"github.com/xanygo/anygo/xt"
)

func TestBit(t *testing.T) {
	ts, errTs := redistest.NewServer()
	if errTs != nil {
		t.Logf("create redis fail: %v", errTs)
		return
	}
	defer ts.Stop()
	t.Logf("uri= %q", ts.URI())
	_, client, errClient := NewClientByURI("demo", ts.URI())
	xt.NoError(t, errClient)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("BitCount", func(t *testing.T) {
		got, err := client.BitCount(ctx, "BitCount-1")
		xt.NoError(t, err)
		xt.Equal(t, 0, got)

		err = client.Set(ctx, "BitCount-1", "demo")
		xt.NoError(t, err)

		got, err = client.BitCount(ctx, "BitCount-1")
		xt.NoError(t, err)
		xt.Equal(t, 18, got)
	})

	t.Run("BitField", func(t *testing.T) {
		got, err := client.BitField(ctx, "BitField-1")
		xt.NoError(t, err)
		xt.Empty(t, got)

		// 会返回  nil，0 的case
		// BITFIELD BitField-1 OVERFLOW FAIL INCRBY u8 0 256  GET u8 0
		got, err = client.BitField(context.Background(), "BitField-1",
			BitFieldOverflow{FAIL: true},
			BitFieldIncrBy{Encoding: "u8", Increment: 256},
			BitFieldGet{Encoding: "u8"})
		xt.NoError(t, err)
		zero := int64(0)
		xt.Equal(t, []*int64{nil, &zero}, got)
	})

	t.Run("BitFieldRo", func(t *testing.T) {
		got, err := client.BitFieldRo(ctx, "BitFieldRo-1", BitFieldSet{Encoding: "u8"})
		xt.Error(t, err)
		xt.Empty(t, got)

		got, err = client.BitFieldRo(ctx, "BitFieldRo-1", BitFieldGet{Encoding: "u8"})
		xt.NoError(t, err)
		xt.Equal(t, []int64{0}, got)
	})

	t.Run("BitOP", func(t *testing.T) {
		got, err := client.BitOP(ctx, "OR", "BitOPDest-1", "BitOP-1", "BitOP-2")
		xt.NoError(t, err)
		xt.Equal(t, 0, got)

		err = client.Set(ctx, "BitOP-1", "demo")
		xt.NoError(t, err)

		got, err = client.BitOP(ctx, "OR", "BitOPDest-1", "BitOP-1", "BitOP-2")
		xt.NoError(t, err)
		xt.Equal(t, 4, got)
	})

	t.Run("BitPos invalid args", func(t *testing.T) {
		_, err := client.BitPos(ctx, "BitPos-invalid", 2)
		xt.Error(t, err)

		_, err = client.BitPos(ctx, "BitPos-invalid", 1, "err")
		xt.Error(t, err)

		_, err = client.BitPos(ctx, "BitPos-invalid", 1, 1, "err")
		xt.Error(t, err)

		_, err = client.BitPos(ctx, "BitPos-invalid", 1, 1, 1, 1)
		xt.Error(t, err)

		_, err = client.BitPos(ctx, "BitPos-invalid", 1, 1, 1, "BIT", 1)
		xt.Error(t, err)
	})

	t.Run("BitPos-ok", func(t *testing.T) {
		got, err := client.BitPos(ctx, "BitPos-1", 1)
		xt.NoError(t, err)
		xt.Equal(t, -1, got)

		got, err = client.BitPos(ctx, "BitPos-1", 0)
		xt.NoError(t, err)
		xt.Equal(t, 0, got)

		err = client.Set(ctx, "BitPos-1", "demo")
		xt.NoError(t, err)

		got, err = client.BitPos(ctx, "BitPos-1", 1)
		xt.NoError(t, err)
		xt.Equal(t, 1, got)
	})

	t.Run("GetBit", func(t *testing.T) {
		got, err := client.GetBit(ctx, "GetBit-1", 1)
		xt.NoError(t, err)
		xt.Equal(t, 0, got)

		err = client.Set(ctx, "GetBit-1", "demo")
		xt.NoError(t, err)

		got, err = client.GetBit(ctx, "GetBit-1", 1)
		xt.NoError(t, err)
		xt.Equal(t, 1, got)
	})

	t.Run("SetBit", func(t *testing.T) {
		got, err := client.SetBit(ctx, "SetBit-1", 1, 1)
		xt.NoError(t, err)
		xt.Equal(t, 0, got)

		got, err = client.GetBit(ctx, "SetBit-1", 1)
		xt.NoError(t, err)
		xt.Equal(t, 1, got)

		got, err = client.SetBit(ctx, "SetBit-1", 1, 0)
		xt.NoError(t, err)
		xt.Equal(t, 1, got)

		got, err = client.GetBit(ctx, "SetBit-1", 1)
		xt.NoError(t, err)
		xt.Equal(t, 0, got)
	})
}
