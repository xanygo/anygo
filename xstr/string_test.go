//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-04

package xstr

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestToSnakeCase(t *testing.T) {
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
				s: "SpaceID",
			},
			want: "space_id",
		},
		{
			name: "case 2",
			args: args{
				s: "id",
			},
			want: "id",
		},
		{
			name: "case 3",
			args: args{
				s: "UserName",
			},
			want: "user_name",
		},
		{
			name: "case 4",
			args: args{
				s: "userName",
			},
			want: "user_name",
		},
		{
			name: "case 5",
			args: args{
				s: "XMLHTTPRequest",
			},
			want: "xmlhttp_request",
		},
		{
			name: "case 6",
			args: args{
				s: "MyAPIV2Test",
			},
			want: "my_apiv2_test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSnakeCase(tt.args.s); got != tt.want {
				t.Errorf("ToSnakeCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatch(t *testing.T) {
	type args struct {
		pattern string
		str     string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "case 1",
			args: args{
				pattern: "wc:hello *",
				str:     "hello world",
			},
			want: true,
		},
		{
			name: "case 2",
			args: args{
				pattern: "wc:hello *",
				str:     "hello",
			},
			want: false,
		},
		{
			name: "case 3",
			args: args{
				pattern: `re:\d+`,
				str:     "hello",
			},
			want: false,
		},
		{
			name: "case 4",
			args: args{
				pattern: `re:\d+`,
				str:     "123",
			},
			want: true,
		},
		{
			name: "case 5",
			args: args{
				pattern: `hello`,
				str:     "123",
			},
			want: false,
		},
		{
			name: "case 6",
			args: args{
				pattern: `hello`,
				str:     "hello",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Match(tt.args.pattern, tt.args.str); got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompileFilters(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		fn, err := CompileFilters(nil, nil)
		xt.NoError(t, err)
		xt.True(t, fn("ok"))
	})
	t.Run("case 2", func(t *testing.T) {
		fn, err := CompileFilters([]string{"wc:*.css", "wc:*.js"}, nil)
		xt.NoError(t, err)
		xt.True(t, fn("ok.css"))
		xt.True(t, fn("ok.js"))
		xt.True(t, fn("tmp/ok.js"))

		xt.False(t, fn("ok.png"))
		xt.False(t, fn("ok.js.png"))
	})
	t.Run("case 3", func(t *testing.T) {
		fn, err := CompileFilters([]string{"wc:*.css", "wc:*.js"}, []string{"wc:tmp/*"})
		xt.NoError(t, err)
		xt.True(t, fn("ok.css"))
		xt.True(t, fn("ok.js"))
		xt.False(t, fn("ok.png"))
		xt.False(t, fn("tmp/ok.js"))
	})
}
