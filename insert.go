package mogi

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/guregu/mogi/internal/sqlparser"
)

type insertCond struct {
	cols []string
}

func (ic insertCond) matches(in input) bool {
	_, ok := in.statement.(*sqlparser.Insert)
	if !ok {
		return false
	}

	// zero parameters means anything
	if len(ic.cols) == 0 {
		return true
	}

	return reflect.DeepEqual(lowercase(ic.cols), lowercase(in.cols()))
}

func (ic insertCond) priority() int {
	if len(ic.cols) > 0 {
		return 2
	}
	return 1
}

func (ic insertCond) String() string {
	cols := "(any)" // TODO support star select
	if len(ic.cols) > 0 {
		cols = strings.Join(ic.cols, ", ")
	}
	return fmt.Sprintf("INSERT %s", cols)
}
