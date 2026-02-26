//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package hook

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/xanygo/anygo/xattr"
	"github.com/xanygo/anygo/xt"
)

func Test_getAttrValue(t *testing.T) {
	xattr.Default = xattr.NewAttribute("demo", "testdata")

	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "RootDir",
			args: args{
				key: "RootDir",
			},
			want: filepath.Join("testdata"),
		},
		{
			name: "IDC",
			args: args{
				key: "IDC",
			},
			want: xattr.IDCOnline,
		},
		{
			name: "DataDir",
			args: args{
				key: "DataDir",
			},
			want: filepath.Join("testdata", "data"),
		},
		{
			name: "ConfDir",
			args: args{
				key: "ConfDir",
			},
			want: filepath.Join("testdata", "conf"),
		},
		{
			name: "LogDir",
			args: args{
				key: "LogDir",
			},
			want: filepath.Join("testdata", "log"),
		},
		{
			name: "RunMode",
			args: args{
				key: "RunMode",
			},
			want: "product",
		},
		{
			name: "other key not support",
			args: args{
				key: "other-key",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getAttrValue(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("getValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				xt.Equal(t, tt.want, got)
			}
		})
	}
}

func TestXAttrVars(t *testing.T) {
	type args struct {
		input []byte
	}
	tests := []struct {
		name       string
		args       args
		wantOutput []byte
		wantErr    bool
	}{
		{
			name: "idc and log dir",
			args: args{
				input: []byte(`{"idc":"{xattr.IDC}","logDir":"{xattr.LogDir}"}`),
			},
			wantOutput: fmt.Appendf(nil, `{"idc":"online","logDir":"%s"}`, filepath.Join("testdata", "log")),
		},
		{
			name: "not support key",
			args: args{
				input: []byte(`{"idc":"{xattr.other}"}`),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOutput, err := XAttrVars(context.Background(), "", tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				xt.Equal(t, string(tt.wantOutput), string(gotOutput))
			}
		})
	}
}
