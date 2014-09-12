package mogi

import (
	"reflect"
	// "database/sql"
	"database/sql/driver"

	// "github.com/davecgh/go-spew/spew"
	"github.com/youtube/vitess/go/vt/sqlparser"
)

type cond interface {
	matches(in input) bool
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

	return reflect.DeepEqual(sc.cols, in.cols())
}

type tableCond struct {
	tables []string
}

func (tc tableCond) matches(in input) bool {
	var inTables []string
	switch x := in.statement.(type) {
	case *sqlparser.Select:
		for _, tex := range x.From {
			extractTableNames(&inTables, tex)
		}
	}
	return reflect.DeepEqual(tc.tables, inTables)
}

func extractTableNames(tables *[]string, from sqlparser.TableExpr) {
	switch x := from.(type) {
	case *sqlparser.AliasedTableExpr:
		if name, ok := x.Expr.(*sqlparser.TableName); ok {
			*tables = append(*tables, string(name.Name))
		}
	case *sqlparser.JoinTableExpr:
		extractTableNames(tables, x.LeftExpr)
		extractTableNames(tables, x.RightExpr)
	}
}

type whereCond struct {
	col string
	v   interface{}
}

func newWhereCond(col string, v interface{}) whereCond {
	return whereCond{
		col: col,
		v:   unify(v),
	}
}

func (wc whereCond) matches(in input) bool {
	vals := in.where()
	return reflect.DeepEqual(vals[wc.col], wc.v)
}

type argsCond struct {
	args []driver.Value
}

func (ac argsCond) matches(in input) bool {
	given := unifyArray(ac.args)
	return reflect.DeepEqual(given, in.args)
}

func unifyArray(arr []driver.Value) []driver.Value {
	for i, v := range arr {
		arr[i] = unify(v)
	}
	return arr
}

// convert args to their 64-bit versions
// for easy comparisons
func unify(v interface{}) interface{} {
	switch x := v.(type) {
	case int:
		return int64(x)
	case int32:
		return int64(x)
	case float32:
		return float64(x)
	}
	return v
}
