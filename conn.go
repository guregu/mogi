package mogi

import (
	"log"
	"sort"

	"database/sql/driver"
)

type conn struct {
	stubs     stubs
	execStubs execStubs
}

func newConn() *conn {
	return &conn{}
}

func addStub(s *Stub) {
	drv.conn.stubs = append(drv.conn.stubs, s)
	sort.Sort(drv.conn.stubs)
}

func addExecStub(s *ExecStub) {
	drv.conn.execStubs = append(drv.conn.execStubs, s)
	sort.Sort(drv.conn.execStubs)
}

func (c *conn) Prepare(query string) (driver.Stmt, error) {
	return &stmt{
		query: query,
	}, nil
}

func (c *conn) Close() error {
	return nil
}

func (c *conn) Begin() (driver.Tx, error) {
	return &tx{}, nil
}

func (c *conn) Query(query string, args []driver.Value) (driver.Rows, error) {
	in, err := newInput(query, args)
	if err != nil {
		return nil, err
	}
	for _, c := range c.stubs {
		if c.matches(in) {
			return c.rows(in)
		}
	}
	if verbose {
		log.Println("Unstubbed query:", query, args)
	}
	return nil, ErrUnstubbed
}

func (c *conn) Exec(query string, args []driver.Value) (driver.Result, error) {
	in, err := newInput(query, args)
	if err != nil {
		return nil, err
	}
	for _, c := range c.execStubs {
		if c.matches(in) {
			return c.results()
		}
	}
	if verbose {
		log.Println("Unstubbed query:", query, args)
	}
	return nil, ErrUnstubbed
}
