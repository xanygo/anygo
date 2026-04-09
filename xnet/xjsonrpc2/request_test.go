//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-03-30

package xjsonrpc2

import (
	"bufio"
	"reflect"
	"strings"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func Test_parserRequest(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Request
		wantErr bool
	}{
		{
			name:    "case 1",
			input:   "",
			wantErr: true,
		},
		{
			name:  "case 2 int id",
			input: `{"jsonrpc": "2.0", "method": "subtract", "params": [42, 23], "id": 1}`,
			want: &Request{
				ID:     Int64ID(1),
				Method: "subtract",
				Params: []byte(`[42, 23]`),
			},
		},
		{
			name:  "case 3 no id",
			input: `{"jsonrpc": "2.0", "method": "update", "params": [1,2,3,4,5]}`,
			want: &Request{
				Method: "update",
				Params: []byte(`[1,2,3,4,5]`),
			},
		},
		{
			name:    "case 4 invalid",
			input:   `{"jsonrpc": "2.0", "method": "foobar, "params": "bar", "baz]`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parserRequest([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("parserRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parserRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadRequests(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []*Request
		want1   bool
		wantErr bool
	}{
		{
			name:    "case 1",
			want1:   false,
			wantErr: true,
		},
		{
			name: "case 2",
			input: `{"jsonrpc": "2.0", "method": "foobar", "id": "1"}
{"jsonrpc": "2.0", "method": "foobar", "id": "2"}
`,
			want: []*Request{
				{
					Method: "foobar",
					ID:     StringID("1"),
				},
			},
			want1:   false,
			wantErr: false,
		},
		{
			name: "case 3",
			input: `[
{"jsonrpc": "2.0", "method": "foobar", "id": "1"},
{"jsonrpc": "2.0", "method": "foobar", "id": 2}
]
`,
			want: []*Request{
				{
					Method: "foobar",
					ID:     StringID("1"),
				},
				{
					Method: "foobar",
					ID:     Int64ID(2),
				},
			},
			want1:   true,
			wantErr: false,
		},
		{
			name: "case 4",
			input: `[{"jsonrpc": "2.0", "method": "foobar", "id": "1"},{"jsonrpc": "2.0", "method": "foobar", "id": 2}]
`, // 尾部有换行符
			want: []*Request{
				{
					Method: "foobar",
					ID:     StringID("1"),
				},
				{
					Method: "foobar",
					ID:     Int64ID(2),
				},
			},
			want1:   true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rd := bufio.NewReader(strings.NewReader(tt.input))
			got, got1, err := ReadRequests(rd)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadRequests() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			xt.Equal(t, got, tt.want)
			xt.Equal(t, got1, tt.want1)
		})
	}
}

func TestReadRequests2(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		input := `{"jsonrpc": "2.0", "method": "foobar", "id": "1"}
{"jsonrpc": "2.0", "method": "foobar", "id": "2"}
`
		rd := bufio.NewReader(strings.NewReader(input))

		t.Run("first read", func(t *testing.T) {
			reqs, batch, err := ReadRequests(rd)
			xt.NoError(t, err)
			xt.False(t, batch)
			want1 := []*Request{
				{
					Method: "foobar",
					ID:     StringID("1"),
				},
			}
			xt.Equal(t, reqs, want1)
		})

		t.Run("second read", func(t *testing.T) {
			reqs, batch, err := ReadRequests(rd)
			xt.NoError(t, err)
			xt.False(t, batch)
			want1 := []*Request{
				{
					Method: "foobar",
					ID:     StringID("2"),
				},
			}
			xt.Equal(t, reqs, want1)
		})

		t.Run("third read", func(t *testing.T) {
			reqs, batch, err := ReadRequests(rd)
			xt.Error(t, err)
			xt.False(t, batch)
			xt.Empty(t, reqs)
		})
	})

	t.Run("case 2 batch", func(t *testing.T) {
		input := `[
{"jsonrpc": "2.0", "method": "foobar", "id": "1"},
{"jsonrpc": "2.0", "method": "foobar", "id": "2"}
]
{"jsonrpc": "2.0", "method": "foobar", "id": "3"}
`
		rd := bufio.NewReader(strings.NewReader(input))

		t.Run("first read", func(t *testing.T) {
			reqs, batch, err := ReadRequests(rd)
			xt.NoError(t, err)
			xt.True(t, batch)
			want1 := []*Request{
				{
					Method: "foobar",
					ID:     StringID("1"),
				},
				{
					Method: "foobar",
					ID:     StringID("2"),
				},
			}
			xt.Equal(t, reqs, want1)
		})

		t.Run("second read", func(t *testing.T) {
			reqs, batch, err := ReadRequests(rd)
			xt.NoError(t, err)
			xt.False(t, batch)
			want1 := []*Request{
				{
					Method: "foobar",
					ID:     StringID("3"),
				},
			}
			xt.Equal(t, reqs, want1)
		})

		t.Run("third read", func(t *testing.T) {
			reqs, batch, err := ReadRequests(rd)
			xt.Error(t, err)
			xt.False(t, batch)
			xt.Empty(t, reqs)
		})
	})
}
