//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-14

package xcachex_test

import (
	"context"
	"testing"
	"time"

	"github.com/xanygo/anygo/internal/redistest"
	"github.com/xanygo/anygo/store/xcache/xcachex"
	"github.com/xanygo/anygo/store/xredis"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xt"
)

func TestRedis(t *testing.T) {
	ts, errTs := redistest.NewServer()
	if errTs != nil {
		t.Logf("create redis fail: %v", errTs)
		return
	}
	defer ts.Stop()
	t.Logf("uri= %q", ts.URI())

	_, client, errClient := xredis.NewClientByURI("demo", ts.URI())
	xt.NoError(t, errClient)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	rc := &xcachex.Redis{
		Client: client,
	}
	err := rc.Set(ctx, "k1", "v1", time.Minute)
	xt.NoError(t, err)

	val, err := rc.Get(ctx, "k1")
	xt.NoError(t, err)
	xt.Equal(t, "v1", val)

	err = rc.Delete(ctx, "k1")
	xt.NoError(t, err)

	val, err = rc.Get(ctx, "k1")
	xt.ErrorIs(t, err, xerror.NotFound)
	xt.Equal(t, "", val)

	vs := map[string]string{
		"k2": "v2",
		"k3": "v3",
	}
	err = rc.MSet(ctx, vs, time.Minute)
	xt.NoError(t, err)

	values, err := rc.MGet(ctx, "k1", "k2", "k3")
	xt.NoError(t, err)
	xt.Equal(t, vs, values)
}
