//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-23

package zroute

import "testing"

func TestCleanPattern(t *testing.T) {
	type args struct {
		p string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case 1",
			args: args{
				p: "/",
			},
			want: "/",
		},
		{
			name: "case 2",
			args: args{
				p: "/user/*",
			},
			want: "/user/*",
		},
		{
			name: "case 3",
			args: args{
				p: "/user/{name}",
			},
			want: "/user/{name}",
		},
		{
			name: "case 4",
			args: args{
				p: "/user/{age:[1-9]+}",
			},
			want: "/user/{age}",
		},
		{
			name: "case 5",
			args: args{
				p: "/user/{age:*}",
			},
			want: "/user/{age}",
		},
		{
			name: "case 6",
			args: args{
				p: "/user/{class}/{age:[1-9]+}/{name:[.*]}",
			},
			want: "/user/{class}/{age}/{name}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CleanPattern(tt.args.p); got != tt.want {
				t.Errorf("CleanPattern() = %v, want %v", got, tt.want)
			}
		})
	}
}
