//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-04-01

package stdio_test

import (
	"bufio"
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xnet/internal/stdio"
	"github.com/xanygo/anygo/xt"
)

func TestDialer_DialContext(t *testing.T) {
	var d stdio.Dialer
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	fp := filepath.Join("../../../cmd/example/stdio-ping/", "main.go")
	data := map[string]any{
		"Path": "go",
		"Args": []string{
			"run",
			fp,
		},
	}
	address, _ := xcodec.EncodeToString(xcodec.JSON, data)

	for i := 0; i < 3; i++ {
		t.Run(fmt.Sprintf("DialContext_%d", i), func(t *testing.T) {
			conn, err := d.DialContext(ctx, "stdio", address)
			xt.NoError(t, err)
			xt.NotNil(t, conn)

			rd := bufio.NewReader(conn)

			for j := 0; j < 3; j++ {
				t.Run(fmt.Sprintf("inner_loop_%d", j), func(t *testing.T) {
					n, err := fmt.Fprint(conn, "hello\n")
					xt.Equal(t, 6, n)
					xt.NoError(t, err)

					line, err := rd.ReadString('\n')
					xt.NoError(t, err)
					xt.Equal(t, "Ok: hello\n", line)
				})
			}

			xt.NoError(t, conn.Close())
		})
	}
}
