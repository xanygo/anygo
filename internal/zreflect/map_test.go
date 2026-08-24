package zreflect

import "testing"

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
			got, err := MapHasKey(tt.args.m, tt.args.key)
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
