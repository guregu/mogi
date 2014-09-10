package mogi

import (
	"database/sql/driver"
	"encoding/csv"
	"io"
	"strings"
)

type rows struct {
	cols []string
	data [][]driver.Value

	cursor int
	closed bool
}

func newRows(cols []string, data [][]driver.Value) *rows {
	return &rows{
		cols: cols,
		data: data,
	}
}

func (r *rows) Columns() []string {
	return r.cols
}

// Close closes the rows iterator.
func (r *rows) Close() error {
	r.closed = true
	return nil
}

func (r *rows) Err() error {
	return nil
}

func (r *rows) Next(dest []driver.Value) error {
	r.cursor++
	if r.cursor > len(r.data) {
		r.closed = true
		return io.EOF
	}

	for i, col := range r.data[r.cursor-1] {
		dest[i] = col
	}

	return nil
}

// cribbed from DATA-DOG/go-sqlmock
// TODO rewrite
// TODO remove trimspace
func csvToValues(cols []string, s string) [][]driver.Value {
	var data [][]driver.Value

	res := strings.NewReader(strings.TrimSpace(s))
	csvReader := csv.NewReader(res)

	for {
		res, err := csvReader.Read()
		if err != nil || res == nil {
			break
		}

		row := make([]driver.Value, len(cols))
		for i, v := range res {
			row[i] = []byte(strings.TrimSpace(v))
		}
		data = append(data, row)
	}
	return data
}
