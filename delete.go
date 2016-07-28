package mogi

import (
	"github.com/guregu/mogi/internal/sqlparser"
)

type deleteCond struct{}

func (uc deleteCond) matches(in input) bool {
	_, ok := in.statement.(*sqlparser.Delete)
	return ok
}

func (uc deleteCond) priority() int {
	return 1
}

func (uc deleteCond) String() string {
	return "DELETE"
}
