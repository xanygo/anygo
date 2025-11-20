//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-20

package dialect

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func Test_pgAnyArrayCodec_Decode(t *testing.T) {
	type args struct {
		b string
		a any
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{
			name: "case 1",
			args: args{
				b: "{1,2,3}",
				a: []int{},
			},
			want: []int{1, 2, 3},
		},
		{
			name: "case 2",
			args: args{
				b: `{{"a{b}","x,y"},{"hello","w\"orld"}}`,
				a: [][]string{},
			},
			want: [][]string{
				{"a{b}", "x,y"},
				{"hello", `w\"orld`},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := pgAnyArrayCodec{}
			err := p.Decode(tt.args.b, &tt.args.a)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				xt.Equal(t, tt.want, tt.args.a)
			}
		})
	}
}
