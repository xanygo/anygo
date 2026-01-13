//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-12

package xtype_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/xanygo/anygo/ds/xtype"
	"github.com/xanygo/anygo/xt"
)

func TestDuration_String(t *testing.T) {
	tests := []struct {
		name string
		d    xtype.Duration
		want string
	}{
		{
			name: "case 1",
			d:    xtype.Duration(time.Second),
			want: "1s",
		},
		{
			name: "case 2",
			d:    xtype.Duration(30*time.Second + 301*time.Millisecond),
			want: "30.301s",
		},
		{
			name: "case 3",
			d:    xtype.Duration(30*time.Second + 1*time.Millisecond),
			want: "30.001s",
		},
		{
			name: "case 4",
			d:    xtype.Duration(30*time.Second + 1*time.Millisecond + 12*time.Microsecond),
			want: "30.001s",
		},
		{
			name: "case 5",
			d:    xtype.Duration(30*time.Second + 1*time.Millisecond + 120*time.Microsecond),
			want: "30.001s",
		},
		{
			name: "case 6",
			d:    xtype.Duration(1*time.Millisecond + 120*time.Microsecond + 111*time.Nanosecond),
			want: "1.12ms",
		},
		{
			name: "case 7",
			d:    xtype.Duration(1*time.Millisecond + 121*time.Microsecond + 111*time.Nanosecond),
			want: "1.121ms",
		},
		{
			name: "case 8",
			d:    xtype.Duration(121*time.Microsecond + 111*time.Nanosecond),
			want: "121.111µs",
		},
		{
			name: "case 9",
			d:    xtype.Duration(time.Hour + 121*time.Millisecond + 111*time.Nanosecond),
			want: "1h0m0.121s",
		},
		{
			name: "case 10",
			d:    xtype.Duration(time.Hour + 121*time.Microsecond + 111*time.Nanosecond),
			want: "1h0m0s",
		},
		{
			name: "case 11",
			d:    xtype.Duration(111 * time.Nanosecond),
			want: "111ns",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDuration_UnmarshalText(t *testing.T) {
	type service struct {
		Timeout xtype.Duration
	}
	tests := []struct {
		name    string
		arg     string
		wantErr bool
		want    service
	}{
		{
			name: "case 1",
			arg:  `{"Timeout":"1s"}`,
			want: service{Timeout: xtype.Duration(time.Second)},
		},
		{
			name: "case 2",
			arg:  `{"Timeout":"2h1s"}`,
			want: service{Timeout: xtype.Duration(2*time.Hour + time.Second)},
		},
		{
			name: "case 3",
			arg:  `{"Timeout":1000}`,
			want: service{Timeout: xtype.Duration(1000 * time.Millisecond)},
		},
		{
			name: "case 4",
			arg:  `{"Timeout":"1000"}`,
			want: service{Timeout: xtype.Duration(1000 * time.Millisecond)},
		},
		{
			name:    "case 5",
			arg:     `{"Timeout":"10day"}`,
			wantErr: true,
		},
		{
			name: "case 6",
			arg:  `{"Timeout":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m1 service
			err := json.Unmarshal([]byte(tt.arg), &m1)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
			}
			xt.Equal(t, tt.want, m1)
		})
	}
}
