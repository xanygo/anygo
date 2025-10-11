//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-07

package xi18n

import (
	"reflect"
	"testing"
)

func TestParserAccept(t *testing.T) {
	type args struct {
		accept string
	}
	tests := []struct {
		name string
		args args
		want []Language
	}{
		{
			name: "case 1",
			args: args{
				accept: "zh-CN,zh;q=0.9,en;q=0.8",
			},
			want: []Language{
				Language("zh-CN"),
				Language("zh"),
				Language("en"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParserAccept(tt.args.accept); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParserAccept() = %v, want %v", got, tt.want)
			}
		})
	}
}
