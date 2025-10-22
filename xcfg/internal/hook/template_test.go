//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package hook

import (
	"context"
	"os"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func mustRead(name string) []byte {
	bf, err := os.ReadFile(name)
	if err != nil {
		panic(err)
	}
	return bf
}

func TestTemplateHook(t *testing.T) {
	type args struct {
		ctx     context.Context
		cfPath  string
		content []byte
	}
	tests := []struct {
		name       string
		args       args
		wantOutput []byte
		wantErr    bool
	}{
		{
			name: "include.toml",
			args: args{
				ctx:     context.Background(),
				cfPath:  "testdata/include.toml",
				content: mustRead("testdata/include.toml"),
			},
			wantOutput: []byte(
				`# hook.template  Enable=true
A="a"
Port = {env.Port1}

B="b"
B1="b1"
C="c"

Z="z"


`),
		},
		{
			name: "include not found",
			args: args{
				ctx:     context.Background(),
				cfPath:  "testdata/include_e1.toml",
				content: mustRead("testdata/include_e1.toml"),
			},
			wantErr: true,
		},
		{
			name: "include FilePath Empty",
			args: args{
				ctx:     context.Background(),
				cfPath:  "",
				content: mustRead("testdata/include_e1.toml"),
			},
			wantErr: true,
		},
		{
			name: "include not enable",
			args: args{
				ctx:     context.Background(),
				cfPath:  "",
				content: mustRead("testdata/include_not_enable.toml"),
			},
			wantOutput: []byte("A=\"a\"\n\n{{ include \"not_found.toml\" }}\n"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Template{}
			gotOutput, err := h.Hook(tt.args.ctx, tt.args.cfPath, tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			xt.Equal(t, string(tt.wantOutput), string(gotOutput))
		})
	}
}
