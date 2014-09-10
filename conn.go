package mogi

import (
	"log"
	"sort"

	"database/sql/driver"
)

type conn struct {
	stubs
}

func newConn() *conn {
	return &conn{}
}

func addStub(s *stub) {
	drv.conn.stubs = append(drv.conn.stubs, s)
	sort.Sort(drv.conn.stubs)
	log.Printf("%#v", drv.conn.stubs)
}

func (c *conn) Prepare(query string) (driver.Stmt, error) {
	return &stmt{}, nil
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
			log.Println("match!", c)
			return c.rows(in)
		}
	}
	log.Println("got query", query, args)
	return &rows{}, nil
}

// TODO exec
