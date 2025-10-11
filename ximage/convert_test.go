//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-10

package ximage

import "testing"

func TestCanvasScale(t *testing.T) {
	type args struct {
		width  int
		height int
		scale  float64
	}
	tests := []struct {
		name       string
		args       args
		wantWidth  int
		wantHeight int
	}{
		{
			name: "case 1",
			args: args{
				width:  998,
				height: 1495,
				scale:  99.0 / 70,
			},
			wantWidth:  1057,
			wantHeight: 1495,
		},
		{
			name: "case 2",
			args: args{
				width:  1495,
				height: 998,
				scale:  99.0 / 70,
			},
			wantWidth:  1495,
			wantHeight: 1057,
		},
		{
			name: "case 3",
			args: args{
				width:  1495,
				height: 1100,
				scale:  99.0 / 70,
			},
			wantWidth:  1555,
			wantHeight: 1100,
		},
		{
			name: "case 4",
			args: args{
				width:  1100,
				height: 1495,
				scale:  99.0 / 70,
			},
			wantWidth:  1100,
			wantHeight: 1555,
		},
		{
			name: "case 5",
			args: args{
				width:  1100,
				height: 1495,
				scale:  70.0 / 99,
			},
			wantWidth:  1100,
			wantHeight: 1555,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := CanvasScale(tt.args.width, tt.args.height, tt.args.scale)
			t.Logf("with=%d height=%d", got, got1)
			if got != tt.wantWidth {
				t.Errorf("CanvasScale() got witdh  = %v, want %v", got, tt.wantWidth)
			}
			if got1 != tt.wantHeight {
				t.Errorf("CanvasScale() got height = %v, want %v", got1, tt.wantHeight)
			}
		})
	}
}
