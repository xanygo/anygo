//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-15

package zroute

import "testing"

func TestSplitCamelCase(t *testing.T) {
	tests := []struct {
		str   string
		want  string
		want1 string
	}{
		{
			str:   "Get",
			want:  "Get",
			want1: "",
		},
		{
			str:   "GetUser",
			want:  "Get",
			want1: "User",
		},
		{
			str:   "GetUserList",
			want:  "Get",
			want1: "UserList",
		},
	}
	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			got, got1 := SplitCamelCase(tt.str)
			if got != tt.want {
				t.Errorf("SplitCamelCase() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("SplitCamelCase() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
