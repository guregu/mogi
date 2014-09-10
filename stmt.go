package mogi

import (
	"database/sql/driver"
)

type stmt struct {
}

func (s *stmt) Close() error {
	return nil
}

// NumInput returns the number of placeholder parameters.
//
// If NumInput returns >= 0, the sql package will sanity check
// argument counts from callers and return errors to the caller
// before the statement's Exec or Query methods are called.
//
// NumInput may also return -1, if the driver doesn't know
// its number of placeholders. In that case, the sql package
// will not sanity check Exec or Query argument counts.
func (s *stmt) NumInput() int {
	return -1
}

// Exec executes a query that doesn't return rows, such
// as an INSERT or UPDATE.
func (s *stmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, nil
}

// Query executes a query that may return rows, such as a
// SELECT.
func (s *stmt) Query(args []driver.Value) (driver.Rows, error) {
	return nil, nil
}
