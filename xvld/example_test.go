package xvld_test

import (
	"context"
	"fmt"

	"github.com/xanygo/anygo/xi18n"
	"github.com/xanygo/anygo/xvld"
)

type User1 struct {
	Name string `xv:"required,min-len=2"`
	Age  int    `xv:"min=10"`
}

func ExampleStruct() {
	u1 := &User1{
		Age: 18,
	}
	fmt.Println("u1=", xvld.Struct(u1))

	u2 := User1{
		Name: "hello",
		Age:  1,
	}
	fmt.Println("u2=", xvld.Struct(u2))

	// 通过 context携带语言信息：中文
	ctx := xi18n.ContextWithLanguages(context.Background(), xi18n.LangZh)
	fmt.Println("u2=", xvld.StructContext(ctx, u2))

	// Output:
	// u1= Field 'Name': empty value
	// u2= Field 'Age': must be 10 or greater, got 1
	// u2= 字段 Age: 值应不小于 10, 当前为 1
}
