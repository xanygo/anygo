//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-30

package xcodec

import (
	"strconv"
	"strings"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestAesCBC_Encrypt(t *testing.T) {
	ac := &AesBlock{
		Key: "hello",
	}
	for i := 0; i < 32; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			txt := []byte(strings.Repeat("hello", i))
			t.Logf("txt= %q, len=%d", txt, len(txt))
			got1, err1 := ac.Encrypt(txt)
			xt.NoError(t, err1)
			t.Logf("got1= %q, len=%d", got1, len(got1))

			got2, err2 := ac.Decrypt(got1)
			xt.NoError(t, err2)
			t.Logf("got2= %q , len=%d", got2, len(got2))
			xt.Equal(t, string(txt), string(got2))
		})
	}
}

func TestAesOFB_Encrypt(t *testing.T) {
	ac := &AesOFB{
		Key: "hello",
	}
	for i := 0; i < 32; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			txt := []byte(strings.Repeat("hello", i))
			t.Logf("txt= %q, len=%d", txt, len(txt))
			got1, err1 := ac.Encrypt(txt)
			xt.NoError(t, err1)
			t.Logf("got1= %q, len=%d", got1, len(got1))

			got2, err2 := ac.Decrypt(got1)
			xt.NoError(t, err2)
			t.Logf("got2= %q , len=%d", got2, len(got2))
			xt.Equal(t, string(txt), string(got2))
		})
	}
}

func BenchmarkAES(b *testing.B) {
	b.Run("Encrypt", func(b *testing.B) {
		b.Run("AesBlock", func(b *testing.B) {
			ac := &AesBlock{
				Key: "hello",
			}
			for i := 0; i < b.N; i++ {
				_, _ = ac.Encrypt([]byte("hello"))
			}
		})

		b.Run("AesOFB", func(b *testing.B) {
			ac := &AesOFB{
				Key: "hello",
			}
			for i := 0; i < b.N; i++ {
				_, _ = ac.Encrypt([]byte("hello"))
			}
		})
	})

	b.Run("Decrypt", func(b *testing.B) {
		ac0 := &AesBlock{
			Key: "hello",
		}
		data0, err0 := ac0.Encrypt([]byte("hello"))
		xt.NoError(b, err0)
		b.Run("AesBlock", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = ac0.Decrypt(data0)
			}
		})

		ac1 := &AesOFB{
			Key: "hello",
		}
		data1, err1 := ac0.Encrypt([]byte("hello"))
		xt.NoError(b, err1)
		b.Run("AesOFB", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = ac1.Decrypt(data1)
			}
		})
	})
}

func TestCiphers(t *testing.T) {
	cs := Ciphers{
		&AesBlock{
			Key: "demo",
		},
		&Base64{},
		&AesOFB{
			Key: "hello",
		},
	}
	txt := []byte("hello")
	got1, err1 := cs.Encrypt(txt)
	xt.NoError(t, err1)
	got2, err2 := cs.Decrypt(got1)
	xt.NoError(t, err2)
	xt.Equal(t, string(txt), string(got2))
}
