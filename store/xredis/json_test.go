//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-12-29

package xredis

import (
	"context"
	"testing"
	"time"

	"github.com/xanygo/anygo/internal/redistest"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xt"
)

func TestClientJSON(t *testing.T) {
	ts, errTs := redistest.NewServer()
	if errTs != nil {
		t.Skipf("create redis-server skipped: %v", errTs)
		return
	}
	defer ts.Stop()
	t.Logf("uri= %q", ts.URI())
	_, client, errClient := NewClientByURI("demo", ts.URI())
	xt.NoError(t, errClient)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("JSONArrAppend", func(t *testing.T) {
		got, err := client.JSONArrAppend(ctx, "JSONArrAppend-1", "$.colors", `"blue"`)
		xt.Error(t, err)
		// ERR could not perform this operation on a key that doesn't exist
		xt.True(t, xerror.IsNotFound(err))
		xt.Empty(t, got)

		err = client.JSONSet(ctx, "JSONArrAppend-1", "$", `{"colors":[]}`)
		xt.NoError(t, err)

		got, err = client.JSONArrAppend(ctx, "JSONArrAppend-1", "$.colors", `"blue"`)
		xt.NoError(t, err)
		xt.NotEmpty(t, got)
		xt.Len(t, got, 1)
		xt.Equal(t, 1, *got[0])

		got, err = client.JSONArrAppend(ctx, "JSONArrAppend-1", "$.colors", `"blue"`)
		xt.NoError(t, err)
		xt.Len(t, got, 1)
		xt.Equal(t, 2, *got[0])
	})

	t.Run("JSONArrIndex", func(t *testing.T) {
		got, err := client.JSONArrIndex(ctx, "JSONArrIndex-1", "$..colors", `"silver"`)
		//  ERR Path '$..colors' does not exist
		xt.Error(t, err)
		xt.True(t, xerror.IsNotFound(err))
		xt.Empty(t, got)

		err = client.JSONSet(ctx, "JSONArrIndex-1", "$", `{"colors":["blue","silver"]}`)
		xt.NoError(t, err)

		got, err = client.JSONArrIndex(ctx, "JSONArrIndex-1", "$..colors", `"silver"`)
		xt.NoError(t, err)
		xt.Len(t, got, 1)
		xt.Equal(t, 1, *got[0])

		got, err = client.JSONArrIndexRange(ctx, "JSONArrIndex-1", "$..colors", `"silver"`, 0, 10)
		xt.NoError(t, err)
		xt.Len(t, got, 1)
		xt.Equal(t, 1, *got[0])
	})

	t.Run("JSONArrInsert", func(t *testing.T) {
		got, err := client.JSONArrInsert(ctx, "JSONArrInsert-1", "$", 0, 123)
		// ERR could not perform this operation on a key that doesn't exist
		xt.Error(t, err)
		xt.True(t, xerror.IsNotFound(err))
		xt.Empty(t, got)

		err = client.JSONSet(ctx, "JSONArrInsert-1", "$", `{"colors":["blue","silver"]}`)
		xt.NoError(t, err)

		got, err = client.JSONArrInsert(ctx, "JSONArrInsert-1", "$.colors", 0, `"hello"`)
		xt.NoError(t, err)
		xt.Len(t, got, 1)
		xt.Equal(t, 3, *got[0])

		gotNums, err := client.JSONArrLen(ctx, "JSONArrInsert-1", "$.colors")
		xt.NoError(t, err)
		xt.Len(t, gotNums, 1)
		xt.Equal(t, 3, *gotNums[0])

		gotPop, err := client.JSONArrPop(ctx, "JSONArrInsert-1", &JSONArrPopOption{Path: "$.colors"})
		xt.NoError(t, err)
		xt.Len(t, gotPop, 1)
		xt.Equal(t, `"silver"`, *gotPop[0])

		gotLen, err := client.JSONClearPath(ctx, "JSONArrInsert-1", "$.colors")
		xt.NoError(t, err)
		xt.Equal(t, 1, gotLen)

		gotDelNum, err := client.JSONDel(ctx, "JSONArrInsert-1")
		xt.NoError(t, err)
		xt.Equal(t, 1, gotDelNum)
	})

	t.Run("JSONSet", func(t *testing.T) {
		err := client.JSONSet(ctx, "JSONSet-1", "$", `{"a":"a1"}`)
		xt.NoError(t, err)

		err = client.JSONSet(ctx, "JSONSet-2", "$.x", 5)
		xt.Error(t, err)
		xt.ErrorContains(t, err, "ERR new objects must be created at the root")
	})

	t.Run("JSONSetNX", func(t *testing.T) {
		got, err := client.JSONSetNX(ctx, "JSONSetNX-1", "$", `{"a":"a1"}`)
		xt.NoError(t, err)
		xt.True(t, got)

		// path 已存在
		got, err = client.JSONSetNX(ctx, "JSONSetNX-1", "$.a", `"abc"`)
		xt.Error(t, err)
		xt.ErrorIs(t, err, ErrNil)
		xt.False(t, got)

		got, err = client.JSONSetNX(ctx, "JSONSetNX-2", "$.x", 5)
		xt.Error(t, err)
		xt.False(t, got)
		xt.ErrorContains(t, err, "ERR new objects must be created at the root")
	})

	t.Run("JSONSetXX", func(t *testing.T) {
		got, err := client.JSONSetXX(ctx, "JSONSetXX-1", "$", `{"a":"a1"}`)
		xt.Error(t, err)
		xt.ErrorIs(t, err, ErrNil)
		xt.False(t, got)

		err = client.JSONSet(ctx, "JSONSetXX-1", "$", `{"a":"a1"}`)
		xt.NoError(t, err)

		got, err = client.JSONSetXX(ctx, "JSONSetXX-1", "$.a", `"abc"`)
		xt.NoError(t, err)
		xt.True(t, got)

		gotStr, err := client.JSONGet(ctx, "JSONSetXX-1")
		xt.NoError(t, err)
		xt.NotEmpty(t, gotStr)

		opt1 := &JSONGetOption{
			Indent:  " ",
			NewLine: "\n",
			Space:   " ",
		}
		gotStr, err = client.JSONGetWithOption(ctx, "JSONSetXX-1", opt1)
		xt.NoError(t, err)
		xt.NotEmpty(t, gotStr)
	})

	t.Run("JSONObjKeys", func(t *testing.T) {
		err := client.JSONSet(ctx, "JSONObjKeys-1", "$", `{"a":"a1","b":"hello"}`)
		xt.NoError(t, err)

		keys, err := client.JSONObjKeys(ctx, "JSONObjKeys-1", "$")
		xt.NoError(t, err)
		xt.Equal(t, [][]string{{"a", "b"}}, keys)
	})

	t.Run("JSONStrAppend", func(t *testing.T) {
		got, err := client.JSONStrAppend(ctx, "JSONStrAppend-1", "", "hello")
		xt.Error(t, err)
		xt.True(t, xerror.IsNotFound(err))
		xt.Nil(t, got)

		err = client.JSONSet(ctx, "JSONStrAppend-1", "$", `{"a":2, "nested": {"a": true},"z":{"a":"hello"}}`)
		xt.NoError(t, err)

		got, err = client.JSONStrAppend(ctx, "JSONStrAppend-1", ".z.a", `"a"`)
		xt.NoError(t, err)
		num := int64(6)
		xt.Equal(t, []*int64{&num}, got)

		got, err = client.JSONStrAppend(ctx, "JSONStrAppend-1", "$..a", `"a"`)
		xt.NoError(t, err)
		num++
		xt.Equal(t, []*int64{nil, nil, &num}, got)
	})

	t.Run("JSONStrLenWithPath", func(t *testing.T) {
		got, err := client.JSONStrLenWithPath(ctx, "JSONStrLenWithPath-1", ".a")
		xt.NoError(t, err)
		xt.Equal(t, []*int64{nil}, got)

		err = client.JSONSet(ctx, "JSONStrLenWithPath-1", "$", `{"a":2, "nested": {"a": true},"z":{"a":"hello"}}`)
		xt.NoError(t, err)

		got, err = client.JSONStrLenWithPath(ctx, "JSONStrLenWithPath-1", ".a")
		// WRONGTYPE wrong type of path value - expected string but found integer
		xt.Error(t, err)
		xt.Nil(t, got)

		got, err = client.JSONStrLenWithPath(ctx, "JSONStrLenWithPath-1", "$..a")
		xt.NoError(t, err)
		num := int64(5)
		xt.Equal(t, []*int64{nil, nil, &num}, got)
	})

	t.Run("JSONStrLenWithPath-1", func(t *testing.T) {
		got, err := client.JSONStrLenWithPath(ctx, "JSONStrLenWithPathN-1", "$.a")
		xt.Error(t, err)
		xt.True(t, xerror.IsNotFound(err))
		xt.Nil(t, got)
	})

	t.Run("JSONStrLen", func(t *testing.T) {
		got, err := client.JSONStrLen(ctx, "JSONStrLen-1")
		xt.Error(t, err)
		xt.ErrorIs(t, err, ErrNil)
		xt.Equal(t, 0, got)

		err = client.JSONSet(ctx, "JSONStrLen-1", "$", `{"a":2, "nested": {"a": true}}`)
		xt.NoError(t, err)

		got, err = client.JSONStrLen(ctx, "JSONStrLen-1")
		// WRONGTYPE wrong type of path value - expected string but found object
		xt.Error(t, err)
		xt.Empty(t, got)

		err = client.JSONSet(ctx, "JSONStrLen-1", "$", `"hello"`)
		xt.NoError(t, err)
		got, err = client.JSONStrLen(ctx, "JSONStrLen-1")
		xt.NoError(t, err)
		xt.Equal(t, 5, got)
	})

	t.Run("JSONToggle-1", func(t *testing.T) {
		got, err := client.JSONToggle(ctx, "JSONToggle-not-found", ".a")
		// ERR could not perform this operation on a key that doesn't exist
		xt.Error(t, err)
		xt.True(t, xerror.IsNotFound(err))
		xt.Nil(t, got)

		err = client.JSONSet(ctx, "JSONToggleOne-1", "$", `{"a":2, "nested": {"a": true}}`)
		xt.NoError(t, err)

		got, err = client.JSONToggle(ctx, "JSONToggleOne-1", ".a")
		// ERR Path '$.a' does not exist or not a bool
		xt.Error(t, err)
		xt.Nil(t, got)

		got, err = client.JSONToggle(ctx, "JSONToggleOne-1", ".nested.a")
		// 返回字符串：“false”
		xt.NoError(t, err)
		xt.NotNil(t, got)
		xt.False(t, *got[0])

		// 再反转一次
		got, err = client.JSONToggle(ctx, "JSONToggleOne-1", ".nested.a")
		xt.NoError(t, err)
		xt.NotNil(t, got)
		xt.True(t, *got[0]) //  <----
	})
	t.Run("JSONToggle-2", func(t *testing.T) {
		got, err := client.JSONToggle(ctx, "JSONToggleN-1", "$.a")
		// ERR could not perform this operation on a key that doesn't exist
		xt.Error(t, err)
		xt.Nil(t, got)

		err = client.JSONSet(ctx, "JSONToggleN-1", "$", `{"a":2, "nested": {"a": true}}`)
		xt.NoError(t, err)

		got, err = client.JSONToggle(ctx, "JSONToggleN-1", "$.a")
		xt.NoError(t, err)
		xt.Equal(t, []*bool{nil}, got)

		got, err = client.JSONToggle(ctx, "JSONToggleN-1", "$..a")
		xt.NoError(t, err)
		ok := false
		xt.Equal(t, []*bool{nil, &ok}, got)
	})

	t.Run("JSONType", func(t *testing.T) {
		got, err := client.JSONType(ctx, "JSONType-1-not-found")
		xt.NoError(t, err)
		xt.Equal(t, []any{nil}, got)

		err = client.JSONSet(ctx, "JSONType-1", "$", `{"a":2, "nested": {"a": true}, "foo": "bar","z":{"a":[1]},"x":{"a":1.2}}`)
		xt.NoError(t, err)

		got, err = client.JSONType(ctx, "JSONType-1")
		xt.NoError(t, err)
		xt.Equal(t, []any{"object"}, got)

		got1, err := client.JSONTypeWithPath(ctx, "JSONType-1", "$.a")
		xt.NoError(t, err)
		xt.Equal(t, [][]any{{"integer"}}, got1)

		got2, err := client.JSONTypeWithPath(ctx, "JSONType-1", "$..a")
		xt.NoError(t, err)
		xt.Equal(t, [][]any{{"integer", "boolean", "array", "number"}}, got2)
	})

	t.Run("JSONNumIncrBy", func(t *testing.T) {
		got, err := client.JSONNumIncrBy(ctx, "JSONNumIncrBy-1", "$.a", 2)
		// ERR could not perform this operation on a key that doesn't exist
		xt.Error(t, err)
		xt.True(t, xerror.IsNotFound(err))
		xt.Nil(t, got)

		err = client.JSONSet(ctx, "JSONNumIncrBy-1", "$", `{"a":2, "nested": {"a": true}, "foo": "bar","z":{"a":[1]},"x":{"a":1.2}}`)
		xt.NoError(t, err)

		got, err = client.JSONNumIncrBy(ctx, "JSONNumIncrBy-1", "$.a", 2)
		xt.NoError(t, err)
		xt.Equal(t, []any{int64(4)}, got)

		got, err = client.JSONNumIncrBy(ctx, "JSONNumIncrBy-1", "$..a", 2)
		xt.NoError(t, err)
		xt.Equal(t, []any{int64(6), nil, nil, float64(3.2)}, got)
	})

	t.Run("JSONNumMultBy", func(t *testing.T) {
		got, err := client.JSONNumMultBy(ctx, "JSONNumMultBy-1", "$.a", 2)
		// ERR could not perform this operation on a key that doesn't exist
		xt.Error(t, err)
		xt.True(t, xerror.IsNotFound(err))
		xt.Nil(t, got)

		err = client.JSONSet(ctx, "JSONNumMultBy-1", "$", `{"a":2, "nested": {"a": true}, "foo": "bar","z":{"a":[1]},"x":{"a":1.2}}`)
		xt.NoError(t, err)

		got, err = client.JSONNumMultBy(ctx, "JSONNumMultBy-1", "$.a", 2)
		xt.NoError(t, err)
		xt.Equal(t, []any{int64(4)}, got)

		got, err = client.JSONNumMultBy(ctx, "JSONNumMultBy-1", "$..a", 2)
		xt.NoError(t, err)
		xt.Equal(t, []any{int64(8), nil, nil, float64(2.4)}, got)
	})

	t.Run("JSONObjLen", func(t *testing.T) {
		got, err := client.JSONObjLen(ctx, "JSONObjLen-1", "")
		xt.NoError(t, err)
		xt.Equal(t, []*int64{nil}, got)
	})
}
