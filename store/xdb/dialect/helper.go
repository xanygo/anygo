package dialect

import (
	"context"
	"errors"

	"github.com/xanygo/anygo/store/xdb/dbtype"
)

func queryOneString(ctx context.Context, q dbtype.Queryer, sql string, args ...any) (string, error) {
	rows, err := q.QueryContext(ctx, sql, args...)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			return "", err
		}
		return name, nil
	}
	return "", errors.New("no result")
}

func querySliceString(ctx context.Context, q dbtype.Queryer, sql string, args ...any) ([]string, error) {
	rows, err := q.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []string
	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			return nil, err
		}
		result = append(result, name)
	}
	return result, nil
}

func queryBool(ctx context.Context, q dbtype.Queryer, sql string, args ...any) (bool, error) {
	rows, err := q.QueryContext(ctx, sql, args...)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	var exists bool
	err = rows.Scan(&exists)
	return exists, err
}
