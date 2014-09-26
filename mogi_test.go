package mogi_test

import (
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/guregu/mogi"
)

func TestMogi(t *testing.T) {
	defer mogi.Reset()
	mogi.Verbose(false)
	db := openDB()

	// select (any columns)
	mogi.Select().StubCSV(beerCSV)
	runBeerSelectQuery(t, db)

	// test .Stub()
	mogi.Select().Stub([][]driver.Value{
		{1, "Yona Yona Ale", "Yo-Ho Brewing", 5.5},
		{2, "Punk IPA", "BrewDog", 5.6}})
	runBeerSelectQuery(t, db)

	// test reset
	mogi.Reset()
	_, err := db.Query("SELECT id, name, brewery, pct FROM beer WHERE pct > ?", 5)
	if err != mogi.ErrUnstubbed {
		t.Error("after reset, err should be ErrUnstubbed but is", err)
	}

	// select specific columns
	mogi.Select("id", "name", "brewery", "pct").StubCSV(beerCSV)
	runBeerSelectQuery(t, db)

	// select the "wrong" columns
	mogi.Reset()
	mogi.Select("hello", "👞").StubCSV(beerCSV)
	runUnstubbedSelect(t, db)
}

func TestNotify(t *testing.T) {
	defer mogi.Reset()
	db := openDB()

	ch := make(chan struct{})

	mogi.Insert().Into("beer").Notify(ch).StubResult(3, 1)
	_, err := db.Exec("INSERT INTO beer (name, brewery, pct) VALUES (?, ?, ?)", "Mikkel’s Dream", "Mikkeller", 4.6)
	checkNil(t, err)

	<-ch
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
