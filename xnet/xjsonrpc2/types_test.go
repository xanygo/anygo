//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-03-30

package xjsonrpc2

import (
	"reflect"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestInt64ID_Bytes(t *testing.T) {
	id := Int64ID(100)
	xt.Equal(t, "100", string(id.Bytes()))
}

func TestStringID_Bytes(t *testing.T) {
	id := StringID("hello")
	xt.Equal(t, `"hello"`, string(id.Bytes()))
}

func Test_parserID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ID
		wantErr bool
	}{
		{
			name:  "case 1 string",
			input: `"hello"`,
			want:  StringID("hello"),
		},
		{
			name:  "case 2 int",
			input: `123`,
			want:  Int64ID(123),
		},
		{
			name:    "case 3 error",
			input:   `hello`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parserID([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("parserID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parserID() got = %v, want %v", got, tt.want)
			}
		})
	}
}
