package mogi

import (
	"database/sql"
	"database/sql/driver"
)

type stub struct {
	chain condchain
	cols  []string
	data  [][]driver.Value
	err   error

	resolve func(input)
}

func Select() *stub {
	return &stub{
		chain: condchain{&selectCond{}},
	}
}

func (s *stub) From(table string) *stub {
	s.chain = append(s.chain, &tableCond{
		table: table,
	})
	return s
}

func (s *stub) StubCSV(data string) {
	s.resolve = func(in input) {
		s.data = csvToValues(in.cols(), data)
	}
	addStub(s)
}

func (s *stub) matches(in input) bool {
	return s.chain.matches(in)
}

func (s *stub) rows(in input) (*rows, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.rows == nil {
		return nil, sql.ErrNoRows
	}
	if s.resolve == nil {
		// TODO better error message
		panic("values not set!")
	}
	s.resolve(in)
	return newRows(in.cols(), s.data), nil
}

// stubs are arranged by how complex they are for now
type stubs []*stub

func (s stubs) Len() int           { return len(s) }
func (s stubs) Less(i, j int) bool { return len(s[i].chain) < len(s[j].chain) }
func (s stubs) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
