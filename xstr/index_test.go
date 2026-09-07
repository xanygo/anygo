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
				n:      0,
			},
			want: 0,
		},
		{
			name: "case 2",
			args: args{
				s:      "abc/abc/abc",
				substr: "abc",
				n:      1,
			},
			want: 4,
		},
		{
			name: "case 3",
			args: args{
				s:      "abc/abc/abc",
				substr: "abc",
				n:      3,
			},
			want: -1,
		},
		{
			name: "case 4",
			args: args{
				s:      "abc/abc/abc",
				substr: "#",
				n:      0,
			},
			want: -1,
		},
		{
			name: "case 5",
			args: args{
				s:      "abc/abc/abc",
				substr: "#",
				n:      1,
			},
			want: -1,
		},
		{
			name: "case 6",
			args: args{
				s:      "abc/abc/abc",
				substr: "#",
				n:      -1,
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
				n:      0,
			},
			want: 8,
		},
		{
			name: "case 2",
			args: args{
				s:      "abc/abc/abc",
				substr: "abc",
				n:      1,
			},
			want: 4,
		},
		{
			name: "case 3",
			args: args{
				s:      "abc/abc/abc",
				substr: "#",
				n:      0,
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
				n:      -1,
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
				n: 0,
			},
			want: 0,
		},
		{
			name: "case 2",
			args: args{
				s: "abc/abc/abc",
				c: 'a',
				n: 1,
			},
			want: 4,
		},
		{
			name: "case 3",
			args: args{
				s: "abc/abc/abc",
				c: 'a',
				n: 3,
			},
			want: -1,
		},
		{
			name: "case 4",
			args: args{
				s: "abc/abc/abc",
				c: '#',
				n: 0,
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
				n: 0,
			},
			want: 8,
		},
		{
			name: "case 2",
			args: args{
				s: "abc/abc/abc",
				c: 'a',
				n: 1,
			},
			want: 4,
		},
		{
			name: "case 3",
			args: args{
				s: "abc/abc/abc",
				c: 'a',
				n: 3,
			},
			want: -1,
		},
		{
			name: "case 4",
			args: args{
				s: "abc/abc/abc",
				c: '#',
				n: 0,
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

func TestBytePairIndex(t *testing.T) {
	type args struct {
		str   string
		left  byte
		right byte
	}
	tests := []struct {
		name           string
		args           args
		wantLeftIndex  int
		wantRightIndex int
		wantOk         bool
	}{
		{
			name: "case 1",
			args: args{
				str:   "(hello(a,b,c,d(e,f))) word(a,b)",
				left:  '(',
				right: ')',
			},
			wantLeftIndex:  0,
			wantRightIndex: 20,
			wantOk:         true,
		},
		{
			name: "case 2",
			args: args{
				str:   "word(a,b)",
				left:  '(',
				right: ')',
			},
			wantLeftIndex:  4,
			wantRightIndex: 8,
			wantOk:         true,
		},
		{
			name: "case 3",
			args: args{
				str:   "word(a,b",
				left:  '(',
				right: ')',
			},
			wantLeftIndex:  4,
			wantRightIndex: -1,
			wantOk:         false,
		},
		{
			name: "case 4",
			args: args{
				str:   "word(a,b()",
				left:  '(',
				right: ')',
			},
			wantLeftIndex:  4,
			wantRightIndex: 9,
			wantOk:         false,
		},
		{
			name: "case 5",
			args: args{
				str:   "/}id:[0-9]+{/",
				left:  '{',
				right: '}',
			},
			wantLeftIndex:  -1,
			wantRightIndex: 1,
			wantOk:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLeftIndex, gotRightIndex, gotOk := BytePairIndex(tt.args.str, tt.args.left, tt.args.right)
			if gotLeftIndex != tt.wantLeftIndex {
				t.Errorf("BytePairIndex() gotLeftIndex = %v, want %v", gotLeftIndex, tt.wantLeftIndex)
			}
			if gotRightIndex != tt.wantRightIndex {
				t.Errorf("BytePairIndex() gotRightIndex = %v, want %v", gotRightIndex, tt.wantRightIndex)
			}
			if gotOk != tt.wantOk {
				t.Errorf("BytePairIndex() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
