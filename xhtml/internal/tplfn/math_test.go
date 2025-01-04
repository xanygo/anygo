//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-01-04

package tplfn

import (
	"reflect"
	"testing"
)

func TestMathAdd(t *testing.T) {
	type args struct {
		items []any
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{
			name: "case 1",
			args: args{
				items: []any{1},
			},
			want: int64(1),
		},
		{
			name: "case 2",
			args: args{
				items: []any{1, uint(2)},
			},
			want: int64(3),
		},
		{
			name: "case 3",
			args: args{
				items: []any{1, uint(2), float64(3)},
			},
			want: float64(6),
		},
		{
			name: "case 4",
			args: args{
				items: []any{1, uint(2), float64(3), "hello"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MathAdd(tt.args.items...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Add() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMathSub(t *testing.T) {
	type args struct {
		first any
		items []any
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{
			name: "case 1",
			args: args{
				first: int64(1),
			},
			want: int64(1),
		},
		{
			name: "case 2",
			args: args{
				first: 1,
				items: []any{int64(1)},
			},
			want: int64(0),
		},
		{
			name: "case 3",
			args: args{
				first: 1,
				items: []any{int64(1), float64(2)},
			},
			want: float64(-2),
		},
		{
			name: "case 4",
			args: args{
				first: 1,
				items: []any{int64(1), float64(2), "hello"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MathSub(tt.args.first, tt.args.items...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sub() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sub() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMathMul(t *testing.T) {
	type args struct {
		items []any
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{
			name: "case 1",
			args: args{},
			want: int64(0),
		},
		{
			name: "case 2",
			args: args{
				items: []any{int64(1), int64(2)},
			},
			want: int64(2),
		},
		{
			name: "case 3",
			args: args{
				items: []any{int64(1), int64(2), 3},
			},
			want: int64(6),
		},
		{
			name: "case 4",
			args: args{
				items: []any{int64(1), int64(2), 3, float32(2)},
			},
			want: float64(12),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MathMul(tt.args.items...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Mul() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Mul() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMathDiv(t *testing.T) {
	type args struct {
		first any
		items []any
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{
			name: "case 1",
			args: args{
				first: int64(20),
			},
			want: float64(20),
		},
		{
			name: "case 2",
			args: args{
				first: int64(20),
				items: []any{int64(2)},
			},
			want: float64(10),
		},
		{
			name: "case 3",
			args: args{
				first: int64(20),
				items: []any{2, float32(2)},
			},
			want: float64(5),
		},
		{
			name: "case 4",
			args: args{
				first: int64(20),
				items: []any{2, float32(2), "hello"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MathDiv(tt.args.first, tt.args.items...)
			if (err != nil) != tt.wantErr {
				t.Errorf("MathDiv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MathDiv() got = %v, want %v", got, tt.want)
			}
		})
	}
}
