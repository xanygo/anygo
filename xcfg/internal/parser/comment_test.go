//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package parser

import (
	"reflect"
	"testing"
)

func TestStripComment(t *testing.T) {
	type args struct {
		input []byte
	}
	tests := []struct {
		name    string
		args    args
		wantOut []byte
	}{
		{
			name: "case 1",
			args: args{
				input: []byte(`line1
#line2
  #line3
#line4

line6 #666`),
			},
			wantOut: []byte(`line1




line6 #666`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOut := StripComment(tt.args.input); !reflect.DeepEqual(gotOut, tt.wantOut) {
				t.Errorf("StripComment() = %q, want %q", gotOut, tt.wantOut)
			}
		})
	}
}

func TestHeadComments(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "case 1",
			args: args{
				input: "#a\n# b\n \n\n ### c\nhello#d",
			},
			want: []string{
				"a",
				"b",
				"c",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HeadComments([]byte(tt.args.input)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HeadComments() = %v, want %v", got, tt.want)
			}
		})
	}
}
