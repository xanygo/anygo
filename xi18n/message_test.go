//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-07

package xi18n

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

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

func TestMessage_Render(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		m := &Message{
			Key:   "demo",
			Other: "hello world",
		}
		xt.NoError(t, m.initAndCheck())
		got, err := m.Render()
		xt.NoError(t, err)
		xt.Equal(t, got, "hello world")

		got, err = m.Render("abc")
		xt.Error(t, err)
		xt.Empty(t, got)
	})
	t.Run("case 2", func(t *testing.T) {
		m := &Message{
			Key:   "demo",
			One:   "只有 {0} 个",
			Two:   "共有 {0} 个",
			Other: "有 {0} 个",
		}
		xt.NoError(t, m.initAndCheck())
		got, err := m.Render()
		xt.Error(t, err)
		xt.Equal(t, got, "")

		got, err = m.Render(0)
		xt.NoError(t, err)
		xt.Equal(t, got, "有 0 个")

		got, err = m.Render(1)
		xt.NoError(t, err)
		xt.Equal(t, got, "只有 1 个")

		got, err = m.Render(2)
		xt.NoError(t, err)
		xt.Equal(t, got, "共有 2 个")
	})

	t.Run("case 3", func(t *testing.T) {
		m := &Message{
			Key:   "demo",
			Zero:  "zero books",
			One:   "one book",
			Other: "{0} books",
		}
		xt.NoError(t, m.initAndCheck())
		got, err := m.Render()
		xt.Error(t, err)
		xt.Equal(t, got, "")

		got, err = m.Render(0)
		xt.NoError(t, err)
		xt.Equal(t, got, "zero books")

		got, err = m.Render(1)
		xt.NoError(t, err)
		xt.Equal(t, got, "one book")

		got, err = m.Render(2)
		xt.NoError(t, err)
		xt.Equal(t, got, "2 books")
	})
}
