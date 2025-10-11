//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-18

package xstr

import "testing"

func TestHasAnyPrefix(t *testing.T) {
	type args struct {
		str    string
		prefix []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "case 1",
			args: args{
				str: "name.js",
			},
			want: false,
		},
		{
			name: "case 2",
			args: args{
				str:    "name.js",
				prefix: []string{"abc"},
			},
			want: false,
		},
		{
			name: "case 3",
			args: args{
				str:    "name.js",
				prefix: []string{"na", "hello"},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasAnyPrefix(tt.args.str, tt.args.prefix...); got != tt.want {
				t.Errorf("HasAnyPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}
