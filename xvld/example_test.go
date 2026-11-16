package xvld_test

import (
	"fmt"

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

	u2 := &User1{
		Name: "hello",
		Age:  1,
	}
	fmt.Println("u2=", xvld.Struct(u2))

	// Output:
	// u1= Field 'Name': is a required field
	// u2= Field 'Age': must be 10 or greater, got 1
}
