//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-30

package resp3

import (
	"bufio"
	"net"
	"time"
)

func NewConn(conn net.Conn, timeout time.Duration) *Conn {
	return &Conn{
		conn:    conn,
		timeout: timeout,
		reader:  bufio.NewReader(conn),
	}
}

type Conn struct {
	conn    net.Conn
	reader  *bufio.Reader
	timeout time.Duration
}

func (c *Conn) Send(req Request) (Result, error) {
	b := bp.Get()
	bf := req.Bytes(b)
	c.conn.SetDeadline(time.Now().Add(c.timeout))
	_, err := c.conn.Write(bf)
	c.conn.SetDeadline(time.Time{})
	if err != nil {
		return nil, err
	}
	return ReadByType(c.reader, req.ResponseType())
}

func (c *Conn) Close() error {
	return c.conn.Close()
}
