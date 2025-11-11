//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-07

package xdb_test

import (
	"testing"

	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/xt"
)

func TestEncode(t *testing.T) {
	user1 := &testUser1{
		sid:    100,
		ID:     1,
		Name:   "name1",
		Enable: true,
		Score:  120.1,
		IDs1:   []int{1, 2, 3},
		IDs2:   []int{1, 5, 3},
		IDs3:   nil,
		Md1:    nil,
		Md2: map[string]string{
			"key1": "value1",
		},
	}
	// id := 1
	// user1.ID2 = &id
	out1, err := xdb.Encode(user1)
	xt.NoError(t, err)
	t.Logf("out: %#v", out1)
	xt.NotEmpty(t, out1)
	want1 := map[string]any{
		"id":     1,
		"name":   "name1",
		"enable": true,
		"score":  120.1,
		"ids1":   "[1,2,3]",
		"ids2":   "[1,5,3]",
		"md1":    "", // 本来是 null
		"md2":    `{"key1":"value1"}`,
		"bs1":    "",
		"id2":    0,
	}
	xt.Equal(t, want1, out1)
}

type testUser1 struct {
	sid    int
	ID     int               `db:"id"`
	Name   string            `db:"name"`
	Enable bool              `db:"enable"`
	Score  float64           `db:"score"`
	IDs1   []int             `db:"ids1,codec:json"`
	IDs2   []int             `db:"ids2,codec:json"`
	IDs3   []int             // 没有定义 db tag，会被忽略
	Md1    map[string]any    `db:"md1,codec:json"`
	Md2    map[string]string `db:"md2,codec:json"`

	Bs1 []byte `db:"bs1"`
	ID2 *int   `db:"id2"`
}
