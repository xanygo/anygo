//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package dialect

import (
	"fmt"

	"github.com/xanygo/anygo/store/xdb/dbtype"
	"github.com/xanygo/anygo/xerror"
)

var dialects = map[string]dbtype.Dialect{}

func Register(d dbtype.Dialect) bool {
	return RegisterWithName(d.Name(), d)
}

func RegisterWithName(name string, d dbtype.Dialect) bool {
	if _, ok := dialects[name]; ok {
		return false
	}
	dialects[name] = d
	return true
}

func Find(name string) (dbtype.Dialect, error) {
	d, ok := dialects[name]
	if ok {
		return d, nil
	}
	return nil, fmt.Errorf("dialect %q %w", name, xerror.NotFound)
}

func init() {
	Register(MySQL{})
	Register(MariaDB{})
	Register(Postgres{})
	RegisterWithName("pgx", Postgres{})
	Register(SQLite3{})
	Register(SQLServerDialect{})
}
