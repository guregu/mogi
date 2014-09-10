package mogi

import (
	"database/sql/driver"
)

var drv *mdriver

type mdriver struct {
	*conn
}

func newDriver() *mdriver {
	return &mdriver{
		conn: newConn(),
	}
}

func (d *mdriver) Open(name string) (driver.Conn, error) {
	return drv.conn, nil
}
