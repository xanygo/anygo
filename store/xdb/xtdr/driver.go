//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-07

package xtdr

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"slices"
	"sync"
	"time"

	"github.com/xanygo/anygo/ds/xstr"
)

// ExpectQuery 提前预埋查询请求( db.Query )的 sql 对应的结果集
func ExpectQuery(query string, columns []string, rows [][]any) {
	mu.Lock()
	defer mu.Unlock()
	q := &queryExpectation{
		query:   query,
		columns: columns,
		rows:    convertRows(rows),
		typ:     expectQuery,
	}
	expectations = append(expectations, q)
}

// ExpectExec 提前预埋提更新请求（ db.Exec ）的 sql 对应的结果集
func ExpectExec(query string, res driver.Result, err error) {
	mu.Lock()
	defer mu.Unlock()
	e := &queryExpectation{
		query: query,
		typ:   expectExec,
		res:   res,
		err:   err,
	}
	expectations = append(expectations, e)
}

func ResultOf(lastInsertID, rowsAffected int64) driver.Result {
	return driverResult{lastInsertID: lastInsertID, rowsAffected: rowsAffected}
}

func LastQueries() []string {
	mu.Lock()
	defer mu.Unlock()
	return slices.Clone(executedQueries)
}

// Reset clears expectations and recorded queries.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	expectations = nil
	executedQueries = nil
}

const Name = "test-driver"

var once sync.Once

func Register() {
	once.Do(func() {
		sql.Register(Name, &Driver{})
	})
}

type driverType int

const (
	expectQuery driverType = iota
	expectExec
)

type queryExpectation struct {
	query   string
	columns []string
	rows    [][]driver.Value // for query
	typ     driverType
	res     driver.Result // for exec
	err     error
	used    bool
}

var (
	mu              sync.Mutex
	expectations    []*queryExpectation
	executedQueries []string
)

// convertRows converts [][]any to [][]driver.Value
func convertRows(rows [][]any) [][]driver.Value {
	out := make([][]driver.Value, 0, len(rows))
	for _, r := range rows {
		row := make([]driver.Value, len(r))
		for i, v := range r {
			row[i] = convertToDriverValue(v)
		}
		out = append(out, row)
	}
	return out
}

func convertToDriverValue(v any) driver.Value {
	// database/sql/driver accepts: int64, float64, bool, []byte, string, time.Time, nil
	switch x := v.(type) {
	case nil:
		return nil
	case int:
		return int64(x)
	case int8:
		return int64(x)
	case int16:
		return int64(x)
	case int32:
		return int64(x)
	case int64:
		return x
	case uint:
		return int64(x)
	case uint8:
		return int64(x)
	case uint16:
		return int64(x)
	case uint32:
		return int64(x)
	case float32:
		return float64(x)
	case float64:
		return x
	case bool:
		return x
	case string:
		return x
	case []byte:
		return x
	case time.Time:
		return x
	default:
		return fmt.Sprint(v)
	}
}

var _ driver.Driver = (*Driver)(nil)

// Driver implements driver.Driver
type Driver struct{}

func (d *Driver) Open(name string) (driver.Conn, error) {
	// name ignored
	return &conn{}, nil
}

var _ driver.Conn = (*conn)(nil)

// conn implements driver.Conn and Prepare/Begin/Close
type conn struct{}

func (c *conn) Prepare(query string) (driver.Stmt, error) {
	return &stmt{query: query}, nil
}

func (c *conn) Close() error { return nil }

func (c *conn) Begin() (driver.Tx, error) {
	return tx{}, nil
}

var _ driver.Stmt = (*stmt)(nil)

// stmt implements driver.Stmt
type stmt struct {
	query string
}

func (s *stmt) Close() error { return nil }

func (s *stmt) NumInput() int { return -1 } // -1 means variable

func (s *stmt) Exec(args []driver.Value) (driver.Result, error) {
	return execSQL(s.query)
}

func (s *stmt) Query(args []driver.Value) (driver.Rows, error) {
	return querySQL(s.query)
}

var _ driver.Tx = (*tx)(nil)

type tx struct{}

func (tx) Commit() error { return nil }

func (tx) Rollback() error { return nil }

var _ driver.Rows = (*rows)(nil)

type rows struct {
	cols []string
	data [][]driver.Value
	pos  int
}

func (r *rows) Columns() []string {
	return r.cols
}

func (r *rows) Next(dest []driver.Value) error {
	if r.pos >= len(r.data) {
		return io.EOF
	}
	row := r.data[r.pos]
	if len(dest) != len(row) {
		return fmt.Errorf("want %d destination values, got %d", len(dest), len(row))
	}

	copy(dest, row)
	r.pos++
	return nil
}

func (r *rows) Close() error { return nil }

var _ driver.Result = (*driverResult)(nil)

type driverResult struct {
	lastInsertID int64
	rowsAffected int64
}

func (d driverResult) LastInsertId() (int64, error) {
	return d.lastInsertID, nil
}

func (d driverResult) RowsAffected() (int64, error) {
	return d.rowsAffected, nil
}

func execSQL(query string) (driver.Result, error) {
	mu.Lock()
	defer mu.Unlock()
	executedQueries = append(executedQueries, query)
	for _, e := range expectations {
		if e.used || e.typ != expectExec {
			continue
		}
		if !xstr.Match(e.query, query) {
			continue
		}
		e.used = true
		return e.res, e.err
	}
	return nil, fmt.Errorf("no exec expectation for query: %q", query)
}

func querySQL(query string) (driver.Rows, error) {
	mu.Lock()
	defer mu.Unlock()
	executedQueries = append(executedQueries, query)
	for _, e := range expectations {
		if e.used || e.typ != expectQuery {
			continue
		}
		if !xstr.Match(e.query, query) {
			continue
		}
		e.used = true
		return &rows{
			cols: slices.Clone(e.columns),
			data: slices.Clone(e.rows),
			pos:  0,
		}, nil
	}
	return nil, fmt.Errorf("no query expectation for query: %q", query)
}

func init() {
	Register()
}
