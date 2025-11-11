//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package dialect

import (
	"fmt"

	"github.com/xanygo/anygo/xerror"
)

var dialects = map[string]Dialect{}

func Register(d Dialect) {
	RegisterWithName(d.Name(), d)
}

func RegisterWithName(name string, d Dialect) {
	dialects[name] = d
}

func Find(name string) (Dialect, error) {
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
	Register(SQLite{})
	Register(SQLServerDialect{})
}
