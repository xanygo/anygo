package model

import (
	"context"
	"testing"
	"time"

	"github.com/xanygo/anygo/store/xdb"
)

func DoCheck(t *testing.T, client *xdb.Client) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	err := client.PingContext(ctx)
	if err != nil {
		t.Skipf("ping db failed: %v", err)
		return
	}

	t.Run("withUser", func(t *testing.T) {
		withUser(ctx, t, client)
	})

	t.Run("withMPK", func(t *testing.T) {
		withMPK(ctx, t, client)
	})
}
