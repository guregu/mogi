package mogi_test

import (
	"database/sql"
	"testing"

	"github.com/guregu/mogi"
)

func TestMogi(t *testing.T) {
	db := openDB()
	mogi.Select().From("beer").StubCSV("Asahi")
	rows, err := db.Query("SELECT brewery FROM beer WHERE pct = ?", 7)
	for rows.Next() {
		var b string
		rows.Scan(&b)
		t.Log(b)
	}
	t.Logf("%#v\n", err)
	t.Fail()
}

func openDB() *sql.DB {
	db, _ := sql.Open("mogi", "")
	return db
}
