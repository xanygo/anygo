//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-09

package xi18n

import (
	"bytes"
	"context"
	"testing"
	"text/template"

	"github.com/xanygo/anygo"
	"github.com/xanygo/anygo/xt"
)

func TestTemplate(t *testing.T) {
	type args struct {
		langs []Language
		ns    string
	}
	tests := []struct {
		name    string
		tpl     string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "case 1",
			tpl:  `hello {{xi "index@k1"}}`,
			want: "hello 你好",
		},
		{
			name: "case 2",
			args: args{
				langs: []Language{
					LangEn,
				},
			},
			tpl:  `hello {{"index@k1" | xi}}`,
			want: "hello hello",
		},
		{
			name: "case 3",
			args: args{
				langs: []Language{
					LangEn,
					LangZh,
				},
			},
			tpl:  `hello {{"index@k1" | xi}}`,
			want: "hello hello",
		},
		{
			name: "case 4",
			tpl:  `hello {{"index@k1" | xi}}`,
			want: "hello 你好",
		},
		{
			name: "case 5",
			tpl:  `hello {{ xit "index@k1" "你好"}}`,
			want: "hello 你好",
		},
		{
			name: "case 6",
			tpl:  `hello {{ "你好" | xit "index@k1"}}`,
			want: "hello 你好",
		},
		{
			name: "case 7",
			args: args{
				langs: []Language{
					LangEn,
				},
			},
			tpl:  `hello {{ "你好" |xit "index@k1"}}`,
			want: "hello hello",
		},
		{
			name: "case 8",
			tpl:  `hello {{ "你好 {0}" | xit "index@k2" "demo"}}`,
			want: "hello 你好 demo",
		},
		{
			name: "case 9",
			args: args{
				langs: []Language{
					LangEn,
				},
			},
			tpl:  `hello {{ "你好 {0}" | xit "index@k2" "demo"}}`,
			want: "hello hello demo",
		},
		{
			name: "case 10",
			tpl:  `hello {{xi "index@k2" "demo"}}`,
			want: "hello 你好 demo",
		},
		{
			name:    "case 11",
			tpl:     `hello {{xi "index@k_error"}}`, // key 不存在
			wantErr: true,
		},
		{
			name: "case 12",
			tpl:  `hello {{ "你好 {0}" | xit "index@k_error" "demo"}}`,
			want: "hello 你好 demo",
		},
		{
			name:    "case 13",
			tpl:     `hello {{ xit "index@k_error"}}`, // key 不存在
			wantErr: true,
		},
	}

	b := &Bundle{}
	e0 := b.MustLocalize(LangZh).Add("index", &Message{Key: "k1", Other: "你好"})
	xt.NoError(t, e0)

	e1 := b.MustLocalize(LangEn).Add("index", &Message{Key: "k1", Other: "hello"})
	xt.NoError(t, e1)

	e2 := b.MustLocalize(LangZh).Add("index", &Message{Key: "k2", Other: "你好 {0}"})
	xt.NoError(t, e2)

	e3 := b.MustLocalize(LangEn).Add("index", &Message{Key: "k2", Other: "hello {0}"})
	xt.NoError(t, e3)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := template.New("test").Funcs(FuncMap(b, tt.args.langs, tt.args.ns)).Parse(tt.tpl)
			xt.NoError(t, err)
			bf := &bytes.Buffer{}
			err = tpl.Execute(bf, nil)
			if tt.wantErr {
				xt.Error(t, err)
				return
			}
			xt.NoError(t, err)
			xt.Equal(t, bf.String(), tt.want)
		})
	}
}

func TestRA(t *testing.T) {
	b := &Bundle{}
	b.MustLocalize(LangZh).MustAdd("index", &Message{Key: "k1", Other: "你好"})
	b.MustLocalize(LangEn).MustAdd("index", &Message{Key: "k1", Other: "hello"})

	b.MustLocalize(LangZh).MustAdd("index", &Message{Key: "k2", Other: "你好 {0}"})
	b.MustLocalize(LangEn).MustAdd("index", &Message{Key: "k2", Other: "hello {0}"})

	_, err0 := RA(context.Background(), "index@k1")
	xt.Error(t, err0)

	ctx1 := ContextWithBundle(context.Background(), b, "")
	xt.Equal(t, anygo.Must1(RA(ctx1, "index@k1")), "你好")
	xt.Equal(t, anygo.Must1(RA(ctx1, "index@k2", "demo")), "你好 demo")

	t.Run("RB", func(t *testing.T) {
		xt.Equal(t, anygo.Must1(RB(ctx1, "abc", "index@k1")), "你好")

		ctx2 := ContextWithLanguages(ctx1, []Language{"jp"})
		xt.Equal(t, anygo.Must1(RB(ctx2, "abc", "index@k1")), "abc")
		xt.Equal(t, anygo.Must1(RB(ctx2, "abc {0}", "index@k1", "demo")), "abc demo")
	})

	ctx3 := ContextWithLanguages(ctx1, []Language{LangEn})
	xt.Equal(t, anygo.Must1(RA(ctx3, "index@k1")), "hello")
}
