package mogi

import (
	"database/sql"
	"database/sql/driver"
)

func init() {
	drv = newDriver()
	sql.Register("mogi", drv)
}

var _ driver.Stmt = &stmt{}
var _ driver.Conn = &conn{}
var _ driver.Driver = &mdriver{}
