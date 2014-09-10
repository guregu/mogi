package mogi

import (
	"database/sql"
	"database/sql/driver"
	"errors"
)

var (
	ErrUnstubbed  = errors.New("query not stubbed")
	ErrUnresolved = errors.New("query matched but no stub data")
)

var (
	verbose = false
)

func init() {
	drv = newDriver()
	sql.Register("mogi", drv)
}

// Reset removes all the stubs that have been set
func Reset() {
	drv.conn.stubs = nil
}

// Verbose turns on unstubbed logging when v is true
func Verbose(v bool) {
	verbose = v
}

// func Replace() {
// 	drv.conn = newConn()
// }

var _ driver.Stmt = &stmt{}
var _ driver.Conn = &conn{}
var _ driver.Driver = &mdriver{}
