package mogi

import (
	"reflect"
	"strings"
	// "database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/youtube/vitess/go/vt/sqlparser"
)

type cond interface {
	matches(in input) bool
	priority() int
	fmt.Stringer
}

type condchain []cond

func (chain condchain) matches(in input) bool {
	for _, c := range chain {
		if !c.matches(in) {
			return false
		}
	}
	return true
}

func (chain condchain) priority() int {
	p := 0
	for _, c := range chain {
		p += c.priority()
	}
	return p
}

func (chain condchain) String() string {
	return "Chain..."
}

type tableCond struct {
	table string
}

func (tc tableCond) matches(in input) bool {
	switch x := in.statement.(type) {
	case *sqlparser.Insert:
		return strings.ToLower(tc.table) == strings.ToLower(string(x.Table.Name))
	case *sqlparser.Update:
		return strings.ToLower(tc.table) == strings.ToLower(string(x.Table.Name))
	case *sqlparser.Delete:
		return strings.ToLower(tc.table) == strings.ToLower(string(x.Table.Name))
	}
	return false
}

func (tc tableCond) priority() int {
	return 1
}

func (tc tableCond) String() string {
	return fmt.Sprintf("TABLE %s", tc.table)
}

type argsCond struct {
	args []driver.Value
}

func (ac argsCond) matches(in input) bool {
	given := unifyValues(ac.args)
	return reflect.DeepEqual(given, in.args)
}

func (ac argsCond) priority() int {
	return 1
}

func (ac argsCond) String() string {
	return fmt.Sprintf("WITH ARGS %+v", ac.args)
}

type valueCond struct {
	row int
	col string
	v   interface{}
}

func newValueCond(row int, col string, v interface{}) valueCond {
	return valueCond{
		row: row,
		col: col,
		v:   unify(v),
	}
}

func (vc valueCond) matches(in input) bool {
	switch in.statement.(type) {
	case *sqlparser.Insert:
		values := in.rows()
		if vc.row > len(values)-1 {
			return false
		}
		v, ok := values[vc.row][vc.col]
		if !ok {
			return false
		}
		return reflect.DeepEqual(vc.v, v)
	case *sqlparser.Update:
		values := in.values()
		v, ok := values[vc.col]
		if !ok {
			return false
		}
		return reflect.DeepEqual(vc.v, v)
	}
	return false
}

func (vc valueCond) priority() int {
	return 1
}

func (vc valueCond) String() string {
	return fmt.Sprintf("VALUE %s ≈ %v (row %d)", vc.col, vc.v, vc.row)
}

type priorityCond struct {
	p int
}

func (pc priorityCond) matches(in input) bool {
	return true
}

func (pc priorityCond) priority() int {
	return pc.p
}

func (pc priorityCond) String() string {
	return "PRIORITY"
}

type notifyCond struct {
	ch chan<- struct{}
}

func (nc notifyCond) matches(in input) bool {
	go func() {
		nc.ch <- struct{}{}
	}()
	return true
}

func (nc notifyCond) priority() int {
	return 0
}

func (nc notifyCond) String() string {
	return "NOTIFY"
}

type dumpCond struct{}

func (dc dumpCond) matches(in input) bool {
	fmt.Println(in.query)
	spew.Dump(in.args)
	switch in.statement.(type) {
	case *sqlparser.Select:
		spew.Dump(in.cols(), in.where())
	case *sqlparser.Insert:
		spew.Dump(in.cols(), in.rows())
	case *sqlparser.Update:
		spew.Dump(in.cols(), in.values(), in.where())
	}
	spew.Dump(in.statement)
	return true
}

func (dc dumpCond) priority() int {
	return 0
}

func (dc dumpCond) String() string {
	return "DUMP"
}
