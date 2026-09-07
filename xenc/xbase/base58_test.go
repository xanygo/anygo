//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-08-18

package xbase

import (
	"bytes"
	"reflect"
	"testing"
)

func TestBase58Codec_Encode(t *testing.T) {
	type args struct {
		input []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "case 1",
			args: args{
				input: []byte("hello world"),
			},
			want: []byte("StV1DL6CwTryKyV"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Base58Codec{}
			if got := b.Encode(tt.args.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBase58Codec_Decode(t *testing.T) {
	type args struct {
		input []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "case 1",
			args: args{
				input: []byte("StV1DL6CwTryKyV"),
			},
			want: []byte("hello world"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Base58Codec{}
			got, err := b.Decode(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("Decode() got = %q, want %q", got, tt.want)
			}
		})
	}
}
