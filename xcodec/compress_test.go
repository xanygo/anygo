//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-30

package xcodec

import (
	"testing"

	"github.com/fsgo/fst"
)

func TestGZipCompress(t *testing.T) {
	data1 := []byte("hello")
	got1, err1 := GZipCompress(data1)
	fst.NoError(t, err1)
	got2, err2 := GZipDecompress(got1)
	fst.NoError(t, err2)
	fst.Equal(t, string(got2), string(data1))
}

func TestZLibCompress(t *testing.T) {
	data1 := []byte("hello")
	got1, err1 := ZLibCompress(data1)
	fst.NoError(t, err1)
	got2, err2 := ZLibDecompress(got1)
	fst.NoError(t, err2)
	fst.Equal(t, string(got2), string(data1))
}

func TestFlateCompress(t *testing.T) {
	data1 := []byte("hello")
	got1, err1 := FlateCompress(data1)
	fst.NoError(t, err1)
	got2, err2 := FlateDecompress(got1)
	fst.NoError(t, err2)
	fst.Equal(t, string(got2), string(data1))
}
