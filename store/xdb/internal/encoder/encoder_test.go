//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-07

package encoder_test

import (
	"testing"
	"time"

	"github.com/xanygo/anygo/store/xdb/dbschema"
	"github.com/xanygo/anygo/store/xdb/dialect"
	"github.com/xanygo/anygo/store/xdb/internal/encoder"
	"github.com/xanygo/anygo/xt"
)

type testUser1 struct {
	sid    int
	ID     int               `db:"id"`
	Name   string            `db:"name"`
	Enable bool              `db:"enable"`
	Score  float64           `db:"score"`
	IDs1   []int             `db:"ids1,codec=json"`
	IDs2   []int             `db:"ids2,codec=json"`
	IDs3   []int             // 没有定义 db tag，会被忽略
	Md1    map[string]any    `db:"md1,codec=json"`
	Md2    map[string]string `db:"md2,codec=json"`

	Bs1 []byte `db:"bs1"`
	ID2 *int   `db:"id2"`
}

type TestUser2 struct {
	CSV1 []int             `db:"csv1,codec=csv"`
	MP1  map[string]string `db:"mp1,codec=json"`
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
	CSV1 []int             `db:"csv1,codec=csv"`
	MP1  map[string]string `db:"mp1,codec=json"`
}

type testUser6 struct {
	Name       string `db:"name"`
	skip       string
	*testUser5 // 这个是不可导出类型，所以不能通过反射设置值
}

var _ = testUser6{skip: "ok"}

type testUser7 struct {
	Name    string    `db:"name"`
	Number  int64     `db:"number,auto=Incr"`
	Leave   time.Time `db:"leave,auto=Now"`
	Created time.Time `db:"c,auto=Created"`
	Updated int64     `db:"u,auto=Updated"`
	Version *int64    `db:"v"`
}

type testUser8 struct {
	Name string   `db:"name"`
	Num1 *int64   `db:"num1"`
	Num2 **int64  `db:"num2"`
	Num3 ***int64 `db:"num3"`
}

func TestEncodeInsert(t *testing.T) {
	dz := dialect.MySQL{}
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
		out1, err := encoder.EncodeInsert(dz, user1)
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
			"md1":    "null",
			"md2":    `{"key1":"value1"}`,
		}
		xt.Equal(t, out1, want1)
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
		out1, err := encoder.EncodeInsert(dz, u3)
		xt.NoError(t, err)
		t.Logf("out: %#v", out1)
		want := map[string]any{
			"name": "name",
			"csv1": "1,2,3",
			"mp1":  `{"key1":"value1"}`,
		}
		xt.Equal(t, out1, want)
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
		out1, err := encoder.EncodeInsert(dz, u3)
		xt.NoError(t, err)
		t.Logf("out: %#v", out1)
		want := map[string]any{
			"name": "name",
			"csv1": "1,2,3",
			"mp1":  `{"key1":"value1"}`,
		}
		xt.Equal(t, out1, want)
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
		out1, err := encoder.EncodeInsert(dz, u22)
		xt.NoError(t, err)
		t.Logf("out: %#v", out1)
		want := map[string]any{
			"u22":  "u22-value",
			"u21":  "u21-hello",
			"csv1": "1,2,3",
			"mp1":  `{"key1":"value1"}`,
		}
		xt.Equal(t, out1, want)
	})

	t.Run("testUser7-1", func(t *testing.T) {
		checkUser7 := func(t *testing.T, v any, name string) {
			t.Run(name, func(t *testing.T) {
				out1, err := encoder.EncodeInsert(dz, v)
				xt.NoError(t, err)
				t.Logf("out: %#v", out1)
				xt.Len(t, out1, 5)
				xt.NotEmpty(t, out1["name"])
				date := time.Now().Format("2006-01-02")
				xt.HasPrefix(t, out1["c"].(string), date)
				xt.NotEmpty(t, out1["u"])
				xt.Equal[any](t, out1["number"], int64(1))    // Incr
				xt.HasPrefix(t, out1["leave"].(string), date) // Now
			})
		}
		u3 := testUser7{
			Name: "name",
		}
		checkUser7(t, u3, "1-struct")
		checkUser7(t, &u3, "2-struct-ptr")
	})
	t.Run("testUser7-2", func(t *testing.T) {
		checkUser7 := func(t *testing.T, v any, name string) {
			t.Run(name, func(t *testing.T) {
				out1, err := encoder.EncodeInsert(dz, v)
				xt.NoError(t, err)
				t.Logf("out: %#v", out1)
				xt.Len(t, out1, 6)
				xt.NotEmpty(t, out1["name"])
				date := time.Now().Format("2006-01-02")
				xt.HasPrefix(t, out1["c"].(string), date)
				xt.NotEmpty(t, out1["u"])
				xt.Equal[any](t, out1["number"], int64(1))    // Incr
				xt.HasPrefix(t, out1["leave"].(string), date) // Now
			})
		}
		num := int64(2)
		u3 := testUser7{
			Name:    "name",
			Version: &num,
		}
		checkUser7(t, u3, "1-struct")
		checkUser7(t, &u3, "2-struct-ptr")
	})

	t.Run("testUser8", func(t *testing.T) {
		u1 := testUser8{
			Name: "name",
		}
		out1, err := encoder.EncodeInsert(dz, u1)
		xt.NoError(t, err)
		xt.Equal(t, out1, map[string]any{"name": "name"})

		num1 := int64(1)
		u2 := testUser8{
			Name: "name",
			Num1: &num1,
		}
		out2, err := encoder.EncodeInsert(dz, u2)
		xt.NoError(t, err)
		xt.Equal(t, out2, map[string]any{"name": "name", "num1": int64(1)})

		num2 := int64(2)
		num2Ptr := &num2
		u3 := testUser8{
			Name: "name",
			Num1: &num1,
			Num2: &num2Ptr,
		}
		out3, err := encoder.EncodeInsert(dz, u3)
		xt.NoError(t, err)
		xt.Equal(t, out3, map[string]any{"name": "name", "num1": int64(1), "num2": int64(2)})

		num3 := int64(3)
		num3Ptr := &num3
		num3PtrPtr := &num3Ptr
		u4 := testUser8{
			Name: "name",
			Num1: &num1,
			Num2: &num2Ptr,
			Num3: &num3PtrPtr,
		}
		out4, err := encoder.EncodeInsert(dz, u4)
		xt.NoError(t, err)
		xt.Equal(t, out4, map[string]any{"name": "name", "num1": int64(1), "num2": int64(2), "num3": int64(3)})
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
		encoder.EncodeInsert(dialect.MySQL{}, u1)
	}
}

type testUser9 struct {
	Name   string `db:"name"`
	Num1   *int64 `db:"num1"`
	Enable bool   `db:"enable"`
}

func TestEncoder_EncodeArgs(t *testing.T) {
	schema, err := dbschema.Schema(dialect.SQLite3{}, testUser9{})
	xt.NoError(t, err)
	enc := encoder.Encoder[testUser9]{
		Schema:  schema,
		Dialect: dialect.SQLite3{},
	}
	ok := true
	got, err := enc.EncodeArgs(1, true, nil, &ok, false)
	xt.NoError(t, err)
	xt.Equal(t, got, []any{1, 1, nil, 1, 0})
}
