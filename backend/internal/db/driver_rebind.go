package db

import (
	"context"
	"database/sql/driver"
)

// rebindDriver wraps a parent driver so application SQL can keep using ?.
type rebindDriver struct{ d driver.Driver }

func (r rebindDriver) Open(name string) (driver.Conn, error) {
	c, err := r.d.Open(name)
	if err != nil {
		return nil, err
	}
	return &rebindConn{Conn: c}, nil
}

type rebindConn struct{ driver.Conn }

func (c *rebindConn) Prepare(query string) (driver.Stmt, error) {
	return c.Conn.Prepare(RebindPostgres(query))
}

func (c *rebindConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if e, ok := c.Conn.(driver.ConnPrepareContext); ok {
		return e.PrepareContext(ctx, RebindPostgres(query))
	}
	return c.Prepare(query)
}

func (c *rebindConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if e, ok := c.Conn.(driver.ExecerContext); ok {
		return e.ExecContext(ctx, RebindPostgres(query), args)
	}
	return nil, driver.ErrSkip
}

func (c *rebindConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if q, ok := c.Conn.(driver.QueryerContext); ok {
		return q.QueryContext(ctx, RebindPostgres(query), args)
	}
	return nil, driver.ErrSkip
}

func (c *rebindConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if e, ok := c.Conn.(driver.ConnBeginTx); ok {
		return e.BeginTx(ctx, opts)
	}
	return c.Conn.Begin()
}

func (c *rebindConn) Ping(ctx context.Context) error {
	if p, ok := c.Conn.(driver.Pinger); ok {
		return p.Ping(ctx)
	}
	return nil
}
