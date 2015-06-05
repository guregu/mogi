package mogi

import (
	"fmt"
	"strings"
)

type whereCond struct {
	col string
	v   []interface{}
}

func newWhereCond(col string, v []interface{}) whereCond {
	return whereCond{
		col: strings.ToLower(col),
		v:   unifyInterfaces(v),
	}
}

func (wc whereCond) matches(in input) bool {
	vals := in.where()
	v, ok := vals[wc.col]
	if !ok {
		return false
	}

	// compare slices
	if slice, ok := v.([]interface{}); ok {
		for i, src := range slice {
			if !equals(src, wc.v[i]) {
				return false
			}
		}
		return true
	}

	// compare single value
	return equals(v, wc.v[0])
}

func (wc whereCond) priority() int {
	return 1
}

func (wc whereCond) String() string {
	return fmt.Sprintf("WHERE %s ≈ %v", wc.col, wc.v)
}

type whereOpCond struct {
	col string
	op  string
	v   []interface{}
}

func newWhereOpCond(col string, v []interface{}, op string) whereOpCond {
	return whereOpCond{
		col: strings.ToLower(col),
		v:   unifyInterfaces(v),
		op:  strings.ToLower(op),
	}
}

func (wc whereOpCond) matches(in input) bool {
	vals := in.whereOp()
	v, ok := vals[colop{wc.col, wc.op}]
	if !ok {
		return false
	}

	// compare slices
	if slice, ok := v.([]interface{}); ok {
		for i, src := range slice {
			if !equals(src, wc.v[i]) {
				return false
			}
		}
		return true
	}

	// compare single value
	return equals(v, wc.v[0])
}

func (wc whereOpCond) priority() int {
	return 2
}

func (wc whereOpCond) String() string {
	return fmt.Sprintf("WHERE %s %s %v", wc.col, strings.ToUpper(wc.op), wc.v)
}
