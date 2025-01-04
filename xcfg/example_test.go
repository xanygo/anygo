//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package xcfg_test

import (
	"fmt"
	"log"

	"github.com/xanygo/anygo/xcfg"
)

func ExampleExists() {
	// 配置文件 {ConfDir}/abc.json 存在：
	fmt.Println(xcfg.Exists("abc"))      // true
	fmt.Println(xcfg.Exists("abc.json")) // true

	// 配置文件 {ConfDir}/not-found.json 不存在：
	fmt.Println(xcfg.Exists("not-found"))      // false
	fmt.Println(xcfg.Exists("not-found.json")) // false

	// Output:
	// true
	// true
	// false
	// false
}

func ExampleMustParse() {
	type Info struct {
		A string
	}
	var info Info

	// 解析 配置文件 {ConfDir}/abc.json，若失败会 panic
	xcfg.MustParse("abc.json", &info)

	fmt.Printf("info.A = %q\n", info.A)

	// Output:
	// info.A = "bb"
}

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

func ExampleMustParseBytes() {
	type User struct {
		Name string
		Age  int
	}
	content := []byte(`{"Name":"Hello","age":18}`)

	var user *User

	xcfg.MustParseBytes(".json", content, &user) // 解析失败会 panic

	fmt.Println("Name=", user.Name)
	fmt.Println("Age=", user.Age)
	// OutPut:
	// Name= Hello
	// Age= 18
}

func ExampleParse() {
	type Info struct {
		A string
	}
	var info Info

	// 解析 配置文件 {ConfDir}/abc.json
	if err := xcfg.Parse("abc.json", &info); err != nil {
		log.Println(err.Error())
	}

	fmt.Printf("info.A = %q\n", info.A)

	// Output:
	// info.A = "bb"
}
