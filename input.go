package mogi

import (
	"database/sql/driver"
	"log"

	// "github.com/davecgh/go-spew/spew"
	"github.com/guregu/mogi/internal/sqlparser"
)

type input struct {
	query     string
	statement sqlparser.Statement
	args      []driver.Value

	whereVars   map[string]interface{}
	whereOpVars map[colop]interface{}
}

func newInput(query string, args []driver.Value) (in input, err error) {
	in = input{
		query: query,
		args:  args,
	}
	in.statement, err = sqlparser.Parse(query)
	return
}

type arg int

type opval struct {
	op string
	v  interface{}
}

type colop struct {
	col string
	op  string
}

/*
Column name rules:
SELECT a        → a
SELECT a.b      → a.b
SELECT a.b AS c → c
*/
func (in input) cols() []string {
	var cols []string

	switch x := in.statement.(type) {
	case *sqlparser.Select:
		for _, sexpr := range x.SelectExprs {
			name := stringify(transmogrify(sexpr))
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
			v := transmogrify(expr.Expr)
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
				v := transmogrify(val)
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

// for SELECT and UPDATE and DELETE
func (in input) where() map[string]interface{} {
	if in.whereVars != nil {
		return in.whereVars
	}
	var w *sqlparser.Where
	switch x := in.statement.(type) {
	case *sqlparser.Select:
		w = x.Where
	case *sqlparser.Update:
		w = x.Where
	case *sqlparser.Delete:
		w = x.Where
	default:
		return nil
	}
	if w == nil {
		return map[string]interface{}{}
	}
	in.whereVars = extractBoolExpr(nil, w.Expr)
	// replace placeholders
	for k, v := range in.whereVars {
		if a, ok := v.(arg); ok {
			in.whereVars[k] = unify(in.args[int(a)])
			continue
		}

		// arrays
		if arr, ok := v.([]interface{}); ok {
			for i, v := range arr {
				if a, ok := v.(arg); ok {
					arr[i] = unify(in.args[int(a)])
				}
			}
		}
	}
	return in.whereVars
}

// for SELECT and UPDATE and DELETE
func (in input) whereOp() map[colop]interface{} {
	if in.whereOpVars != nil {
		return in.whereOpVars
	}
	var w *sqlparser.Where
	switch x := in.statement.(type) {
	case *sqlparser.Select:
		w = x.Where
	case *sqlparser.Update:
		w = x.Where
	case *sqlparser.Delete:
		w = x.Where
	default:
		return nil
	}
	if w == nil {
		return map[colop]interface{}{}
	}
	in.whereOpVars = extractBoolExprWithOps(nil, w.Expr)
	// replace placeholders
	for k, v := range in.whereOpVars {
		if a, ok := v.(arg); ok {
			in.whereOpVars[k] = unify(in.args[int(a)])
			continue
		}

		// arrays
		if arr, ok := v.([]interface{}); ok {
			for i, v := range arr {
				if a, ok := v.(arg); ok {
					arr[i] = unify(in.args[int(a)])
				}
			}
		}
	}
	return in.whereOpVars
}
