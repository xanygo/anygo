//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-10

package xdb

import (
	"context"
	"database/sql"
	"fmt"
	"io"

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

func (c *Client) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return c.db.QueryContext(ctx, query, args...)
}

func (c *Client) BeginTx(ctx context.Context, opts *sql.TxOptions) (TxExecutor, error) {
	t, err := c.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &tx{Raw: t, driver: c.driver}, nil
}

var _ Execer = (*Client)(nil)

func (c *Client) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return c.db.ExecContext(ctx, query, args...)
}

var _ Preparer = (*Client)(nil)

func (c *Client) PrepareContext(ctx context.Context, query string) (Statement, error) {
	s, err := c.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return &stmt{Raw: s, driver: c.driver, query: query}, nil
}

var _ RowQuerier = (*Client)(nil)

func (c *Client) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return c.db.QueryRowContext(ctx, query, args...)
}

var _ io.Closer = (*Client)(nil)

func (c *Client) Close() error {
	return c.db.Close()
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
