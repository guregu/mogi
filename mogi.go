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
	verbose    = false
	timeLayout = ""
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

// ParseTime will configure mogi to convert dates of the given layout
// (e.g. time.RFC3339) to time.Time when using StubCSV.
// Give it an empty string to turn off time parsing.
func ParseTime(layout string) {
	timeLayout = layout
}

// Dump prints all the current stubs, in order of priority.
// Helpful for debugging.
func Dump() {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintf(w, ">>\t\tQuery stubs: (%d total)\t\n", len(drv.conn.stubs))
	fmt.Fprintf(w, "\t\t=========================\t\n")
	for rank, s := range drv.conn.stubs {
		for i, c := range s.chain {
			if i == 0 {
				fmt.Fprintf(w, "#%d\t[%d]\t%s\t[%+d]\n", rank+1, s.priority(), c, c.priority())
				continue
			}
			fmt.Fprintf(w, "\t\t%s\t[%+d]\n", c, c.priority())
		}
		switch {
		case s.err != nil:
			fmt.Fprintf(w, "\t\t→ error: %v\t\n", s.err)
		case s.data != nil, s.resolve != nil:
			fmt.Fprintln(w, "\t\t→ data\t\n")
		}
	}
	fmt.Fprintf(w, "\t\t\t\n")
	fmt.Fprintf(w, ">>\t\tExec stubs: (%d total)\t\n", len(drv.conn.execStubs))
	fmt.Fprintf(w, "\t\t=========================\t\n")
	for rank, s := range drv.conn.execStubs {
		for i, c := range s.chain {
			if i == 0 {
				fmt.Fprintf(w, "#%d\t[%d]\t%s\t[%+d]\n", rank+1, s.priority(), c, c.priority())
				continue
			}
			fmt.Fprintf(w, "\t\t%s\t[%+d]\n", c, c.priority())
		}
		switch {
		case s.err != nil:
			fmt.Fprintf(w, "\t\t→ error: %v\t\n", s.err)
		case s.result != nil:
			if r, ok := s.result.(execResult); ok {
				fmt.Fprintf(w, "\t\t→ result ID: %d, rows: %d\t\n", r.lastInsertID, r.rowsAffected)
			} else {
				fmt.Fprintf(w, "\t\t→ result %T\t\n", s.result)
			}
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
