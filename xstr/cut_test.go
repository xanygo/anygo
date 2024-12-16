//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-21

package xstr

import "testing"

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
			wantBefore: "hello",
			wantAfter:  "",
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
	type args struct {
		s      string
		start  int
		length int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case 1",
			args: args{
				s:      "hello",
				start:  0,
				length: 2,
			},
			want: "he",
		},
		{
			name: "case 2",
			args: args{
				s:      "hello",
				start:  0,
				length: 5,
			},
			want: "hello",
		},
		{
			name: "case 3",
			args: args{
				s:      "hello",
				start:  0,
				length: 6,
			},
			want: "hello",
		},
		{
			name: "case 4",
			args: args{
				s:      "hello",
				start:  -1,
				length: 1,
			},
			want: "o",
		},
		{
			name: "case 5",
			args: args{
				s:      "hello",
				start:  -2,
				length: 1,
			},
			want: "l",
		},
		{
			name: "case 6",
			args: args{
				s:      "hello",
				start:  -2,
				length: 2,
			},
			want: "lo",
		},
		{
			name: "case 7",
			args: args{
				s:      "hello",
				start:  -5,
				length: 2,
			},
			want: "he",
		},
		{
			name: "case 8",
			args: args{
				s:      "hello",
				start:  -10,
				length: 5,
			},
			want: "hello",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Substr(tt.args.s, tt.args.start, tt.args.length); got != tt.want {
				t.Errorf("Substr() = %v, want %v", got, tt.want)
			}
		})
	}
}
