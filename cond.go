package mogi

import (
	"reflect"
	// "database/sql"
	// "database/sql/driver"

	"github.com/youtube/vitess/go/vt/sqlparser"
)

type cond interface {
	matches(in input) bool
}

type condchain []cond

func (chain *condchain) matches(in input) bool {
	for _, c := range *chain {
		if !c.matches(in) {
			return false
		}
	}
	return true
}

type selectCond struct {
	cols []string
}

func (sc *selectCond) matches(in input) bool {
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
	table string
}

func (tc *tableCond) matches(in input) bool {
	switch x := in.statement.(type) {
	case *sqlparser.Select:
		for _, tex := range x.From {
			table_expr := tex.(*sqlparser.AliasedTableExpr)
			if tn, ok := table_expr.Expr.(*sqlparser.TableName); ok {
				if tc.table == string(tn.Name) {
					return true
				}
			}
		}
	}
	return false
}
