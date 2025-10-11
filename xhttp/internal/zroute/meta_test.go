//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-24

package zroute

import (
	"reflect"
	"testing"
)

func Test_parserMeta(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name    string
		args    args
		want    Meta
		wantErr bool
	}{
		{
			name: "case 1",
			want: Meta{
				Other: map[string]string{},
			},
		},
		{
			name: "case 2",
			args: args{
				str: "id=1,k1=v1,,",
			},
			want: Meta{
				ID: "1",
				Other: map[string]string{
					"k1": "v1",
				},
			},
		},
		{
			name: "case 3",
			args: args{
				str: "id=1,k1=v1,,k2=",
			},
			want: Meta{
				ID: "1",
				Other: map[string]string{
					"k1": "v1",
					"k2": "",
				},
			},
		},
		{
			name: "case 4",
			args: args{
				str: "id=1,=",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parserMeta(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("parserMeta() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parserMeta() got = %v, want %v", got, tt.want)
			}
		})
	}
}
