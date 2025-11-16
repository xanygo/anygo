//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-15

package xdb

import "github.com/xanygo/anygo/store/xdb/dbschema"

func SetTagName(name string) {
	dbschema.SetTagName(name)
}

func TagName() string {
	return dbschema.TagName()
}
