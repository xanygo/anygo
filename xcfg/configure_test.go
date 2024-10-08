//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package xcfg

import (
	"reflect"
	"testing"

	"github.com/fsgo/fst"

	"github.com/xanygo/anygo/xcfg/internal/hook"
	"github.com/xanygo/anygo/xcfg/internal/parser"
	"github.com/xanygo/anygo/xcodec"
)

func Test_confImpl(t *testing.T) {
	conf := &Configure{}
	testReset()
	var a any
	fst.Error(t, conf.Parse("abc.json", &a))
	fst.NoError(t, conf.WithDecoder(".json", xcodec.DecodeFunc(parser.JSON)))
	fst.Error(t, conf.Parse("abc.xyz", &a))
	fst.NoError(t, conf.Parse("testdata/db10.json", &a))
}

func TestNewDefault1(t *testing.T) {
	hd := append([]Hook{}, defaultHooks...)
	defer func() {
		defaultHooks = hd
		if re := recover(); re == nil {
			t.Errorf("want panic")
		}
	}()
	h := newHook("test", hook.OsEnvVars)
	// helper 有重复的时候
	defaultHooks = append(defaultHooks, h, h)
	NewDefault()
}

func Test_confImpl_ParseBytes(t *testing.T) {
	type args struct {
		fileExt string
		content []byte
		obj     map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    map[string]string
	}{
		{
			name: "case 1",
			args: args{
				fileExt: "",
				content: nil,
				obj:     map[string]string{},
			},
			wantErr: true,
		},
		{
			name: "case 2",
			args: args{
				fileExt: ".json",
				content: []byte(`{"Name":"Hello"}`),
				obj:     map[string]string{},
			},
			wantErr: false,
			want:    map[string]string{"Name": "Hello"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewDefault()
			if err := c.ParseBytes(tt.args.fileExt, tt.args.content, &tt.args.obj); (err != nil) != tt.wantErr {
				t.Errorf("ParseBytes() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if !reflect.DeepEqual(tt.args.obj, tt.want) {
					t.Errorf("ParseBytes(), obj=%v, got=%v", tt.args.obj, tt.want)
				}
			}
		})
	}
}
