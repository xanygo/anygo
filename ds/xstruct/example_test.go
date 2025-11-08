//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-07

package xstruct_test

import (
	"fmt"

	"github.com/xanygo/anygo/ds/xstruct"
)

func ExampleParserTag() {
	tag1 := xstruct.ParserTag(`name,omitempty`)

	fmt.Println("Name:", tag1.Name())                  // name
	fmt.Printf("Value: %q\n", tag1.Value("omitempty")) // ""
	fmt.Printf("Has: %v\n", tag1.Has("omitempty"))     // true
	// Output:
	// Name: name
	// Value: ""
	// Has: true
}
