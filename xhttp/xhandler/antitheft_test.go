//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-08

package xhandler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fsgo/fst"
)

func TestAntiTheft_check(t *testing.T) {
	tests := []struct {
		name  string
		at    *AntiTheft
		url   string
		refer string
		want  bool
	}{
		{
			name:  "case 1",
			at:    &AntiTheft{},
			url:   "http://a.example.com/index",
			refer: "http://a.example.com/index",
			want:  true,
		},
		{
			name:  "case 2",
			at:    &AntiTheft{},
			url:   "https://a.example.com/index",
			refer: "https://a.example.com/index",
			want:  true,
		},
		{
			name:  "case 3",
			at:    &AntiTheft{},
			url:   "http://a.example.com/index",
			refer: "",
			want:  true,
		},
		{
			name:  "case 4",
			at:    &AntiTheft{},
			url:   "https://example.com/index",
			refer: "https://a.example.com/index",
			want:  true,
		},
		{
			name: "case 5",
			at: &AntiTheft{
				AllowDomain: []string{"hello.com"},
			},
			url:   "https://abc.com/index",
			refer: "https://a.example.com/index",
			want:  false,
		},
		{
			name: "case 6",
			at: &AntiTheft{
				AllowDomain: []string{"hello.com"},
			},
			url:   "https://abc.com/index",
			refer: "https://a.hello.com/index",
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			req.Header.Set("referer", tt.refer)
			fst.Equal(t, tt.want, tt.at.check(req))
		})
	}
}
