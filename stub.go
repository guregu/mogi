package mogi

import (
	"database/sql/driver"
)

// Stub is a SQL query stub (for SELECT)
type Stub struct {
	chain condchain
	data  [][]driver.Value
	err   error

	resolve func(input)
}

type subquery struct {
	chain condchain
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

// Where further filters this stub by values of input in the WHERE clause.
// You can pass multiple values for IN clause matching.
func (s *Stub) Where(col string, v ...interface{}) *Stub {
	s.chain = append(s.chain, newWhereCond(col, v))
	return s
}

// WhereOp further filters this stub by values of input and the operator used in the WHERE clause.
func (s *Stub) WhereOp(col string, operator string, v ...interface{}) *Stub {
	s.chain = append(s.chain, newWhereOpCond(col, v, operator))
	return s
}

// Args further filters this stub, matching based on the args passed to the query
func (s *Stub) Args(args ...driver.Value) *Stub {
	s.chain = append(s.chain, argsCond{args})
	return s
}

// Priority adds the given priority to this stub, without performing any matching.
func (s *Stub) Priority(p int) *Stub {
	s.chain = append(s.chain, priorityCond{p})
	return s
}

// Notify will have this stub send to the given channel when matched.
// You should put this as the last part of your stub chain.
func (s *Stub) Notify(ch chan<- struct{}) *Stub {
	s.chain = append(s.chain, notifyCond{ch})
	return s
}

// Dump outputs debug information, without performing any matching.
func (s *Stub) Dump() *Stub {
	s.chain = append(s.chain, dumpCond{})
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

func (s *Stub) Subquery() subquery {
	return subquery{chain: s.chain}
}

func (s *Stub) matches(in input) bool {
	return s.chain.matches(in)
}

func (s *Stub) rows(in input) (*rows, error) {
	switch {
	case s.err != nil:
		return nil, s.err
	case s.data == nil && s.resolve != nil:
		s.resolve(in)
	}
	return newRows(in.cols(), s.data), nil
}

func (s *Stub) priority() int {
	return s.chain.priority()
}

// stubs are arranged by how complex they are for now
type stubs []*Stub

func (s stubs) Len() int           { return len(s) }
func (s stubs) Less(i, j int) bool { return s[i].priority() > s[j].priority() }
func (s stubs) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
