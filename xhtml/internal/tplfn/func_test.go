//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-12-07

package tplfn

import "testing"

func TestInputObjectName(t *testing.T) {
	type args struct {
		values []any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case 1",
			args: args{},
			want: "",
		},
		{
			name: "case 2",
			args: args{
				values: []any{nil, "name"},
			},
			want: "name",
		},
		{
			name: "case 3",
			args: args{
				values: []any{"widget", "name"},
			},
			want: "widget[name]",
		},
		{
			name: "case 3",
			args: args{
				values: []any{"widget", "name", 1},
			},
			want: "widget[name][1]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InputObjectName(tt.args.values...); got != tt.want {
				t.Errorf("xObjName() = %v, want %v", got, tt.want)
			}
		})
	}
}
