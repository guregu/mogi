package mogi

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"os"
	"text/tabwriter"
)

var (
	ErrUnstubbed  = errors.New("query not stubbed")
	ErrUnresolved = errors.New("query matched but no stub data")

	errNotSet = errors.New("value set to -1")
)

var (
	verbose = false
)

func init() {
	drv = newDriver()
	sql.Register("mogi", drv)
}

// Reset removes all the stubs that have been set
func Reset() {
	drv.conn.stubs = nil
	drv.conn.execStubs = nil
}

// Verbose turns on unstubbed logging when v is true
func Verbose(v bool) {
	verbose = v
}

// Dump prints all the current stubs, in order of priority.
// Helpful for debugging.
func Dump() {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintf(w, "Query stubs: (%d total)\n", len(drv.conn.stubs))
	for rank, s := range drv.conn.stubs {
		for i, c := range s.chain {
			if i == 0 {
				fmt.Fprintf(w, "#%d\t[%d]\t%s\t[%+d]\n", rank+1, s.priority(), c, c.priority())
				continue
			}
			fmt.Fprintf(w, "\t\t%s\t[%+d]\n", c, c.priority())
		}
	}
	fmt.Fprintf(w, "Exec stubs: (%d total)\n", len(drv.conn.execStubs))
	for rank, s := range drv.conn.execStubs {
		for i, c := range s.chain {
			if i == 0 {
				fmt.Fprintf(w, "#%d\t[%d]\t%s\t[%+d]\n", rank+1, s.priority(), c, c.priority())
				continue
			}
			fmt.Fprintf(w, "\t\t%s\t[%+d]\n", c, c.priority())
		}
	}
	w.Flush()
}

// func Replace() {
// 	drv.conn = newConn()
// }

var _ driver.Stmt = &stmt{}
var _ driver.Conn = &conn{}
var _ driver.Driver = &mdriver{}
