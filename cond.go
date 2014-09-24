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

type selectCond struct {
	cols []string
}

func (sc selectCond) matches(in input) bool {
	_, ok := in.statement.(*sqlparser.Select)
	if !ok {
		return false
	}

	// zero parameters means anything
	if len(sc.cols) == 0 {
		return true
	}
	return reflect.DeepEqual(lowercase(sc.cols), lowercase(in.cols()))
}

func (sc selectCond) priority() int {
	if len(sc.cols) > 0 {
		return 2
	}
	return 1
}

func (sc selectCond) String() string {
	cols := "(any)" // TODO support star select
	if len(sc.cols) > 0 {
		cols = strings.Join(sc.cols, ", ")
	}
	return fmt.Sprintf("SELECT %s", cols)
}

type fromCond struct {
	tables []string
}

func (fc fromCond) matches(in input) bool {
	var inTables []string
	switch x := in.statement.(type) {
	case *sqlparser.Select:
		for _, tex := range x.From {
			extractTableNames(&inTables, tex)
		}
	}
	return reflect.DeepEqual(lowercase(fc.tables), lowercase(inTables))
}

func (fc fromCond) priority() int {
	if len(fc.tables) > 0 {
		return 1
	}
	return 0
}

func (fc fromCond) String() string {
	return fmt.Sprintf("FROM %s", strings.Join(fc.tables, ", "))
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

type whereCond struct {
	col string
	v   []interface{}
}

func newWhereCond(col string, v []interface{}) whereCond {
	return whereCond{
		col: col,
		v:   unifyArray(v),
	}
}

func (wc whereCond) matches(in input) bool {
	vals := in.where()
	v, ok := vals[wc.col]
	if !ok {
		return false
	}
	// if we aren't comparing against an array, use the first value
	if _, isArray := v.([]interface{}); !isArray {
		return reflect.DeepEqual(wc.v[0], v)
	}
	return reflect.DeepEqual(wc.v, v)
}

func (wc whereCond) priority() int {
	return 1
}

func (wc whereCond) String() string {
	return fmt.Sprintf("WHERE %s ≈ %v", wc.col, wc.v)
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

type insertCond struct {
	cols []string
}

func (ic insertCond) matches(in input) bool {
	_, ok := in.statement.(*sqlparser.Insert)
	if !ok {
		return false
	}

	// zero parameters means anything
	if len(ic.cols) == 0 {
		return true
	}

	return reflect.DeepEqual(lowercase(ic.cols), lowercase(in.cols()))
}

func (ic insertCond) priority() int {
	if len(ic.cols) > 0 {
		return 2
	}
	return 1
}

func (ic insertCond) String() string {
	cols := "(any)" // TODO support star select
	if len(ic.cols) > 0 {
		cols = strings.Join(ic.cols, ", ")
	}
	return fmt.Sprintf("INSERT %s", cols)
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

type updateCond struct {
	cols []string
}

func (uc updateCond) matches(in input) bool {
	_, ok := in.statement.(*sqlparser.Update)
	if !ok {
		return false
	}

	// zero parameters means anything
	if len(uc.cols) == 0 {
		return true
	}

	return reflect.DeepEqual(lowercase(uc.cols), lowercase(in.cols()))
}

func (uc updateCond) priority() int {
	if len(uc.cols) > 0 {
		return 2
	}
	return 1
}

func (uc updateCond) String() string {
	cols := "(any)" // TODO support star select
	if len(uc.cols) > 0 {
		cols = strings.Join(uc.cols, ", ")
	}
	return fmt.Sprintf("UPDATE %s", cols)
}

type deleteCond struct{}

func (uc deleteCond) matches(in input) bool {
	_, ok := in.statement.(*sqlparser.Delete)
	return ok
}

func (uc deleteCond) priority() int {
	return 1
}

func (uc deleteCond) String() string {
	return "DELETE"
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
