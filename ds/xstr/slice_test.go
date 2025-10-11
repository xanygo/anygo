//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-24

package xstr

import (
	"reflect"
	"testing"
)

func TestToInts(t *testing.T) {
	type args struct {
		str string
		sep string
	}
	tests := []struct {
		name    string
		args    args
		want    []int
		wantErr bool
	}{
		{
			name: "case 1",
			args: args{
				str: "123",
				sep: ",",
			},
			want: []int{123},
		},
		{
			name: "case 2",
			args: args{
				str: "123,,234,,",
				sep: ",",
			},
			want: []int{123, 234},
		},
		{
			name: "case 3",
			args: args{
				str: "123,,234,,abc",
				sep: ",",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToInts(tt.args.str, tt.args.sep)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToInts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToInts() got = %v, want %v", got, tt.want)
			}
		})
	}
}
