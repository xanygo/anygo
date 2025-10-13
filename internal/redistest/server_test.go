//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-13

package redistest

import "testing"

func TestNewServer(t *testing.T) {
	ts, err := NewServer()
	if err != nil {
		t.Logf("create redis fail: %v", err)
		return
	}
	defer ts.Stop()
	t.Logf("uri=%q", ts.URI())
}
