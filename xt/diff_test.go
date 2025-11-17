//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-17

package xt

import (
	"strings"
	"testing"
)

func Test_cutDiffAfter(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case 1",
			args: args{
				s: "hello world",
			},
			want: "hello world",
		},
		{
			name: "case 2",
			args: args{
				s: "hello\nworld\n",
			},
			want: "hello\nworld\n",
		},
		{
			name: "case 3",
			args: args{
				s: strings.Repeat("hello world\n", 35),
			},
			want: strings.TrimSpace(strings.Repeat("hello world\n", 31)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cutDiffAfter(tt.args.s); got != tt.want {
				t.Errorf("cutDiffAfter() = %v, want %v", got, tt.want)
			}
		})
	}
}
