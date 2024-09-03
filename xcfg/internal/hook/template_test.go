//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package hook

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/fsgo/fst"
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
Port = {osenv.Port1}

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
			fst.Equal(t, string(tt.wantOutput), string(gotOutput))
		})
	}
}

func Test_fnFetch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		k := r.URL.Query().Get("k")
		_, _ = w.Write([]byte("hello-" + k))
	}))
	defer ts.Close()
	api := ts.URL
	t.Run("server ok", func(t *testing.T) {
		txt := `# hook.template  Enable=true
{
 "K1":"{{ fetch "` + api + `?k=k1" }}",
 "K2":"{{ fetch "` + api + `?k=k2" }}",
 "K3":"{{ fetch "` + api + `?k=k3" "timeout=5s&cache=1h" }}"
}
`
		h := &Template{}
		nc, err := h.Hook(context.Background(), "", []byte(txt))
		fst.NoError(t, err)
		want := `# hook.template  Enable=true
{
 "K1":"hello-k1",
 "K2":"hello-k2",
 "K3":"hello-k3"
}
`
		fst.Equal(t, want, string(nc))
	})

	t.Run("server unreachable with cache", func(t *testing.T) {
		ts.Close()
		txt := `# hook.template  Enable=true
{
  "K3" : "{{ fetch "` + api + `?k=k3" "timeout=5s&cache=1h" }}"
}
`
		h := &Template{}
		want := `# hook.template  Enable=true
{
  "K3" : "hello-k3"
}
`
		nc, err := h.Hook(context.Background(), "", []byte(txt))
		fst.NoError(t, err)
		fst.Equal(t, want, string(nc))
	})
}
