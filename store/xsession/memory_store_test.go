//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-31

package xsession

import (
	"context"
	"testing"
	"time"

	"github.com/xanygo/anygo/xt"
)

func TestNewMemoryStore(t *testing.T) {
	ms := NewMemoryStore(100, time.Minute)
	session := ms.Get(context.Background(), "123")
	xt.NotNil(t, session)

	session.Set(context.Background(), "k1", "v1")
	got1, err1 := session.Get(context.Background(), "k1")
	xt.NoError(t, err1)
	xt.Equal(t, "v1", got1)
	session.Delete(context.Background(), "k1")

	got2, err2 := session.Get(context.Background(), "k1")
	xt.NoError(t, err2)
	xt.Empty(t, got2)
	xt.NoError(t, session.Save(context.Background()))
}
