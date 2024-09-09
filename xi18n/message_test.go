//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-07

package xi18n

import "testing"

func TestMessage_plural(t *testing.T) {
	type args struct {
		data []any
	}
	tests := []struct {
		name string
		args args
		want pluralRule
	}{
		{
			name: "case 1",
			args: args{},
			want: pluralOther,
		},
		{
			name: "case 2",
			args: args{
				data: []any{
					0,
					"v2",
				},
			},
			want: pluralZero,
		},
		{
			name: "case 3",
			args: args{
				data: []any{
					uint8(1),
					"v2",
				},
			},
			want: pluralOne,
		},
		{
			name: "case 4",
			args: args{
				data: []any{
					float64(2),
					"v2",
				},
			},
			want: pluralTwo,
		},
		{
			name: "case 5",
			args: args{
				data: []any{
					float32(4),
					"v2",
				},
			},
			want: pluralFew,
		},
		{
			name: "case 6",
			args: args{
				data: []any{
					float32(10),
					"v2",
					1,
				},
			},
			want: pluralMany,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{}
			if got := m.plural(tt.args.data...); got != tt.want {
				t.Errorf("plural() = %v, want %v", got, tt.want)
			}
		})
	}
}
