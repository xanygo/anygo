//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-13

package xtype_test

import (
	"encoding/json"
	"testing"

	"github.com/xanygo/anygo/ds/xtype"
	"github.com/xanygo/anygo/xt"
)

func TestByteCount_MarshalText(t *testing.T) {
	m := map[string]any{
		"A": xtype.ByteCount(100),
	}
	bf, err := json.Marshal(m)
	xt.NoError(t, err)
	xt.Equal(t, `{"A":100}`, string(bf))
}

func TestByteCount_String(t *testing.T) {
	tests := []struct {
		name string
		d    xtype.ByteCount
		want string
	}{
		{
			name: "case 1",
			d:    0,
			want: "0B",
		},
		{
			name: "case 2",
			d:    100,
			want: "100B",
		},
		{
			name: "case 3",
			d:    1024,
			want: "1KiB",
		},
		{
			name: "case 4",
			d:    1024 * 1024 * 1024,
			want: "1GiB",
		},
		{
			name: "case 5",
			d:    1024 * 1024 * 1024 * 1024,
			want: "1TiB",
		},
		{
			name: "case 6",
			d:    1024 * 1024 * 1024 * 1024 * 1024,
			want: "1PiB",
		},
		{
			name: "case 7",
			d:    1024 * 1024 * 1024 * 1024 * 1024 * 1024,
			want: "1EiB",
		},
		{
			name: "case 8",
			d:    1024*1024*1024*1024*1024*1024 + 1024*1024*1024*1024*1024 + 234*1024*1024*1024*1024*1024 + 1024*1024*1024 + 1024*1024 + 1024 + 123,
			want: "1EiB 235PiB 1GiB 1MiB 1KiB 123B",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestByteCount_Parser(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    xtype.ByteCount
		wantErr bool
	}{
		{
			name: "case 1",
			args: args{
				s: "0B",
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "case 2",
			args: args{
				s: "1EiB 235PiB 1GiB 1MiB 1KiB 123B",
			},
			want:    1024*1024*1024*1024*1024*1024 + 1024*1024*1024*1024*1024 + 234*1024*1024*1024*1024*1024 + 1024*1024*1024 + 1024*1024 + 1024 + 123,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := xtype.ParserByteCount(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser() error = %v, wantErr %v", err, tt.wantErr)
			}
			if d != tt.want {
				t.Errorf("Parser() got = %d, want %d", d, tt.want)
			}
		})
	}
}

func TestByteCount_UnmarshalText(t *testing.T) {
	type memory struct {
		Size xtype.ByteCount
	}
	tests := []struct {
		name    string
		arg     string
		wantErr bool
		want    memory
	}{
		{
			name: "case 1",
			arg:  `{"Size":100}`,
			want: memory{Size: 100},
		},
		{
			name: "case 2",
			arg:  `{"Size":"100"}`,
			want: memory{Size: 100},
		},
		{
			name: "case 3",
			arg:  `{"Size":"1EiB 23PiB8B"}`,
			want: memory{Size: 1024*1024*1024*1024*1024*1024 + 23*1024*1024*1024*1024*1024 + 8},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m1 memory
			err := json.Unmarshal([]byte(tt.arg), &m1)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
			}
			xt.Equal(t, tt.want, m1)
		})
	}
}
