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
	beer1 = beer{
		id:      1,
		name:    "Yona Yona Ale",
		brewery: "Yo-Ho Brewing",
		pct:     5.5,
	}
	beer2 = beer{
		id:      2,
		name:    "Punk IPA",
		brewery: "BrewDog",
		pct:     5.6,
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

	db := openDB()
	mogi.Select().From("beer").StubCSV(beerCSV)
	expectCols := []string{"id", "name", "brewery", "pct"}
	rows, err := db.Query("SELECT id, name, brewery, pct FROM beer WHERE pct > 5", 7)
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
	var cmp beer
	switch id {
	case 1:
		cmp = beer1
	case 2:
		cmp = beer2
	default:
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
