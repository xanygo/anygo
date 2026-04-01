//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-03-29

package xcmdpool_test

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"math/rand/v2"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xanygo/anygo/ds/xpool/xcmdpool"
	"github.com/xanygo/anygo/xt"
)

func TestCommand1(t *testing.T) {
	cmd := &xcmdpool.Command{
		Path: "go",
		Args: []string{"run", filepath.Join("../../../cmd/example/stdio-ping/", "main.go")},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ts := &xt.Collector{}

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Go(func() {
			ts.Run(fmt.Sprintf("i_%d", i), func(t xt.TB) {
				for i := 0; i < 3; i++ {
					rw, err := cmd.Spawn(ctx)
					xt.NoError(t, err)
					input := []byte(fmt.Sprintf("hello i=%d rand=%d\n", i, rand.Int()))
					n, err := rw.Write(input)
					xt.NoError(t, err)
					xt.Equal(t, len(input), n)
					rd := bufio.NewReader(rw)
					output, err := rd.ReadString('\n')
					xt.NoError(t, err)
					output = strings.TrimSpace(output)
					want := "Ok: " + string(bytes.TrimSpace(input))
					xt.Equal(t, want, output)
					xt.NoError(t, rw.Close())
				}
			})
		})
	}
	wg.Wait()
	ts.Check(t)

	xt.NoError(t, cmd.Close())

	t.Run("after closed", func(t *testing.T) {
		rw, err := cmd.Spawn(ctx)
		t.Logf("cmd.Spawn err=%v", err)
		xt.Error(t, err)
		xt.Nil(t, rw)
	})
}
