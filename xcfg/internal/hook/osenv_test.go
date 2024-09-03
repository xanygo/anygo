//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package hook

import (
	"context"
	"os"
	"reflect"
	"testing"
)

func Test_helperOsEnvVars(t *testing.T) {
	os.Setenv("appname", "fsgo/fsconf")
	os.Setenv("port", "8081")

	type args struct {
		content []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "case 1",
			args: args{
				content: []byte(`{"app":"{osenv.appname|def}","port":"{osenv.port|80}","mem":"{osenv.abc|10}{osenv.def}"}`),
			},
			want:    []byte(`{"app":"fsgo/fsconf","port":"8081","mem":"10"}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := OsEnvVars(context.Background(), "", tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("helperOsEnvVars() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("helperOsEnvVars() got = %q, want %q", got, tt.want)
			}
		})
	}
}
