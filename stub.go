package mogi

import (
	"database/sql/driver"
)

type stub interface {
	matches(in input) bool
	rows(in input) (*rows, error)
	result(in input) (driver.Result, error)
	priority() int
}

// Stub is a SQL query stub (for SELECT)
type Stub struct {
	chain condchain
	data  [][]driver.Value
	err   error

	resolve func(input)
}

// Select starts a new stub for SELECT statements.
// You can filter out which columns to use this stub for.
// If you don't pass any columns, it will stub all SELECT queries.
func Select(cols ...string) *Stub {
	return &Stub{
		chain: condchain{selectCond{
			cols: cols,
		}},
	}
}

// From further filters this stub by table names in the FROM and JOIN clauses (in order).
// You need to give it the un-aliased table names.
func (s *Stub) From(tables ...string) *Stub {
	s.chain = append(s.chain, fromCond{
		tables: tables,
	})
	return s
}

// Where further filters this stub by values of input in the WHERE clause
func (s *Stub) Where(col string, v interface{}) *Stub {
	s.chain = append(s.chain, newWhereCond(col, v))
	return s
}

// Args further filters this stub, matching based on the args passed to the query
func (s *Stub) Args(args ...driver.Value) *Stub {
	s.chain = append(s.chain, argsCond{args})
	return s
}

// StubCSV takes CSV data and registers this stub with the driver
func (s *Stub) StubCSV(data string) {
	s.resolve = func(in input) {
		s.data = csvToValues(in.cols(), data)
	}
	addStub(s)
}

// Stub takes row data and registers this stub with the driver
func (s *Stub) Stub(rows [][]driver.Value) {
	s.data = rows
	addStub(s)
}

// StubError registers this stub to return the given error
func (s *Stub) StubError(err error) {
	s.err = err
	addStub(s)
}

func (s *Stub) matches(in input) bool {
	return s.chain.matches(in)
}

func (s *Stub) rows(in input) (*rows, error) {
	switch {
	case s.err != nil:
		return nil, s.err
	case s.data == nil && s.err == nil:
		// try to resolve the values
		if s.resolve == nil {
			return nil, ErrUnresolved
		}
		s.resolve(in)
	}
	return newRows(in.cols(), s.data), nil
}

func (s *Stub) priority() int {
	return len(s.chain)
}

// ExecStub is a SQL exec stub (for INSERT, UPDATE, DELETE)
type ExecStub struct {
	chain  condchain
	result driver.Result
	err    error
}

// Select starts a new stub for INSERT statements.
// You can filter out which columns to use this stub for.
// If you don't pass any columns, it will stub all INSERT queries.
func Insert(cols ...string) *ExecStub {
	return &ExecStub{
		chain: condchain{insertCond{
			cols: cols,
		}},
	}
}

// Into further filters this stub, matching the target table in INSERT or UPDATEs.
func (s *ExecStub) Table(table string) *ExecStub {
	s.chain = append(s.chain, tableCond{
		table: table,
	})
	return s
}

// Into further filters this stub, matching based on the INTO table specified.
func (s *ExecStub) Into(table string) *ExecStub {
	return s.Table(table)
}

// Value further filters this stub, matching based on values supplied to the query
// It matches the first row of values, so it is a shortcut for ValueN(0, ...)
func (s *ExecStub) Value(col string, v interface{}) *ExecStub {
	s.ValueN(0, col, v)
	return s
}

// ValueN further filters this stub, matching based on values supplied to the query
func (s *ExecStub) ValueN(row int, col string, v interface{}) *ExecStub {
	s.chain = append(s.chain, newValueCond(row, col, v))
	return s
}

// Args further filters this stub, matching based on the args passed to the query
func (s *ExecStub) Args(args ...driver.Value) *ExecStub {
	s.chain = append(s.chain, argsCond{args})
	return s
}

// Stub takes a driver.Result and registers this stub with the driver
func (s *ExecStub) Stub(res driver.Result) {
	s.result = res
	addExecStub(s)
}

// StubResult is an easy way to stub a driver.Result.
// Given a value of -1, the result will return an error for that particular part.
func (s *ExecStub) StubResult(lastInsertID, rowsAffected int64) {
	s.result = execResult{
		lastInsertID: lastInsertID,
		rowsAffected: rowsAffected,
	}
	addExecStub(s)
}

// Stub takes an error and registers this stub with the driver
func (s *ExecStub) StubError(err error) {
	s.err = err
	addExecStub(s)
}

func (s *ExecStub) matches(in input) bool {
	return s.chain.matches(in)
}

func (s *ExecStub) results() (driver.Result, error) {
	return s.result, s.err
}

func (s *ExecStub) priority() int {
	return len(s.chain)
}

// stubs are arranged by how complex they are for now
type stubs []*Stub

func (s stubs) Len() int           { return len(s) }
func (s stubs) Less(i, j int) bool { return len(s[i].chain) < len(s[j].chain) }
func (s stubs) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type execStubs []*ExecStub

func (s execStubs) Len() int           { return len(s) }
func (s execStubs) Less(i, j int) bool { return len(s[i].chain) < len(s[j].chain) }
func (s execStubs) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
