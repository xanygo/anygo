//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-21

package xstr

import "testing"

func TestIndexN(t *testing.T) {
	type args struct {
		s      string
		substr string
		n      int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "case 1",
			args: args{
				s:      "abc/abc/abc",
				substr: "abc",
				n:      1,
			},
			want: 0,
		},
		{
			name: "case 2",
			args: args{
				s:      "abc/abc/abc",
				substr: "abc",
				n:      2,
			},
			want: 4,
		},
		{
			name: "case 3",
			args: args{
				s:      "abc/abc/abc",
				substr: "abc",
				n:      4,
			},
			want: -1,
		},
		{
			name: "case 4",
			args: args{
				s:      "abc/abc/abc",
				substr: "#",
				n:      1,
			},
			want: -1,
		},
		{
			name: "case 5",
			args: args{
				s:      "abc/abc/abc",
				substr: "#",
				n:      2,
			},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IndexN(tt.args.s, tt.args.substr, tt.args.n); got != tt.want {
				t.Errorf("IndexN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLastIndexN(t *testing.T) {
	type args struct {
		s      string
		substr string
		n      int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "case 1",
			args: args{
				s:      "abc/abc/abc",
				substr: "abc",
				n:      1,
			},
			want: 8,
		},
		{
			name: "case 2",
			args: args{
				s:      "abc/abc/abc",
				substr: "abc",
				n:      2,
			},
			want: 4,
		},
		{
			name: "case 3",
			args: args{
				s:      "abc/abc/abc",
				substr: "#",
				n:      1,
			},
			want: -1,
		},
		{
			name: "case 4",
			args: args{
				s:      "abc/abc/abc",
				substr: "#",
				n:      2,
			},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LastIndexN(tt.args.s, tt.args.substr, tt.args.n); got != tt.want {
				t.Errorf("LastIndexN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIndexByteN(t *testing.T) {
	type args struct {
		s string
		c byte
		n int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "case 1",
			args: args{
				s: "abc/abc/abc",
				c: 'a',
				n: 1,
			},
			want: 0,
		},
		{
			name: "case 2",
			args: args{
				s: "abc/abc/abc",
				c: 'a',
				n: 2,
			},
			want: 4,
		},
		{
			name: "case 3",
			args: args{
				s: "abc/abc/abc",
				c: 'a',
				n: 4,
			},
			want: -1,
		},
		{
			name: "case 4",
			args: args{
				s: "abc/abc/abc",
				c: '#',
				n: 1,
			},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IndexByteN(tt.args.s, tt.args.c, tt.args.n); got != tt.want {
				t.Errorf("IndexByteN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLastIndexByteN(t *testing.T) {
	type args struct {
		s string
		c byte
		n int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "case 1",
			args: args{
				s: "abc/abc/abc",
				c: 'a',
				n: 1,
			},
			want: 8,
		},
		{
			name: "case 2",
			args: args{
				s: "abc/abc/abc",
				c: 'a',
				n: 2,
			},
			want: 4,
		},
		{
			name: "case 3",
			args: args{
				s: "abc/abc/abc",
				c: 'a',
				n: 4,
			},
			want: -1,
		},
		{
			name: "case 4",
			args: args{
				s: "abc/abc/abc",
				c: '#',
				n: 1,
			},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LastIndexByteN(tt.args.s, tt.args.c, tt.args.n); got != tt.want {
				t.Errorf("LastIndexByteN() = %v, want %v", got, tt.want)
			}
		})
	}
}
