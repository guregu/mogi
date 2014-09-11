package mogi

import (
	"database/sql/driver"
	"strconv"

	"github.com/youtube/vitess/go/vt/sqlparser"
)

type input struct {
	query     string
	statement sqlparser.Statement
	args      []driver.Value

	whereVars map[string]interface{}
	argCursor int
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

func (in input) where() map[string]interface{} {
	if in.whereVars != nil {
		return in.whereVars
	}
	switch x := in.statement.(type) {
	case *sqlparser.Select:
		in.whereVars = in.extract(nil, x.Where.Expr)

		// replace placeholders
		for k, v := range in.whereVars {
			if a, ok := v.(arg); ok {
				in.whereVars[k] = in.args[int(a)]
			}
		}
		return in.whereVars
	}
	return nil
}

func (in input) extract(vals map[string]interface{}, expr sqlparser.BoolExpr) map[string]interface{} {
	if vals == nil {
		vals = make(map[string]interface{})
	}
	switch x := expr.(type) {
	case *sqlparser.AndExpr:
		in.extract(vals, x.Left)
		in.extract(vals, x.Right)
	case *sqlparser.ComparisonExpr:
		column := in.valToInterface(x.Left).(string)
		vals[column] = in.valToInterface(x.Right)
	}
	return vals
}

type arg int

func (in input) valToInterface(v sqlparser.ValExpr) interface{} {
	switch x := v.(type) {
	case *sqlparser.ColName:
		return string(x.Name)
	case sqlparser.ValArg:
		defer func() { in.argCursor++ }()
		return arg(in.argCursor)
	case sqlparser.StrVal:
		return string(x)
	case sqlparser.NumVal:
		s := string(x)
		n, err := strconv.Atoi(s)
		if err == nil {
			return n
		}
		f, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return f
		}
	default:
		//panic(x)
	}
	return nil
}
