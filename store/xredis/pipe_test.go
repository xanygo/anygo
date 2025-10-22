//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-15

package xredis

import (
	"context"
	"testing"
	"time"

	"github.com/xanygo/anygo/internal/redistest"
	"github.com/xanygo/anygo/xt"
)

func TestClient_Pipeline(t *testing.T) {
	ts, errTs := redistest.NewServer()
	if errTs != nil {
		t.Logf("create redis fail: %v", errTs)
		return
	}
	defer ts.Stop()
	t.Logf("uri= %q", ts.URI())

	_, client, errClient := NewClientByURI("demo", ts.URI())
	xt.NoError(t, errClient)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("Pipelined", func(t *testing.T) {
		cmds, err := client.Pipelined(ctx, func(ctx context.Context, pipe *Pipeline) error {
			pipe.NewAnyCmd("incr", "k1")
			pipe.NewAnyCmd("incr", "k2")
			return nil
		})
		xt.NoError(t, err)
		xt.Len(t, cmds, 2)
		var num any = int64(1)
		xt.Equal(t, num, cmds[0].Value())
		xt.NoError(t, cmds[0].Err())
		xt.Equal(t, num, cmds[1].Value())
		xt.NoError(t, cmds[1].Err())

		cmds, err = client.Pipelined(ctx, func(ctx context.Context, pipe *Pipeline) error {
			pipe.NewAnyCmd("incr", "k1-1")
			pipe.NewAnyCmd("incr") // invalid cmd
			pipe.NewAnyCmd("incr", "k1-2")
			return nil
		})
		xt.NoError(t, err)
		xt.Len(t, cmds, 3)

		xt.Equal(t, num, cmds[0].Value())
		xt.NoError(t, cmds[0].Err())

		xt.Error(t, cmds[1].Err())

		xt.Equal(t, num, cmds[2].Value())
		xt.NoError(t, cmds[0].Err())
	})

	t.Run("TxPipelined", func(t *testing.T) {
		cmds, err := client.TxPipelined(ctx, func(ctx context.Context, pipe *Pipeline) error {
			pipe.NewAnyCmd("incr", "k3")
			pipe.NewAnyCmd("incr", "k4")
			return nil
		})
		xt.NoError(t, err)
		xt.Len(t, cmds, 2)
		var num any = int64(1)
		xt.Equal(t, num, cmds[0].Value())
		xt.NoError(t, cmds[0].Err())
		xt.Equal(t, num, cmds[1].Value())
		xt.NoError(t, cmds[1].Err())
	})
}
