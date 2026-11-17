//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-21

package xstr

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestCutIndex(t *testing.T) {
	type args struct {
		s     string
		index int
		len   int
	}
	tests := []struct {
		name       string
		args       args
		wantBefore string
		wantAfter  string
		wantFound  bool
	}{
		{
			name: "case 1",
			args: args{
				s:     "hello",
				index: -1,
			},
			wantBefore: "",
			wantAfter:  "hello",
			wantFound:  false,
		},
		{
			name: "case 2",
			args: args{
				s:     "hello",
				index: 5,
			},
			wantBefore: "hello",
			wantAfter:  "",
			wantFound:  true,
		},
		{
			name: "case 3",
			args: args{
				s:     "hello",
				index: 1,
			},
			wantBefore: "h",
			wantAfter:  "ello",
			wantFound:  true,
		},
		{
			name: "case 4",
			args: args{
				s:     "hello",
				index: 10,
			},
			wantBefore: "hello",
			wantAfter:  "",
		},
		{
			name: "case 5",
			args: args{
				s:     "hello",
				index: 1,
				len:   2,
			},
			wantBefore: "h",
			wantAfter:  "lo",
			wantFound:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBefore, gotAfter, gotFound := CutIndex(tt.args.s, tt.args.index, tt.args.len)
			if gotFound != tt.wantFound {
				t.Errorf("CutIndex() gotFound = %v, want %v", gotFound, tt.wantFound)
			}
			if gotBefore != tt.wantBefore {
				t.Errorf("CutIndex() gotBefore = %v, want %v", gotBefore, tt.wantBefore)
			}
			if gotAfter != tt.wantAfter {
				t.Errorf("CutIndex() gotAfter = %v, want %v", gotAfter, tt.wantAfter)
			}
		})
	}
}

func TestSubstr(t *testing.T) {
	xt.Equal(t, Substr("hello", 0, 2), "he")
	xt.Equal(t, Substr("hello", 0, 5), "hello")
	xt.Equal(t, Substr("hello", 0, 6), "hello")
	xt.Equal(t, Substr("hello", -1, 1), "o")
	xt.Equal(t, Substr("hello", -2, 2), "lo")
	xt.Equal(t, Substr("hello", -5, 2), "he")
	xt.Equal(t, Substr("hello", -10, 5), "hello")
	xt.Equal(t, Substr("hello", -10, 4), "hell")
	xt.Equal(t, Substr("hello", 5, 4), "")
	xt.Equal(t, Substr("hello", 4, 4), "o")
	xt.Equal(t, Substr("hello", 10, 4), "")
}

func TestCutLastNAfter(t *testing.T) {
	type args struct {
		s      string
		substr string
		n      int
	}
	tests := []struct {
		name      string
		args      args
		wantAfter string
	}{
		{
			name: "case 1",
			args: args{
				s:      "abc/hello/world",
				substr: "/",
				n:      1,
			},
			wantAfter: "hello/world",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotAfter := CutLastNAfter(tt.args.s, tt.args.substr, tt.args.n); gotAfter != tt.wantAfter {
				t.Errorf("CutLastNAfter() = %v, want %v", gotAfter, tt.wantAfter)
			}
		})
	}
}

func TestCutLastNBefore(t *testing.T) {
	type args struct {
		s      string
		substr string
		n      int
	}
	tests := []struct {
		name       string
		args       args
		wantBefore string
	}{
		{
			name: "case 1",
			args: args{
				s:      "abc/hello/world",
				substr: "/",
				n:      1,
			},
			wantBefore: "abc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotBefore := CutLastNBefore(tt.args.s, tt.args.substr, tt.args.n); gotBefore != tt.wantBefore {
				t.Errorf("CutLastNBefore() = %v, want %v", gotBefore, tt.wantBefore)
			}
		})
	}
}

func TestSplitLen(t *testing.T) {
	xt.Equal(t, SplitLen("hello", 2), []string{"he", "ll", "o"})
}
