//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-03-11

package sse

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestEvent_Bytes(t *testing.T) {
	tests := []struct {
		name string
		ev   Event
		want string
	}{
		{
			name: "case 1",
			want: "",
		},
		{
			name: "case 2",
			ev: Event{
				Data: "hello",
			},
			want: "data: hello\n\n",
		},
		{
			name: "case 3",
			ev: Event{
				ID:      "1",
				Event:   "write",
				Data:    "hello\nworld",
				Comment: "comment1\ncomment2",
			},
			want: "id: 1\nevent: write\ndata: hello\ndata: world\n: comment1\n: comment2\n\n",
		},
		{
			name: "case 4",
			ev: Event{
				ID:      "1",
				Event:   "write",
				Data:    "hello",
				Comment: "comment1",
				Retry:   1,
			},
			want: "id: 1\nevent: write\ndata: hello\nretry: 1\n: comment1\n\n",
		},
		{
			name: "case 5",
			ev: Event{
				Comment: "comment1",
			},
			want: ": comment1\n\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(tt.ev.Bytes())
			xt.Equal(t, got, tt.want)
		})
	}
}
