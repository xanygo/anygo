//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-07

package xtype_test

import (
	"encoding/json"
	"testing"

	"github.com/xanygo/anygo/ds/xtype"
	"github.com/xanygo/anygo/xt"
)

func TestTriStateDecode(t *testing.T) {
	t.Run("bool true", func(t *testing.T) {
		bf1 := []byte(`{"State":true,"name":"hello"}`)
		var u1 *user
		xt.NoError(t, json.Unmarshal(bf1, &u1))
		want1 := &user{Name: "hello", State: xtype.TriTrue}
		xt.Equal(t, want1, u1)
	})
	t.Run("string true", func(t *testing.T) {
		bf1 := []byte(`{"State":"true","name":"hello"}`)
		var u1 *user
		xt.NoError(t, json.Unmarshal(bf1, &u1))
		want1 := &user{Name: "hello", State: xtype.TriTrue}
		xt.Equal(t, want1, u1)
	})

	t.Run("bool false", func(t *testing.T) {
		bf1 := []byte(`{"State":false,"name":"hello"}`)
		var u1 *user
		xt.NoError(t, json.Unmarshal(bf1, &u1))
		want1 := &user{Name: "hello", State: xtype.TriFalse}
		xt.Equal(t, want1, u1)
	})

	t.Run("string false", func(t *testing.T) {
		bf1 := []byte(`{"State":"false","name":"hello"}`)
		var u1 *user
		xt.NoError(t, json.Unmarshal(bf1, &u1))
		want1 := &user{Name: "hello", State: xtype.TriFalse}
		xt.Equal(t, want1, u1)
	})

	t.Run("null", func(t *testing.T) {
		bf1 := []byte(`{"State":null,"name":"hello"}`)
		var u1 *user
		xt.NoError(t, json.Unmarshal(bf1, &u1))
		want1 := &user{Name: "hello", State: xtype.TriNull}
		xt.Equal(t, want1, u1)
	})

	t.Run("string null", func(t *testing.T) {
		bf1 := []byte(`{"State":"null","name":"hello"}`)
		var u1 *user
		xt.NoError(t, json.Unmarshal(bf1, &u1))
		want1 := &user{Name: "hello", State: xtype.TriNull}
		xt.Equal(t, want1, u1)
	})

	t.Run("invalid", func(t *testing.T) {
		bf1 := []byte(`{"State":"error","name":"hello"}`)
		var u1 *user
		xt.Error(t, json.Unmarshal(bf1, &u1))
	})
}

func TestTriStateEncode(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		u1 := user{
			Name:  "hello",
			State: xtype.TriTrue,
		}
		bf, err := json.Marshal(u1)
		xt.NoError(t, err)
		xt.Equal(t, `{"Name":"hello","State":true}`, string(bf))
	})
	t.Run("false", func(t *testing.T) {
		u1 := user{
			Name:  "hello",
			State: xtype.TriFalse,
		}
		bf, err := json.Marshal(u1)
		xt.NoError(t, err)
		xt.Equal(t, `{"Name":"hello","State":false}`, string(bf))
	})
	t.Run("null", func(t *testing.T) {
		u1 := user{
			Name: "hello",
		}
		bf, err := json.Marshal(u1)
		xt.NoError(t, err)
		xt.Equal(t, `{"Name":"hello","State":null}`, string(bf))
	})
}

type user struct {
	Name  string
	State xtype.TriState
}
