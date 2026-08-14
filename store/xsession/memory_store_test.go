//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-31

package xsession

import (
	"testing"
	"time"

	"github.com/xanygo/anygo/xt"
)

func TestNewMemoryStore(t *testing.T) {
	ms := NewMemoryStore(100, time.Minute)
	session := ms.Get(t.Context(), "123")
	xt.NotNil(t, session)

	session.Set(t.Context(), "k1", "v1")
	got1, err1 := session.Get(t.Context(), "k1")
	xt.NoError(t, err1)
	xt.Equal(t, got1, "v1")
	session.Delete(t.Context(), "k1")

	got2, err2 := session.Get(t.Context(), "k1")
	xt.NoError(t, err2)
	xt.Empty(t, got2)
	xt.NoError(t, session.Save(t.Context()))
}
