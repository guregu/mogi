package mogi_test

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/guregu/mogi"
)

const (
	beerCSV = `1,Yona Yona Ale,Yo-Ho Brewing,5.5
	2,Punk IPA,BrewDog,5.6`
)

var (
	beers = map[int]beer{
		1: beer{
			id:      1,
			name:    "Yona Yona Ale",
			brewery: "Yo-Ho Brewing",
			pct:     5.5,
		},
		2: beer{
			id:      2,
			name:    "Punk IPA",
			brewery: "BrewDog",
			pct:     5.6,
		},
	}
)

type beer struct {
	id      int64
	name    string
	brewery string
	pct     float64
}

func TestMogi(t *testing.T) {
	defer mogi.Reset()
	mogi.Verbose(false)
	db := openDB()

	// select (any columns)
	mogi.Select().From("beer").StubCSV(beerCSV)
	runBeerSelectQuery(t, db)

	// test reset
	mogi.Reset()
	_, err := db.Query("SELECT id, name, brewery, pct FROM beer WHERE pct > ?", 5)
	if err != mogi.ErrUnstubbed {
		t.Error("after reset, err should be ErrUnstubbed but is", err)
	}

	// select specific columns
	mogi.Select("id", "name", "brewery", "pct").From("beer").StubCSV(beerCSV)
	runBeerSelectQuery(t, db)

	// select the "wrong" columns
	mogi.Reset()
	mogi.Select("hello", "👞").From("beer").StubCSV(beerCSV)
	runUnstubbedSelect(t, db)

	// select the wrong table
	mogi.Reset()
	mogi.Select("id", "name", "brewery", "pct").From("酒").StubCSV(beerCSV)
	runUnstubbedSelect(t, db)

	// where
	mogi.Reset()
	mogi.Select().Where("pct", 5).StubCSV(beerCSV)
	db.Query("SELECT id, name, brewery, pct FROM beer WHERE a = ? AND b = ? AND c IS NULL", 5)
	//runBeerSelectQuery(t, db)
}

func runUnstubbedSelect(t *testing.T, db *sql.DB) {
	_, err := db.Query("SELECT id, name, brewery, pct FROM beer WHERE pct > ?", 5)
	if err != mogi.ErrUnstubbed {
		t.Error("with unmatched query, err should be ErrUnstubbed but is", err)
	}
}

func runBeerSelectQuery(t *testing.T, db *sql.DB) {
	expectCols := []string{"id", "name", "brewery", "pct"}
	rows, err := db.Query("SELECT id, name, brewery, pct FROM beer WHERE pct > ?", 5)
	checkNil(t, err)
	cols, err := rows.Columns()
	checkNil(t, err)
	if !reflect.DeepEqual(cols, expectCols) {
		t.Error("bad columns", cols, "≠", expectCols)
	}
	i := 0
	for rows.Next() {
		var b beer
		rows.Scan(&b.id, &b.name, &b.brewery, &b.pct)
		checkBeer(t, b, i+1)
		i++
	}
}

func checkBeer(t *testing.T, b beer, id int) {
	cmp, ok := beers[id]
	if !ok {
		t.Error("unknown beer", id)
		return
	}
	if b != cmp {
		t.Error("beers don't match", b, "≠", cmp, id)
	}
}

func checkNil(t *testing.T, err error) {
	if err != nil {
		t.Error("error should be nil but is", err)
	}
}

func openDB() *sql.DB {
	db, _ := sql.Open("mogi", "")
	return db
}
