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

func TestSelectTable(t *testing.T) {
	defer mogi.Reset()
	db := openDB()

	// filter by table
	mogi.Select("id", "name", "brewery", "pct").From("beer").StubCSV(beerCSV)
	runBeerSelectQuery(t, db)

	// select the wrong table
	mogi.Reset()
	mogi.Select("id", "name", "brewery", "pct").From("酒").StubCSV(beerCSV)
	runUnstubbedSelect(t, db)
}

func TestSelectWhere(t *testing.T) {
	defer mogi.Reset()
	db := openDB()

	// where
	mogi.Select().From("beer").Where("pct", 5).StubCSV(beerCSV)
	runBeerSelectQuery(t, db)

	// where with weird type
	type 数字 int
	五 := 数字(5)
	mogi.Reset()
	mogi.Select().From("beer").Where("pct", &五).StubCSV(beerCSV)
	runBeerSelectQuery(t, db)

	// wrong where
	mogi.Reset()
	mogi.Select().From("beer").Where("pct", 98).StubCSV(beerCSV)
	runUnstubbedSelect(t, db)
}

func TestSelectArgs(t *testing.T) {
	defer mogi.Reset()
	db := openDB()

	// where
	mogi.Select().Args(5).StubCSV(beerCSV)
	runBeerSelectQuery(t, db)

	// wrong args
	mogi.Reset()
	mogi.Select().Args("サービス残業").StubCSV(beerCSV)
	runUnstubbedSelect(t, db)
}

func TestStubError(t *testing.T) {
	defer mogi.Reset()
	db := openDB()

	mogi.Select().StubError(sql.ErrNoRows)
	_, err := db.Query("SELECT id, name, brewery, pct FROM beer WHERE pct > ?", 5)
	if err != sql.ErrNoRows {
		t.Error("after StubError, err should be ErrNoRows but is", err)
	}
}

func TestSelectMultipleTables(t *testing.T) {
	defer mogi.Reset()
	db := openDB()

	mogi.Select().From("a", "b").StubCSV(`foo,bar`)
	_, err := db.Query("SELECT a.thing, b.thing FROM a, b WHERE a.id = b.id")
	checkNil(t, err)
	_, err = db.Query("SELECT a.thing, b.thing FROM a JOIN b ON a.id = b.id")
	checkNil(t, err)

	mogi.Reset()
	mogi.Select().From("a", "b", "c").StubCSV(`foo,bar,baz`)
	_, err = db.Query("SELECT a.thing, b.thing, c.thing FROM a, b, c WHERE a.id = b.id")
	checkNil(t, err)
	_, err = db.Query("SELECT a.thing, b.thing, c.thing FROM a JOIN b ON a.id = b.id JOIN c ON a.id = c.id")
	checkNil(t, err)
}

func TestSelectColumnNames(t *testing.T) {
	defer mogi.Reset()
	db := openDB()

	// qualified names
	mogi.Select("a.thing", "b.thing", "c.thing").From("qqqq", "b", "c").StubCSV(`foo,bar,baz`)
	_, err := db.Query("SELECT a.thing, b.thing, c.thing FROM qqqq as a, b, c WHERE a.id = b.id")
	checkNil(t, err)

	// aliased names
	mogi.Reset()
	mogi.Select("dog", "cat", "hamster").From("a", "b", "c").StubCSV(`foo,bar,baz`)
	_, err = db.Query("SELECT a.thing AS dog, b.thing AS cat, c.thing AS hamster FROM a JOIN b ON a.id = b.id JOIN c ON a.id = c.id")
	checkNil(t, err)
}

func TestSelectCount(t *testing.T) {
	defer mogi.Reset()
	db := openDB()

	mogi.Select("COUNT(abc)", "count(*)").StubCSV("1,5")
	_, err := db.Query("SELECT COUNT(abc), COUNT(*) FROM beer")
	checkNil(t, err)
}

func TestSelectWhereIn(t *testing.T) {
	defer mogi.Reset()
	db := openDB()

	mogi.Select().Where("pct", 5.4, 10.2).StubCSV("2")
	_, err := db.Query("SELECT COUNT(*) FROM beer WHERE pct IN (5.4, ?)", 10.2)
	checkNil(t, err)

	mogi.Reset()
	mogi.Select().WhereOp("pct", "IN", 5.4, 10.2).StubCSV("2")
	_, err = db.Query("SELECT COUNT(*) FROM beer WHERE pct IN (5.4, ?)", 10.2)
	checkNil(t, err)
}

func TestSelectStar(t *testing.T) {
	defer mogi.Reset()
	db := openDB()

	mogi.Select("*").StubCSV("a,b,c")
	_, err := db.Query("SELECT * FROM beer")
	checkNil(t, err)
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
