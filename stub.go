package mogi

import (
	"database/sql/driver"
)

type Stub struct {
	chain condchain
	cols  []string
	data  [][]driver.Value
	err   error

	resolve func(input)
}

func Select(cols ...string) *Stub {
	return &Stub{
		chain: condchain{selectCond{
			cols: cols,
		}},
	}
}

func (s *Stub) From(table string) *Stub {
	s.chain = append(s.chain, tableCond{
		table: table,
	})
	return s
}

func (s *Stub) Where(col string, v interface{}) *Stub {
	s.chain = append(s.chain, newWhereCond(col, v))
	return s
}

func (s *Stub) StubCSV(data string) {
	s.resolve = func(in input) {
		s.data = csvToValues(in.cols(), data)
	}
	addStub(s)
}

func (s *Stub) Stub(rows [][]driver.Value) {
	s.data = rows
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

// stubs are arranged by how complex they are for now
type stubs []*Stub

func (s stubs) Len() int           { return len(s) }
func (s stubs) Less(i, j int) bool { return len(s[i].chain) < len(s[j].chain) }
func (s stubs) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
