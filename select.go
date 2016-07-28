package mogi

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/guregu/mogi/internal/sqlparser"
)

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

	return reflect.DeepEqual(lowercase(sc.cols), lowercase(in.cols()))
}

func (sc selectCond) priority() int {
	if len(sc.cols) > 0 {
		return 2
	}
	return 1
}

func (sc selectCond) String() string {
	cols := "(any)" // TODO support star select
	if len(sc.cols) > 0 {
		cols = strings.Join(sc.cols, ", ")
	}
	return fmt.Sprintf("SELECT %s", cols)
}

type fromCond struct {
	tables []string
}

func (fc fromCond) matches(in input) bool {
	var inTables []string
	switch x := in.statement.(type) {
	case *sqlparser.Select:
		for _, tex := range x.From {
			extractTableNames(&inTables, tex)
		}
	}
	return reflect.DeepEqual(lowercase(fc.tables), lowercase(inTables))
}

func (fc fromCond) priority() int {
	if len(fc.tables) > 0 {
		return 1
	}
	return 0
}

func (fc fromCond) String() string {
	return fmt.Sprintf("FROM %s", strings.Join(fc.tables, ", "))
}
