//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-07

package xdb_test

import (
	"testing"

	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/xt"
)

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

type TestUser2 struct {
	CSV1 []int             `db:"csv1,codec:csv"`
	MP1  map[string]string `db:"mp1,codec:json"`
}

type TestUser21 struct {
	U21 string `db:"u21"`
	TestUser2
}

type TestUser22 struct {
	// 多层嵌套
	U22 string `db:"u22"`

	TestUser21
}

type testUser3 struct {
	Name      string `db:"name"`
	skip      string
	TestUser2 ``
}

type testUser4 struct {
	Name       string `db:"name"`
	skip       string
	*TestUser2 // 在 scanner 中，这个要求是可导出类型的
}

var _ = testUser4{skip: "ok"}

type testUser5 struct {
	CSV1 []int             `db:"csv1,codec:csv"`
	MP1  map[string]string `db:"mp1,codec:json"`
}

type testUser6 struct {
	Name       string `db:"name"`
	skip       string
	*testUser5 // 这个是不可导出类型，所以不能通过反射设置值
}

var _ = testUser6{skip: "ok"}

func TestEncode(t *testing.T) {
	t.Run("testUser1", func(t *testing.T) {
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
	})

	t.Run("testUser3", func(t *testing.T) {
		u3 := testUser3{
			Name: "name",
			skip: "skip",
			TestUser2: TestUser2{
				CSV1: []int{1, 2, 3},
				MP1:  map[string]string{"key1": "value1"},
			},
		}
		out1, err := xdb.Encode(u3)
		xt.NoError(t, err)
		t.Logf("out: %#v", out1)
		want := map[string]any{
			"name": "name",
			"csv1": "1,2,3",
			"mp1":  `{"key1":"value1"}`,
		}
		xt.Equal(t, want, out1)
	})
	t.Run("testUser4", func(t *testing.T) {
		u3 := testUser4{
			Name: "name",
			skip: "skip",
			TestUser2: &TestUser2{
				CSV1: []int{1, 2, 3},
				MP1:  map[string]string{"key1": "value1"},
			},
		}
		out1, err := xdb.Encode(u3)
		xt.NoError(t, err)
		t.Logf("out: %#v", out1)
		want := map[string]any{
			"name": "name",
			"csv1": "1,2,3",
			"mp1":  `{"key1":"value1"}`,
		}
		xt.Equal(t, want, out1)
	})

	t.Run("TestUser22", func(t *testing.T) {
		u22 := TestUser22{
			U22: "u22-value",
			TestUser21: TestUser21{
				U21: "u21-hello",
				TestUser2: TestUser2{
					CSV1: []int{1, 2, 3},
					MP1:  map[string]string{"key1": "value1"},
				},
			},
		}
		out1, err := xdb.Encode(u22)
		xt.NoError(t, err)
		t.Logf("out: %#v", out1)
		want := map[string]any{
			"u22":  "u22-value",
			"u21":  "u21-hello",
			"csv1": "1,2,3",
			"mp1":  `{"key1":"value1"}`,
		}
		xt.Equal(t, want, out1)
	})
}

// BenchmarkEncodeStruct-4           328579              3513 ns/op            3432 B/op         41 allocs/op
func BenchmarkEncodeStruct(b *testing.B) {
	u1 := testUser1{
		ID:   1000,
		Name: "name1",
		Md2:  map[string]string{"k": "v"},
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		xdb.Encode(u1)
	}
}
