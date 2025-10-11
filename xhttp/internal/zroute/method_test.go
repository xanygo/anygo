//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-15

package zroute

import (
	"net/http"
	"testing"
)

func TestGetPrefixMethod(t *testing.T) {
	tests := []struct {
		str  string
		want string
	}{
		{
			str:  "",
			want: http.MethodGet,
		},
		{
			str:  "Get",
			want: http.MethodGet,
		},
		{
			str:  "GetUser",
			want: http.MethodGet,
		},
		{
			str:  "GetUserList",
			want: http.MethodGet,
		},
		{
			str:  "DeleteByID",
			want: http.MethodDelete,
		},
		{
			str:  "Index",
			want: http.MethodGet,
		},
		{
			str:  "Search",
			want: http.MethodGet,
		},
		{
			str:  "Save",
			want: http.MethodPost,
		},
		{
			str:  "SaveByID",
			want: http.MethodPost,
		},
		{
			str:  "Update",
			want: http.MethodPut,
		},
		{
			str:  "UpdateByID",
			want: http.MethodPut,
		},
	}
	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			got := GetPrefixMethod(tt.str)
			if got != tt.want {
				t.Errorf("SplitCamelCase() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStripPrefixMethod(t *testing.T) {
	type args struct {
		s      string
		prefix string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case 1",
			args: args{
				s:      "Get",
				prefix: "Get",
			},
			want: "",
		},
		{
			name: "case 2",
			args: args{
				s:      "PostByID",
				prefix: "POST",
			},
			want: "ByID",
		},
		{
			name: "case 3",
			args: args{
				s:      "QueryByID",
				prefix: "POST",
			},
			want: "QueryByID",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripPrefixMethod(tt.args.s, tt.args.prefix); got != tt.want {
				t.Errorf("StripPrefixMethod() = %v, want %v", got, tt.want)
			}
		})
	}
}
