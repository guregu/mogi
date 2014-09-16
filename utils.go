package mogi

import (
	"database/sql/driver"
	"fmt"
	"log"
	"strconv"

	"github.com/youtube/vitess/go/vt/sqlparser"
)

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

func unifyArray(arr []driver.Value) []driver.Value {
	for i, v := range arr {
		arr[i] = unify(v)
	}
	return arr
}

func valToInterface(v sqlparser.ValExpr) interface{} {
	switch x := v.(type) {
	case *sqlparser.ColName:
		return string(x.Name)
	case sqlparser.ValArg:
		// vitess makes args like :v1
		str := string(x)
		hdr, num := str[:2], str[2:]
		if hdr != ":v" {
			log.Panicln("unexpected arg format", str)
		}
		idx, err := strconv.Atoi(num)
		if err != nil {
			panic(err)
		}
		return arg(idx - 1)
	case sqlparser.StrVal:
		return string(x)
	case sqlparser.NumVal:
		s := string(x)
		n, err := strconv.ParseInt(s, 10, 64)
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

func extractBoolExpr(vals map[string]interface{}, expr sqlparser.BoolExpr) map[string]interface{} {
	if vals == nil {
		vals = make(map[string]interface{})
	}
	switch x := expr.(type) {
	case *sqlparser.AndExpr:
		extractBoolExpr(vals, x.Left)
		extractBoolExpr(vals, x.Right)
	case *sqlparser.ComparisonExpr:
		column := valToInterface(x.Left).(string)
		vals[column] = valToInterface(x.Right)
	}
	return vals
}
