package zreflect_test

import (
	"testing"

	"github.com/xanygo/anygo/internal/zerror"
	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/xt"
)

func TestMapHasKey(t *testing.T) {
	type args struct {
		m   any
		key any
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "case 1",
			args: args{
				m:   map[string]any{"k1": "v1"},
				key: "k1",
			},
			want: true,
		},
		{
			name: "case 2",
			args: args{
				m:   map[string]any{"k1": "v1"},
				key: "k2",
			},
			want: false,
		},
		{
			name: "case 3",
			args: args{
				m:   map[string]any{"k1": "v1"},
				key: 123,
			},
			want: false,
		},
		{
			name: "case 4",
			args: args{
				m:   map[string]any{"k1": "v1"},
				key: nil,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "case 5",
			args: args{
				m:   nil,
				key: "k1",
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "case 6",
			args: args{
				m:   map[int64]any{123: "v1"},
				key: int(123),
			},
			want: true,
		},
		{
			name: "case 7",
			args: args{
				m:   map[int64]any{123: "v1"},
				key: uint8(123),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := zreflect.MapHasKey(tt.args.m, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapHasKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MapHasKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRangMap(t *testing.T) {
	xt.NoError(t, zreflect.RangeMap(nil, func(key string, value string) error { return nil }))
	xt.NoError(t, zreflect.RangeMap(map[string]string{}, func(key string, value string) error { return nil }))

	var a map[string]string
	xt.NoError(t, zreflect.RangeMap(a, func(key string, value string) error { return nil }))

	a = map[string]string{"k1": "v1"}
	got1 := map[string]string{}
	xt.NoError(t, zreflect.RangeMap(a, func(key string, value string) error {
		got1[key] = value
		return nil
	}))
	xt.Equal(t, got1, a)

	b := map[string]any{"k1": "v1", "k2": nil, "k3": 123}
	got2 := map[string]string{}
	xt.NoError(t, zreflect.RangeMap(b, func(key string, value string) error {
		got2[key] = value
		return nil
	}))
	xt.Equal(t, got2, map[string]string{"k1": "v1"})

	c := map[any]any{"k1": "v1", "k2": nil, "k3": 123}
	got3 := map[string]string{}
	xt.NoError(t, zreflect.RangeMap(c, func(key string, value string) error {
		got3[key] = value
		return nil
	}))
	xt.Equal(t, got3, map[string]string{"k1": "v1"})

	xt.NoError(t, zreflect.RangeMap(c, func(key string, value string) error {
		return zerror.ErrBreak
	}))

	xt.Error(t, zreflect.RangeMap("hello", func(key string, value string) error {
		return nil
	}))
}
