//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-10

package xdb

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"time"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/ds/xoption"
	"github.com/xanygo/anygo/ds/xstr"
	"github.com/xanygo/anygo/xnet/xservice"
)

type FactoryFunc func(ctx context.Context) (*sql.DB, error)

func NewClient(driver string, name string, db *sql.DB) *Client {
	return &Client{
		driver: driver,
		name:   name,
		db:     db,
	}
}

func NewClientWithService(name any) (*Client, error) {
	srv, err := xservice.FindService(name)
	if err != nil {
		return nil, err
	}
	opt := srv.Option()
	const key = "Database"
	data := map[string]any{
		"Network":      xservice.Network,
		"HOST_PORT":    srv.Name(),
		"ReadTimeout":  xoption.ReadTimeout(opt).String(),
		"WriteTimeout": xoption.WriteTimeout(opt).String(),
		"Timeout":      xoption.TotalTimeout(opt).String(),
	}
	var driver, dsn string
	xmap.Range[string, string](xoption.Extra(srv.Option(), key), func(k, v string) bool {
		switch k {
		case "Driver":
			driver = v
		case "Username":
			data["Username"] = v
		case "Password":
			data["Password"] = v
		case "DSN":
			dsn = v
		}
		return true
	})
	if driver == "" {
		return nil, fmt.Errorf("%s[Driver] missing", key)
	}
	if dsn == "" {
		return nil, fmt.Errorf("%s[DSN] missing", key)
	}
	dsn, err = xstr.RenderTemplate(dsn, data)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	return NewClient(driver, srv.Name(), db), nil
}

type HasDriver interface {
	Driver() string
}

type Client struct {
	name   string
	driver string
	db     *sql.DB
}

func (c *Client) Name() string {
	return c.name
}

// Driver 驱动名称，同时也是方言名称
func (c *Client) Driver() string {
	return c.driver
}

var _ Queryer = (*Client)(nil)

func (c *Client) QueryContext(ctx context.Context, query string, args ...any) (rows *sql.Rows, err error) {
	its := allInterceptors(ctx)
	if len(its) > 0 {
		event := Event{
			Action: "Query",
			Start:  time.Now(),
			Client: c.Name(),
			Driver: c.Driver(),
			Query:  query,
			Args:   args,
		}
		defer func() {
			event.End = time.Now()
			event.Error = err
			its.CallAfter(ctx, event)
		}()
	}
	return c.db.QueryContext(ctx, query, args...)
}

func (c *Client) BeginTx(ctx context.Context, opts *sql.TxOptions) (te TxExecutor, err error) {
	var txID string
	its := allInterceptors(ctx)
	if len(its) > 0 {
		txID = xstr.RandNChar(5)
		event := Event{
			Action: "BeginTx",
			Start:  time.Now(),
			Client: c.Name(),
			Driver: c.Driver(),
			TxID:   txID,
		}
		defer func() {
			event.End = time.Now()
			event.Error = err
			its.CallAfter(ctx, event)
		}()
	}
	t, err := c.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &myTx{Raw: t, client: c, txID: txID, its: its, ctx: ctx}, nil
}

var _ Execer = (*Client)(nil)

func (c *Client) ExecContext(ctx context.Context, query string, args ...any) (ret sql.Result, err error) {
	its := allInterceptors(ctx)
	if len(its) > 0 {
		event := Event{
			Action: "Exec",
			Start:  time.Now(),
			Client: c.Name(),
			Driver: c.Driver(),
			Query:  query,
			Args:   args,
		}
		defer func() {
			event.End = time.Now()
			event.Error = err
			its.CallAfter(ctx, event)
		}()
	}
	return c.db.ExecContext(ctx, query, args...)
}

var _ Preparer = (*Client)(nil)

func (c *Client) PrepareContext(ctx context.Context, query string) (ns Statement, err error) {
	its := allInterceptors(ctx)
	var stmtID string
	if len(its) > 0 {
		stmtID = xstr.RandNChar(5)
		event := Event{
			Action: "Prepare",
			Start:  time.Now(),
			Client: c.Name(),
			Driver: c.Driver(),
			StmtID: stmtID,
			Query:  query,
		}
		defer func() {
			event.End = time.Now()
			event.Error = err
			its.CallAfter(ctx, event)
		}()
	}

	s, err := c.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return &myStmt{Raw: s, client: c, query: query, stmtID: stmtID}, nil
}

var _ RowQuerier = (*Client)(nil)

func (c *Client) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	its := allInterceptors(ctx)
	if len(its) > 0 {
		event := Event{
			Action: "QueryRow",
			Start:  time.Now(),
			Client: c.Name(),
			Driver: c.Driver(),
			Query:  query,
			Args:   args,
		}
		defer func() {
			event.End = time.Now()
			its.CallAfter(ctx, event)
		}()
	}
	return c.db.QueryRowContext(ctx, query, args...)
}

var _ io.Closer = (*Client)(nil)

func (c *Client) Close() error {
	return c.db.Close()
}

var _ TxExecutor = (*myTx)(nil)
var _ HasDriver = (*myTx)(nil)

type myTx struct {
	Raw    *sql.Tx
	client *Client
	txID   string
	its    interceptors
	ctx    context.Context // 创建 myTx 时候的 ctx
}

func (t *myTx) Driver() string {
	return t.client.Driver()
}

func (t *myTx) QueryContext(ctx context.Context, query string, args ...any) (rows *sql.Rows, err error) {
	if len(t.its) > 0 {
		event := Event{
			Action: "Query",
			Start:  time.Now(),
			Driver: t.Driver(),
			Client: t.client.Name(),
			Query:  query,
			Args:   args,
			TxID:   t.txID,
		}
		defer func() {
			event.End = time.Now()
			event.Error = err
			t.its.CallAfter(ctx, event)
		}()
	}
	return t.Raw.QueryContext(ctx, query, args...)
}

func (t *myTx) ExecContext(ctx context.Context, query string, args ...any) (ret sql.Result, err error) {
	if len(t.its) > 0 {
		event := Event{
			Action: "Exec",
			Start:  time.Now(),
			Driver: t.Driver(),
			Client: t.client.Name(),
			Query:  query,
			Args:   args,
			TxID:   t.txID,
		}
		defer func() {
			event.End = time.Now()
			event.Error = err
			t.its.CallAfter(ctx, event)
		}()
	}
	return t.Raw.ExecContext(ctx, query, args...)
}

func (t *myTx) PrepareContext(ctx context.Context, query string) (ns Statement, err error) {
	var stmtID string
	if len(t.its) > 0 {
		stmtID = xstr.RandNChar(5)
		event := Event{
			Action: "Exec",
			Start:  time.Now(),
			Driver: t.Driver(),
			Client: t.client.Name(),
			Query:  query,
			TxID:   t.txID,
			StmtID: stmtID,
		}
		defer func() {
			event.End = time.Now()
			event.Error = err
			t.its.CallAfter(ctx, event)
		}()
	}
	s, err := t.Raw.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return &myStmt{Raw: s, query: query, stmtID: stmtID, txID: t.txID}, nil
}

func (t *myTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	if len(t.its) > 0 {
		event := Event{
			Action: "QueryRow",
			Start:  time.Now(),
			Driver: t.Driver(),
			Client: t.client.Name(),
			Query:  query,
			Args:   args,
			TxID:   t.txID,
		}
		defer func() {
			event.End = time.Now()
			t.its.CallAfter(ctx, event)
		}()
	}
	return t.Raw.QueryRowContext(ctx, query, args...)
}

func (t *myTx) StmtContext(ctx context.Context, s Statement) Statement {
	st := t.Raw.StmtContext(ctx, s.Unwrap())
	nst := &myStmt{
		Raw:    st,
		txID:   t.txID,
		client: t.client,
	}
	if hq, ok := s.(hasSQLQuery); ok {
		nst.query = hq.SQLQuery()
	}
	return nst
}

func (t *myTx) Commit() (err error) {
	if len(t.its) > 0 {
		event := Event{
			Action: "Commit",
			Start:  time.Now(),
			Driver: t.Driver(),
			TxID:   t.txID,
		}
		defer func() {
			event.End = time.Now()
			event.Error = err
			t.its.CallAfter(t.ctx, event)
		}()
	}
	return t.Raw.Commit()
}

func (t *myTx) Rollback() (err error) {
	if len(t.its) > 0 {
		event := Event{
			Action: "Rollback",
			Start:  time.Now(),
			Driver: t.Driver(),
			TxID:   t.txID,
		}
		defer func() {
			event.End = time.Now()
			event.Error = err
			t.its.CallAfter(t.ctx, event)
		}()
	}
	return t.Raw.Rollback()
}

var _ Statement = (*myStmt)(nil)

type myStmt struct {
	Raw    *sql.Stmt
	client *Client
	query  string
	txID   string
	stmtID string
}

var _ HasDriver = (*myStmt)(nil)

func (s *myStmt) Driver() string {
	return s.client.Driver()
}

func (s *myStmt) Unwrap() *sql.Stmt {
	return s.Raw
}

type hasSQLQuery interface {
	SQLQuery() string
}

var _ hasSQLQuery = (*myStmt)(nil)

func (s *myStmt) SQLQuery() string {
	return s.query
}

func (s *myStmt) QueryContext(ctx context.Context, args ...any) (rows *sql.Rows, err error) {
	its := allInterceptors(ctx)
	if len(its) > 0 {
		event := Event{
			Action: "StmtQuery",
			Start:  time.Now(),
			Driver: s.Driver(),
			Client: s.client.Name(),
			TxID:   s.txID,
			StmtID: s.stmtID,
			Query:  s.query,
			Args:   args,
		}
		defer func() {
			event.End = time.Now()
			event.Error = err
			its.CallAfter(ctx, event)
		}()
	}
	return s.Raw.QueryContext(ctx, args...)
}

func (s *myStmt) ExecContext(ctx context.Context, args ...any) (ret sql.Result, err error) {
	its := allInterceptors(ctx)
	if len(its) > 0 {
		event := Event{
			Action: "StmtExec",
			Start:  time.Now(),
			Driver: s.Driver(),
			Client: s.client.Name(),
			TxID:   s.txID,
			StmtID: s.stmtID,
			Query:  s.query,
			Args:   args,
		}
		defer func() {
			event.End = time.Now()
			event.Error = err
			its.CallAfter(ctx, event)
		}()
	}
	return s.Raw.ExecContext(ctx, args...)
}

func (s *myStmt) QueryRowContext(ctx context.Context, args ...any) *sql.Row {
	its := allInterceptors(ctx)
	if len(its) > 0 {
		event := Event{
			Action: "StmtQueryRow",
			Start:  time.Now(),
			Driver: s.Driver(),
			Client: s.client.Name(),
			TxID:   s.txID,
			StmtID: s.stmtID,
			Query:  s.query,
			Args:   args,
		}
		defer func() {
			event.End = time.Now()
			its.CallAfter(ctx, event)
		}()
	}
	return s.Raw.QueryRowContext(ctx, args...)
}

func (s *myStmt) Close() error {
	return s.Raw.Close()
}
