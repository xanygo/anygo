//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-21

package xstr

import "testing"

func TestCutIndex(t *testing.T) {
	type args struct {
		s     string
		index int
		len   int
	}
	tests := []struct {
		name       string
		args       args
		wantBefore string
		wantAfter  string
	}{
		{
			name: "case 1",
			args: args{
				s:     "hello",
				index: -1,
			},
			wantBefore: "",
			wantAfter:  "hello",
		},
		{
			name: "case 2",
			args: args{
				s:     "hello",
				index: 5,
			},
			wantBefore: "hello",
			wantAfter:  "",
		},
		{
			name: "case 3",
			args: args{
				s:     "hello",
				index: 1,
			},
			wantBefore: "h",
			wantAfter:  "ello",
		},
		{
			name: "case 4",
			args: args{
				s:     "hello",
				index: 10,
			},
			wantBefore: "hello",
			wantAfter:  "",
		},
		{
			name: "case 5",
			args: args{
				s:     "hello",
				index: 1,
				len:   2,
			},
			wantBefore: "h",
			wantAfter:  "lo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBefore, gotAfter := CutIndex(tt.args.s, tt.args.index, tt.args.len)
			if gotBefore != tt.wantBefore {
				t.Errorf("CutIndex() gotBefore = %v, want %v", gotBefore, tt.wantBefore)
			}
			if gotAfter != tt.wantAfter {
				t.Errorf("CutIndex() gotAfter = %v, want %v", gotAfter, tt.wantAfter)
			}
		})
	}
}
