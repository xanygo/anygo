//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-11

package zroute

import (
	"reflect"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func Test_splitPattern(t *testing.T) {
	type args struct {
		pattern string
	}
	tests := []struct {
		name  string
		args  args
		want  []string
		want1 string
		want2 string
	}{
		{
			name: "case 1",
			args: args{
				pattern: "/index",
			},
			want:  []string{"ANY"},
			want1: "/index",
		},
		{
			name: "case 2",
			args: args{
				pattern: "GET /index",
			},
			want:  []string{"GET"},
			want1: "/index",
		},
		{
			name: "case 3",
			args: args{
				pattern: "GET,POST /index",
			},
			want:  []string{"GET", "POST"},
			want1: "/index",
		},
		{
			name: "case 4",
			args: args{
				pattern: "",
			},
			want:  nil,
			want1: "",
		},
		{
			name: "case 5",
			args: args{
				pattern: "GET,POST /index meta|id=1,a=2",
			},
			want:  []string{"GET", "POST"},
			want1: "/index",
			want2: "id=1,a=2",
		},
		{
			name: "case 6",
			args: args{
				pattern: "/index meta|id=1,a=2",
			},
			want:  []string{"ANY"},
			want1: "/index",
			want2: "id=1,a=2",
		},
		{
			name: "case 7",
			args: args{
				pattern: `/{id} meta|id=1,a=2`,
			},
			want:  []string{"ANY"},
			want1: "/{id}",
			want2: "id=1,a=2",
		},
		{
			name: "case 8",
			args: args{
				pattern: `/{id:[0-9]+} meta|id=1,a=2`,
			},
			want:  []string{"ANY"},
			want1: "/{id:[0-9]+}",
			want2: "id=1,a=2",
		},
		{
			name: "case 9",
			args: args{
				pattern: `/{id:\d{3}} meta|id=1,a=2`,
			},
			want:  []string{"ANY"},
			want1: "/{id:\\d{3}}",
			want2: "id=1,a=2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := splitPattern(tt.args.pattern)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitPattern() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("splitPattern() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("splitPattern() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func Test_parserWordNode(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name    string
		args    args
		want    *wordNode
		wantErr bool
	}{
		{
			name: "case 1",
			args: args{
				str: "index",
			},
			want: &wordNode{
				Prefix: "index",
			},
		},
		{
			name: "case 2",
			args: args{
				str: "{name}",
			},
			want: &wordNode{
				Name: "name",
			},
		},
		{
			name: "case 3",
			args: args{
				str: "{name}.html",
			},
			want: &wordNode{
				Name:   "name",
				Suffix: ".html",
			},
		},
		{
			name: "case 4",
			args: args{
				str: "id-{name}",
			},
			want: &wordNode{
				Name:   "name",
				Prefix: "id-",
			},
		},
		{
			name: "case 5",
			args: args{
				str: "id-{name}.html",
			},
			want: &wordNode{
				Name:   "name",
				Prefix: "id-",
				Suffix: ".html",
			},
		},
		{
			name: "case 6",
			args: args{
				str: "}{",
			},
			wantErr: true,
		},
		{
			name: "case 7",
			args: args{
				str: "{{name}}",
			},
			wantErr: true,
		},
		{
			name: "case 8",
			args: args{
				str: "{name",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parserWordNode(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("parserWordNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parserWordNode() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parserRegexpPattern(t *testing.T) {
	type args struct {
		pattern string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "case 1",
			args: args{
				pattern: "/*",
			},
			want: "/(?P<p0>.*)",
		},
		{
			name: "case 2",
			args: args{
				pattern: "/*.html",
			},
			want: `/(?P<p0>.*)\.html`,
		},
		{
			name: "case 3",
			args: args{
				pattern: "/{id:[0-9]+}",
			},
			want: `/(?P<id>[0-9]+)`,
		},
		{
			name: "case 4",
			args: args{
				pattern: `/{id:[0-9]+}/{name:\w+}`,
			},
			want: `/(?P<id>[0-9]+)/(?P<name>\w+)`,
		},
		{
			name: "case 5",
			args: args{
				pattern: `/{id:[0-9]+}/*.html`,
			},
			want: `/(?P<id>[0-9]+)/(?P<p1>.*)\.html`,
		},
		{
			name: "case 6",
			args: args{
				pattern: `/{id:[0-9]+}/{`, // 多了一个 {
			},
			wantErr: true,
		},
		{
			name: "case 7",
			args: args{
				pattern: `/}id:[0-9]+{/`,
			},
			wantErr: true,
		},
		{
			name: "case 8",
			args: args{
				pattern: `/*/{id:[0-9]+}.html`,
			},
			want: `/(?P<p0>.*)/(?P<id>[0-9]+)\.html`,
		},
		{
			name: "case 9",
			args: args{
				pattern: `/{id:UUID}.{ext}`,
			},
			want: `/(?P<id>` + uuidReg + `)\.(?P<ext>[^/]+)`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parserRegexpPattern(tt.args.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("parserRegexpPattern() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			xt.Equal(t, tt.want, got)
		})
	}
}

func Test_wordNode_Match(t *testing.T) {
	type fields struct {
		Prefix string
		Suffix string
		Name   string
	}
	type args struct {
		str string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
		want1  bool
	}{
		{
			name:   "case 1",
			fields: fields{},
			args: args{
				str: "hello",
			},
			want:  "",
			want1: true,
		},
		{
			name: "case 2",
			fields: fields{
				Prefix: "hello",
			},
			args: args{
				str: "hello",
			},
			want:  "",
			want1: true,
		},
		{
			name: "case 3",
			fields: fields{
				Prefix: "hello",
				Suffix: "world",
			},
			args: args{
				str: "hello-world",
			},
			want:  "",
			want1: true,
		},
		{
			name: "case 4",
			fields: fields{
				Prefix: "hello-",
				Suffix: "-world",
				Name:   "id",
			},
			args: args{
				str: "hello-123-world",
			},
			want:  "123",
			want1: true,
		},
		{
			name: "case 4",
			fields: fields{
				Prefix: "hello-",
				Suffix: "-world",
				Name:   "id",
			},
			args: args{
				str: "hello",
			},
			want:  "",
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &wordNode{
				Prefix: tt.fields.Prefix,
				Suffix: tt.fields.Suffix,
				Name:   tt.fields.Name,
			}
			got, got1 := n.Match(tt.args.str)
			if got != tt.want {
				t.Errorf("Match() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Match() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_getPatternType(t *testing.T) {
	type args struct {
		pattern string
	}
	tests := []struct {
		name string
		args args
		want PatternType
	}{
		{
			name: "case 1",
			args: args{
				pattern: "index",
			},
			want: PatternExact,
		},
		{
			name: "case 2",
			args: args{
				pattern: "/{id}",
			},
			want: PatternWord,
		},
		{
			name: "case 3",
			args: args{
				pattern: "/*",
			},
			want: PatternRegexp,
		},
		{
			name: "case 4",
			args: args{
				pattern: "/{id:[0-9]+}",
			},
			want: PatternRegexp,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPatternType(tt.args.pattern); got != tt.want {
				t.Errorf("getPatternType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParserPattern(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {})
}
