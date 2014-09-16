package mogi

import (
	"database/sql/driver"
	"log"

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

type arg int

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
		for _, expr := range x.Exprs {
			// TODO qualifiers
			name := string(expr.Name.Name)
			cols = append(cols, name)
		}
	}
	return cols
}

// for UPDATEs
func (in input) values() map[string]interface{} {
	vals := make(map[string]interface{})

	switch x := in.statement.(type) {
	case *sqlparser.Update:
		for _, expr := range x.Exprs {
			// TODO qualifiers
			colName := string(expr.Name.Name)
			v := valToInterface(expr.Expr)
			if a, ok := v.(arg); ok {
				// replace placeholders
				v = unify(in.args[int(a)])
			}
			vals[colName] = v
		}
	}

	return vals
}

// for INSERTs
func (in input) rows() []map[string]interface{} {
	var vals []map[string]interface{}
	cols := in.cols()

	switch x := in.statement.(type) {
	case *sqlparser.Insert:
		insertRows := x.Rows.(sqlparser.Values)
		vals = make([]map[string]interface{}, len(insertRows))
		for i, rowTuple := range insertRows {
			vals[i] = make(map[string]interface{})
			row := rowTuple.(sqlparser.ValTuple)
			for j, val := range row {
				colName := cols[j]
				v := valToInterface(val)
				if a, ok := v.(arg); ok {
					// replace placeholders
					v = unify(in.args[int(a)])
				}
				vals[i][colName] = v
			}
		}
	}
	return vals
}

func (in input) where() map[string]interface{} {
	if in.whereVars != nil {
		return in.whereVars
	}

	switch x := in.statement.(type) {
	case *sqlparser.Select:
		in.whereVars = extractBoolExpr(nil, x.Where.Expr)
		// replace placeholders
		for k, v := range in.whereVars {
			if a, ok := v.(arg); ok {
				in.whereVars[k] = unify(in.args[int(a)])
			}
		}
		return in.whereVars
	case *sqlparser.Update:
		in.whereVars = extractBoolExpr(nil, x.Where.Expr)
		// replace placeholders
		for k, v := range in.whereVars {
			if a, ok := v.(arg); ok {
				in.whereVars[k] = unify(in.args[int(a)])
			}
		}
		return in.whereVars
	}
	return nil
}
