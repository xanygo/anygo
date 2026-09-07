//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-12

package xtime_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/xanygo/anygo/xtime"
)

func TestTimestampSecond_Time(t *testing.T) {
	ts := xtime.TimestampSecond(1719158400)

	got := ts.Time()
	want := time.Unix(1719158400, 0)

	if !got.Equal(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestTimestampSecond_Unix(t *testing.T) {
	ts := xtime.TimestampSecond(1719158400)

	if ts.Unix() != 1719158400 {
		t.Fatalf("got %d", ts.Unix())
	}
}

func TestTimestampSecond_SetTime(t *testing.T) {
	var ts xtime.TimestampSecond

	now := time.Unix(1719158400, 0)
	ts.SetTime(now)

	if ts != xtime.TimestampSecond(1719158400) {
		t.Fatalf("got %d", ts)
	}
}

func TestTimestampSecond_MarshalJSON(t *testing.T) {
	b, err := json.Marshal(xtime.TimestampSecond(1719158400))
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != "1719158400" {
		t.Fatalf("got %s", b)
	}
}

func TestTimestampSecond_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		data string
		want xtime.TimestampSecond
		ok   bool
	}{
		{name: "number", data: `1719158400`, want: 1719158400, ok: true},
		{name: "string", data: `"1719158400"`, want: 1719158400, ok: true},
		{name: "null", data: `null`, want: 0, ok: true},
		{name: "invalid", data: `"abc"`, want: 0, ok: false},
		{name: "invalid_type", data: `{}`, want: 0, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got xtime.TimestampSecond
			err := json.Unmarshal([]byte(tt.data), &got)

			if tt.ok {
				if err != nil {
					t.Fatal(err)
				}
				if got != tt.want {
					t.Fatalf("got %d, want %d", got, tt.want)
				}
			} else if err == nil {
				t.Fatal("want error")
			}
		})
	}
}

func TestTimestampSecond_MarshalText(t *testing.T) {
	b, err := xtime.TimestampSecond(1719158400).MarshalText()
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != "1719158400" {
		t.Fatalf("got %s", b)
	}
}

func TestTimestampSecond_UnmarshalText(t *testing.T) {
	var ts xtime.TimestampSecond

	if err := ts.UnmarshalText([]byte("1719158400")); err != nil {
		t.Fatal(err)
	}

	if ts != 1719158400 {
		t.Fatalf("got %d", ts)
	}
}
