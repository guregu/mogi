package mogi

import (
	"database/sql/driver"
	"fmt"
	"log"
	"strconv"

	// "github.com/davecgh/go-spew/spew"
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

/*
Column name rules:
SELECT a        → a
SELECT a.b      → a.b
SELECT a.b AS c → c
*/
func (in input) cols() []string {
	var cols []string
	in.argCursor = 0

	switch x := in.statement.(type) {
	case *sqlparser.Select:
		for _, sexpr := range x.SelectExprs {
			nse, ok := sexpr.(*sqlparser.NonStarExpr)
			if !ok {
				log.Println("something other than NonStarExpr", sexpr)
				continue
			}
			name := extractColumnName(nse)
			cols = append(cols, name)
		}
	case *sqlparser.Insert:
		for _, c := range x.Columns {
			nse, ok := c.(*sqlparser.NonStarExpr)
			if !ok {
				log.Println("something other than NonStarExpr", c)
				continue
			}
			name := extractColumnName(nse)
			cols = append(cols, name)
		}
	case *sqlparser.Update:
		// TODO
	}
	return cols
}

func (in input) values() []map[string]interface{} {
	cols := in.cols()
	argCursor := 0

	var vals []map[string]interface{}
	switch x := in.statement.(type) {
	case *sqlparser.Insert:
		insertRows := x.Rows.(sqlparser.Values)
		vals = make([]map[string]interface{}, len(insertRows))
		for i, rowTuple := range insertRows {
			vals[i] = make(map[string]interface{})
			row := rowTuple.(sqlparser.ValTuple)
			for j, val := range row {
				colName := cols[j]
				v := valToInterface(&argCursor, val)
				if a, ok := v.(arg); ok {
					// replace placeholders
					v = in.args[int(a)]
				}
				vals[i][colName] = v
			}
		}
	}
	return vals
}

func (in input) where() map[string]interface{} {
	argCursor := 0

	if in.whereVars != nil {
		return in.whereVars
	}
	switch x := in.statement.(type) {
	case *sqlparser.Select:
		in.whereVars = extractBoolExpr(nil, &argCursor, x.Where.Expr)

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

func extractBoolExpr(vals map[string]interface{}, cursor *int, expr sqlparser.BoolExpr) map[string]interface{} {
	if vals == nil {
		vals = make(map[string]interface{})
	}
	switch x := expr.(type) {
	case *sqlparser.AndExpr:
		extractBoolExpr(vals, cursor, x.Left)
		extractBoolExpr(vals, cursor, x.Right)
	case *sqlparser.ComparisonExpr:
		column := valToInterface(cursor, x.Left).(string)
		vals[column] = valToInterface(cursor, x.Right)
	}
	return vals
}

type arg int

func valToInterface(cursor *int, v sqlparser.ValExpr) interface{} {
	switch x := v.(type) {
	case *sqlparser.ColName:
		return string(x.Name)
	case sqlparser.ValArg:
		defer func() { *cursor = *cursor + 1 }()
		return arg(*cursor)
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

func extractColumnName(nse *sqlparser.NonStarExpr) string {
	colname, ok := nse.Expr.(*sqlparser.ColName)
	if !ok {
		log.Println("something other than ColName", nse.Expr)
		panic(colname)
	}
	name := string(colname.Name)
	switch {
	case nse.As != nil:
		name = string(nse.As)
	case colname.Qualifier != nil:
		name = fmt.Sprintf("%s.%s", colname.Qualifier, colname.Name)
	}
	return name
}
