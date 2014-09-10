package mogi

import (
	"database/sql/driver"

	"github.com/youtube/vitess/go/vt/sqlparser"
)

type input struct {
	query     string
	statement sqlparser.Statement
	args      []driver.Value
}

func newInput(query string, args []driver.Value) (in input, err error) {
	in = input{
		query: query,
		args:  args,
	}
	in.statement, err = sqlparser.Parse(query)
	return
}

func (in input) cols() []string {
	var cols []string
	switch x := in.statement.(type) {
	case *sqlparser.Select:
		for _, sexpr := range x.SelectExprs {
			nse, ok := sexpr.(*sqlparser.NonStarExpr)
			if !ok {
				continue
			}
			colname, ok := nse.Expr.(*sqlparser.ColName)
			if !ok {
				panic(colname)
			}
			cols = append(cols, string(colname.Name))
		}
	}
	return cols
}
