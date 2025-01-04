//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package xcfg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fsgo/fst"

	"github.com/xanygo/anygo/xcfg/internal/hook"
	"github.com/xanygo/anygo/xcfg/internal/parser"
	"github.com/xanygo/anygo/xcodec"
)

func TestExists(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err.Error())
	}

	type args struct {
		confName string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "case 1",
			args: args{
				confName: "not_exists.toml",
			},
			want: false,
		},
		{
			name: "case 2",
			args: args{
				confName: "abc.json",
			},
			want: true,
		},
		{
			name: "case 2.1",
			args: args{
				confName: "abc",
			},
			want: true,
		},
		{
			name: "case 3 relative path",
			args: args{
				confName: "./testdata/conf/abc.json",
			},
			want: true,
		},
		{
			name: "case 4 relative path",
			args: args{
				confName: "../" + filepath.Base(wd) + "/testdata/conf/abc.json",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Exists(tt.args.confName); got != tt.want {
				t.Errorf("Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	type args struct {
		confName string
		obj      map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name: "case 1",
			args: args{
				confName: "not_exists.json",
				obj:      map[string]string{},
			},
			want:    map[string]string{},
			wantErr: true,
		},
		{
			name: "case 2",
			args: args{
				confName: "abc.json",
				obj:      map[string]string{},
			},
			want: map[string]string{
				"A": "bb",
			},
			wantErr: false,
		},
		{
			name: "case 3",
			args: args{
				confName: "db1.toml",
				obj:      map[string]string{},
			},
			wantErr: true,
		},
		{
			name: "case 4",
			args: args{
				confName: "db1",
				obj:      map[string]string{},
			},
			wantErr: true,
		},
		{
			name: "case 5",
			args: args{
				confName: "db2", // 存在同名目录的情况
				obj:      map[string]string{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Parse(tt.args.confName, &tt.args.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				fst.Equal(t, tt.want, tt.args.obj)
			}
		})
	}
}

func TestWithHook(t *testing.T) {
	type args struct {
		name string
		fn   hook.Func
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "case 1",
			args: args{
				name: "",
				fn:   hook.OsEnvVars,
			},
			wantErr: true,
		},
		{
			name: "case 2",
			args: args{
				name: "test_helper",
				fn:   hook.OsEnvVars,
			},
			wantErr: false,
		},
		{
			name: "case 3- name is same as case 2",
			args: args{
				name: "test_helper",
				fn:   hook.OsEnvVars,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHook(tt.args.name, tt.args.fn)
			if err := WithHook(h); (err != nil) != tt.wantErr {
				t.Errorf("RegisterInterceptor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWithDecoder(t *testing.T) {
	type args struct {
		fileExt string
		fn      xcodec.Decoder
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "case 1",
			args: args{
				fileExt: ".json",
				fn:      xcodec.DecodeFunc(parser.JSON),
			},
			wantErr: true,
		},
		{
			name: "case 2",
			args: args{
				fileExt: ".myjson",
				fn:      xcodec.DecodeFunc(parser.JSON),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := WithDecoder(tt.args.fileExt, tt.args.fn); (err != nil) != tt.wantErr {
				t.Errorf("WithParser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
