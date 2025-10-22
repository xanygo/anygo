//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-30

package xurl

import (
	"net/url"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestHostPort(t *testing.T) {
	type args struct {
		u string
	}
	tests := []struct {
		name     string
		args     args
		wantHost string
		wantPort uint16
		wantErr  bool
	}{
		{
			name: "case 1",
			args: args{
				u: "http://example.com",
			},
			wantHost: "example.com",
			wantPort: 80,
		},
		{
			name: "case 2",
			args: args{
				u: "https://example.com",
			},
			wantHost: "example.com",
			wantPort: 443,
		},
		{
			name: "case 3",
			args: args{
				u: "ftp://example.com",
			},
			wantHost: "example.com",
			wantPort: 21,
		},
		{
			name: "case 4",
			args: args{
				u: "http://example.com:8080/",
			},
			wantHost: "example.com",
			wantPort: 8080,
		},
		{
			name: "case 5",
			args: args{
				u: "https://example.com:8443",
			},
			wantHost: "example.com",
			wantPort: 8443,
		},
		{
			name: "case 6",
			args: args{
				u: "https://example.com:8443/user",
			},
			wantHost: "example.com",
			wantPort: 8443,
		},
		{
			name: "case 7",
			args: args{
				u: "https://example.com:78443/user",
			},
			wantHost: "example.com",
			wantPort: 0,
			wantErr:  true,
		},
		{
			name: "case 8",
			args: args{
				u: "example.com/user",
			},
			wantErr: true,
		},
		{
			name: "case 9",
			args: args{
				u: "https://[2001:db8::1]/user",
			},
			wantHost: "2001:db8::1",
			wantPort: 443,
			wantErr:  false,
		},
		{
			name: "case 10",
			args: args{
				u: "https://[2001:db8::1]:8443/user",
			},
			wantHost: "2001:db8::1",
			wantPort: 8443,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.args.u)
			xt.NoError(t, err)
			gotHost, gotPort, err := HostPort(u)
			if (err != nil) != tt.wantErr {
				t.Errorf("HostPort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotHost != tt.wantHost {
				t.Errorf("HostPort() gotHost = %v, want %v", gotHost, tt.wantHost)
			}
			if gotPort != tt.wantPort {
				t.Errorf("HostPort() gotPort = %v, want %v", gotPort, tt.wantPort)
			}
		})
	}
}
