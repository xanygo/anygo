//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package xcfg_test

import (
	"fmt"
	"log"

	"github.com/xanygo/anygo/xcfg"
)

func ExampleParseBytes() {
	type User struct {
		Name string
		Age  int
	}
	content := []byte(`{"Name":"Hello","age":18}`)

	var user *User
	if err := xcfg.ParseBytes(".json", content, &user); err != nil {
		log.Fatalln("ParseBytes with error:", err)
	}
	fmt.Println("Name=", user.Name)
	fmt.Println("Age=", user.Age)
	// OutPut:
	// Name= Hello
	// Age= 18
}
