package mogi

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/youtube/vitess/go/vt/sqlparser"
)

type updateCond struct {
	cols []string
}

func (uc updateCond) matches(in input) bool {
	_, ok := in.statement.(*sqlparser.Update)
	if !ok {
		return false
	}

	// zero parameters means anything
	if len(uc.cols) == 0 {
		return true
	}

	return reflect.DeepEqual(lowercase(uc.cols), lowercase(in.cols()))
}

func (uc updateCond) priority() int {
	if len(uc.cols) > 0 {
		return 2
	}
	return 1
}

func (uc updateCond) String() string {
	cols := "(any)" // TODO support star select
	if len(uc.cols) > 0 {
		cols = strings.Join(uc.cols, ", ")
	}
	return fmt.Sprintf("UPDATE %s", cols)
}
