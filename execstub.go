package mogi

import (
	"database/sql/driver"
)

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

// Select starts a new stub for UPDATE statements.
// You can filter out which columns (from the SET statement) this stub is for.
// If you don't pass any columns, it will stub all UPDATE queries.
func Update(cols ...string) *ExecStub {
	return &ExecStub{
		chain: condchain{updateCond{
			cols: cols,
		}},
	}
}

// Delete starts a new stub for DELETE statements.
func Delete() *ExecStub {
	return &ExecStub{
		chain: condchain{deleteCond{}},
	}
}

// Into further filters this stub, matching the target table in INSERT, UPDATE, or DELETE.
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

// From further filters this stub, matching based on the FROM table specified.
func (s *ExecStub) From(table string) *ExecStub {
	return s.Table(table)
}

// Value further filters this stub, matching based on values supplied to the query
// For INSERTs, it matches the first row of values, so it is a shortcut for ValueAt(0, ...)
// For UPDATEs, it matches on the SET clause.
func (s *ExecStub) Value(col string, v interface{}) *ExecStub {
	s.ValueAt(0, col, v)
	return s
}

// ValueAt further filters this stub, matching based on values supplied to the query
func (s *ExecStub) ValueAt(row int, col string, v interface{}) *ExecStub {
	s.chain = append(s.chain, newValueCond(row, col, v))
	return s
}

// Where further filters this stub by values of input in the WHERE clause.
// You can pass multiple values for IN clause matching.
func (s *ExecStub) Where(col string, v ...interface{}) *ExecStub {
	s.chain = append(s.chain, newWhereCond(col, v))
	return s
}

// Args further filters this stub, matching based on the args passed to the query
func (s *ExecStub) Args(args ...driver.Value) *ExecStub {
	s.chain = append(s.chain, argsCond{args})
	return s
}

// Priority adds the given priority to this stub, without performing any matching.
func (s *ExecStub) Priority(p int) *ExecStub {
	s.chain = append(s.chain, priorityCond{p})
	return s
}

// Notify will have this stub send to the given channel when matched.
// You should put this as the last part of your stub chain.
func (s *ExecStub) Notify(ch chan<- struct{}) *ExecStub {
	s.chain = append(s.chain, notifyCond{ch})
	return s
}

// Dump outputs debug information, without performing any matching.
func (s *ExecStub) Dump() *ExecStub {
	s.chain = append(s.chain, dumpCond{})
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

// StubRowsAffected is an easy way to stub a driver.Result when you only need to specify the rows affected.
func (s *ExecStub) StubRowsAffected(rowsAffected int64) {
	s.StubResult(-1, rowsAffected)
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
	return s.chain.priority()
}

type execStubs []*ExecStub

func (s execStubs) Len() int           { return len(s) }
func (s execStubs) Less(i, j int) bool { return s[i].priority() > s[j].priority() }
func (s execStubs) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
