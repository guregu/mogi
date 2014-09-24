package mogi

import (
	"database/sql/driver"
	"fmt"
	"log"
	"strconv"
	"strings"

	// "github.com/davecgh/go-spew/spew"
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

func unifyValues(arr []driver.Value) []driver.Value {
	for i, v := range arr {
		arr[i] = unify(v)
	}
	return arr
}

func unifyArray(arr []interface{}) []interface{} {
	for i, v := range arr {
		arr[i] = unify(v)
	}
	return arr
}

func lowercase(arr []string) []string {
	lower := make([]string, 0, len(arr))
	for _, str := range arr {
		lower = append(lower, strings.ToLower(str))
	}
	return lower
}

// transmogrify takes sqlparser expressions and turns them into useful go values
func transmogrify(v interface{}) interface{} {
	switch x := v.(type) {
	case *sqlparser.ColName:
		name := string(x.Name)
		if x.Qualifier != nil {
			name = fmt.Sprintf("%s.%s", x.Qualifier, name)
		}
		return name
	case *sqlparser.NonStarExpr:
		if x.As != nil {
			return string(x.As)
		}
		return transmogrify(x.Expr)
	case *sqlparser.StarExpr:
		return "*"
	case *sqlparser.FuncExpr:
		name := strings.ToUpper(string(x.Name))
		var args []string
		for _, expr := range x.Exprs {
			args = append(args, stringify(transmogrify(expr)))
		}
		return fmt.Sprintf("%s(%s)", name, strings.Join(args, ", "))
	case *sqlparser.BinaryExpr:
		// TODO: figure out some way to make this work
		return transmogrify(x.Left)
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
	case sqlparser.ValTuple:
		vals := make([]interface{}, 0, len(x))
		for _, item := range x {
			vals = append(vals, transmogrify(item))
		}
		return vals
	default:
		log.Printf("unknown transmogrify: (%T) %v\n", v, v)
		//panic(x)
	}
	return nil
}

func extractColumnName(nse *sqlparser.NonStarExpr) string {
	if nse.As != nil {
		return string(nse.As)
	}
	return stringify(transmogrify(nse.Expr))
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
	case *sqlparser.OrExpr:
		extractBoolExpr(vals, x.Left)
		extractBoolExpr(vals, x.Right)
	case *sqlparser.AndExpr:
		extractBoolExpr(vals, x.Left)
		extractBoolExpr(vals, x.Right)
	case *sqlparser.ComparisonExpr:
		column := transmogrify(x.Left).(string)
		vals[column] = transmogrify(x.Right)
	}
	return vals
}

func stringify(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case nil:
		return "NULL"
	default:
		fmt.Printf("stringify unknown type %T: %v\n", v, v)
	}
	return "???"
}
