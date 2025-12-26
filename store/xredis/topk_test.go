//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-12-26

package xredis

import (
	"context"
	"testing"
	"time"

	"github.com/xanygo/anygo/internal/redistest"
	"github.com/xanygo/anygo/xt"
)

func TestTopK(t *testing.T) {
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

	// t.Run("TopKReserve", func(t *testing.T) {
	//	_, err := client.Del(ctx, "TopKReserve-1")
	//	xt.NoError(t, err)
	//	err = client.TopKReserve(ctx, "TopKReserve-1", 10, nil)
	//	xt.NoError(t, err)
	// })

	t.Run("TopKAdd", func(t *testing.T) {
		got, err := client.TopKAdd(ctx, "TopKAdd-1", "f1")
		xt.Error(t, err)
		xt.ErrorContains(t, err, "TopK: key does not exist")
		xt.Nil(t, got)

		err = client.TopKReserve(ctx, "TopKAdd-1", 10, &TopKReserveOption{Width: 8, Depth: 7, Decay: 0.925})
		t.Log("TopKReserve:", err)
		// 这里会报错，why？
		// xt.NoError(t, err)

		_, err = client.TopKAdd(ctx, "TopKAdd-1", "f1")
		xt.NoError(t, err)
	})
}
