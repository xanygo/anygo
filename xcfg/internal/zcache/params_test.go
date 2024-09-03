//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package zcache

import (
	"reflect"
	"testing"
	"time"
)

func TestParserParam(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name    string
		args    args
		want    *Param
		wantErr bool
	}{
		{
			name: "case 1 empty",
			want: &Param{},
		},
		{
			name: "case 2",
			args: args{
				str: "timeout=1h&cache=10s",
			},
			want: &Param{
				Timeout: time.Hour,
				TTL:     10 * time.Second,
			},
		},
		{
			name: "case 3",
			args: args{
				str: "timeout=abc&cache=10s",
			},
			wantErr: true,
		},
		{
			name: "case 4",
			args: args{
				str: "timeout=1h&cache=abc",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParserParam(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParserParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParserParam() got = %v, want %v", got, tt.want)
			}
		})
	}
}
