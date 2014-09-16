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

type execResult struct {
	lastInsertID int64
	rowsAffected int64
}

func (r execResult) LastInsertId() (int64, error) {
	if r.lastInsertID == -1 {
		return 0, errNotSet
	}
	return r.lastInsertID, nil
}

// RowsAffected returns the number of rows affected by the
// query.
func (r execResult) RowsAffected() (int64, error) {
	if r.rowsAffected == -1 {
		return 0, errNotSet
	}
	return r.rowsAffected, nil
}
