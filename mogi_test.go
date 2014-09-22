package mogi_test

import (
	"database/sql"
	"database/sql/driver"
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

func TestMultipleTables(t *testing.T) {
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

func TestColumnNames(t *testing.T) {
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

func TestInsert(t *testing.T) {
	//mogi.Insert().Into("device_tokens").Expect("device_type", "gunosy_lite").StubResult(1337, 1)
	defer mogi.Reset()
	db := openDB()

	// naked INSERT stub
	mogi.Insert().StubResult(3, 1)
	_, err := db.Exec("INSERT INTO beer (name, brewery, pct) VALUES (?, ?, ?)", "Mikkel’s Dream", "Mikkeller", 4.6)
	checkNil(t, err)

	// INSERT with columns
	mogi.Reset()
	mogi.Insert("name", "brewery", "pct").Into("beer").StubResult(3, 1)
	_, err = db.Exec("INSERT INTO beer (name, brewery, pct) VALUES (?, ?, ?)", "Mikkel’s Dream", "Mikkeller", 4.6)
	checkNil(t, err)

	// INSERT with wrong columns
	mogi.Reset()
	mogi.Insert("犬", "🐱", "かっぱ").Into("beer").StubResult(3, 1)
	_, err = db.Exec("INSERT INTO beer (name, brewery, pct) VALUES (?, ?, ?)", "Mikkel’s Dream", "Mikkeller", 4.6)
	if err != mogi.ErrUnstubbed {
		t.Error("err should be ErrUnstubbed but is", err)
	}
}

func TestInsertArgs(t *testing.T) {
	defer mogi.Reset()
	db := openDB()

	mogi.Insert().Args("Mikkel’s Dream", "Mikkeller", 4.6).StubResult(3, 1)
	_, err := db.Exec("INSERT INTO beer (name, brewery, pct) VALUES (?, ?, ?)", "Mikkel’s Dream", "Mikkeller", 4.6)
	checkNil(t, err)

	// wrong args
	mogi.Reset()
	mogi.Insert().Args("Nodogoshi", "Kirin", 5).StubResult(4, 1)
	_, err = db.Exec("INSERT INTO beer (name, brewery, pct) VALUES (?, ?, ?)", "Mikkel’s Dream", "Mikkeller", 4.6)
	if err != mogi.ErrUnstubbed {
		t.Error("err should be ErrUnstubbed but is", err)
	}
}

func TestInsertInto(t *testing.T) {
	defer mogi.Reset()
	db := openDB()

	mogi.Insert().Into("beer").StubResult(3, 1)
	_, err := db.Exec("INSERT INTO beer (name, brewery, pct) VALUES (?, ?, ?)", "Mikkel’s Dream", "Mikkeller", 4.6)
	checkNil(t, err)
	// make sure .Into() and .Table() are the same
	mogi.Reset()
	mogi.Insert().Table("beer").StubResult(3, 1)
	_, err = db.Exec("INSERT INTO beer (name, brewery, pct) VALUES (?, ?, ?)", "Mikkel’s Dream", "Mikkeller", 4.6)
	checkNil(t, err)
}

func TestStubResult(t *testing.T) {
	defer mogi.Reset()
	db := openDB()

	mogi.Insert().StubResult(3, 1)
	res, err := db.Exec("INSERT INTO beer (name, brewery, pct) VALUES (?, ?, ?)", "Mikkel’s Dream", "Mikkeller", 4.6)
	checkNil(t, err)
	lastID, err := res.LastInsertId()
	checkNil(t, err)
	if lastID != 3 {
		t.Error("LastInsertId() should be 3 but is", lastID)
	}
	affected, err := res.RowsAffected()
	checkNil(t, err)
	if affected != 1 {
		t.Error("RowsAffected() should be 1 but is", affected)
	}
}

func TestStubResultWithErrors(t *testing.T) {
	defer mogi.Reset()
	db := openDB()

	mogi.Insert().StubResult(-1, -1)
	res, err := db.Exec("INSERT INTO beer (name, brewery, pct) VALUES (?, ?, ?)", "Mikkel’s Dream", "Mikkeller", 4.6)
	checkNil(t, err)
	_, err = res.LastInsertId()
	if err == nil {
		t.Error("error is nil but shouldn't be:", err)
	}
	_, err = res.RowsAffected()
	if err == nil {
		t.Error("error is nil but shouldn't be:", err)
	}
}

func TestInsertValues(t *testing.T) {
	defer mogi.Reset()
	db := openDB()

	// single row
	mogi.Insert().Value("brewery", "Mikkeller").Value("pct", 4.6).StubResult(3, 1)
	_, err := db.Exec("INSERT INTO beer (name, brewery, pct) VALUES (?, ?, ?)", "Mikkel’s Dream", "Mikkeller", 4.6)
	checkNil(t, err)

	// multiple rows
	mogi.Reset()
	mogi.Insert().
		ValueAt(0, "brewery", "Mikkeller").ValueAt(0, "pct", 4.6).
		ValueAt(1, "brewery", "BrewDog").ValueAt(1, "pct", 18.2).
		StubResult(4, 2)
	_, err = db.Exec(`INSERT INTO beer (name, brewery, pct) VALUES (?, "Mikkeller", 4.6), (?, ?, ?)`,
		"Mikkel’s Dream",
		"Tokyo*", "BrewDog", 18.2,
	)
	checkNil(t, err)
}

func TestUpdate(t *testing.T) {
	defer mogi.Reset()
	db := openDB()

	// naked update
	mogi.Update().StubResult(-1, 1)
	_, err := db.Exec(`UPDATE beer
					   SET name = "Mikkel’s Dream", brewery = "Mikkeller", pct = 4.6
					   WHERE id = 3`)
	checkNil(t, err)

	// update with cols
	mogi.Reset()
	mogi.Update("name", "brewery", "pct").StubResult(-1, 1)
	_, err = db.Exec(`UPDATE beer
					   SET name = "Mikkel’s Dream", brewery = "Mikkeller", pct = 4.6
					   WHERE id = 3`)
	checkNil(t, err)

	// with wrong cols
	mogi.Reset()
	mogi.Update("犬", "🐱", "かっぱ").StubResult(-1, 1)
	_, err = db.Exec(`UPDATE beer
					   SET name = "Mikkel’s Dream", brewery = "Mikkeller", pct = 4.6
					   WHERE id = 3`)
	if err != mogi.ErrUnstubbed {
		t.Error("err should be ErrUnstubbed but is", err)
	}
}

func TestUpdateTable(t *testing.T) {
	defer mogi.Reset()
	db := openDB()

	// table
	mogi.Update().Table("beer").StubResult(-1, 1)
	_, err := db.Exec(`UPDATE beer
					   SET name = "Mikkel’s Dream", brewery = "Mikkeller", pct = 4.6
					   WHERE id = 3`)
	checkNil(t, err)

	// with wrong table
	mogi.Reset()
	mogi.Update().Table("酒").StubResult(-1, 1)
	_, err = db.Exec(`UPDATE beer
					   SET name = "Mikkel’s Dream", brewery = "Mikkeller", pct = 4.6
					   WHERE id = 3`)
	if err != mogi.ErrUnstubbed {
		t.Error("err should be ErrUnstubbed but is", err)
	}
}

func TestUpdateValues(t *testing.T) {
	defer mogi.Reset()
	db := openDB()

	mogi.Update().Value("name", "Mikkel’s Dream").Value("brewery", "Mikkeller").Value("pct", 4.6).StubRowsAffected(1)
	_, err := db.Exec(`UPDATE beer
					   SET name = "Mikkel’s Dream", brewery = "Mikkeller", pct = ?
					   WHERE id = 3`, 4.6)
	checkNil(t, err)

	// with wrong values
	mogi.Reset()
	mogi.Update().Value("name", "7-Premium THE BREW").Value("brewery", "Suntory").Value("pct", 5.0).StubResult(-1, 1)
	_, err = db.Exec(`UPDATE beer
					   SET name = "Mikkel’s Dream", brewery = "Mikkeller", pct = ?
					   WHERE id = 3`, 4.6)
	if err != mogi.ErrUnstubbed {
		t.Error("err should be ErrUnstubbed but is", err)
	}
}

func TestUpdateWhere(t *testing.T) {
	defer mogi.Reset()
	db := openDB()

	mogi.Update().Where("id", 3).Where("moon", "full").StubResult(-1, 1)
	_, err := db.Exec(`UPDATE beer
					   SET name = "Mikkel’s Dream", brewery = "Mikkeller", pct = ?
					   WHERE id = ? AND moon = "full"`, 4.6, 3)
	checkNil(t, err)

	mogi.Reset()
	mogi.Update().Where("foo", 555).Where("bar", "qux").StubResult(-1, 1)
	_, err = db.Exec(`UPDATE beer
					   SET name = "Mikkel’s Dream", brewery = "Mikkeller", pct = ?
					   WHERE id = 3 AND moon = "full"`, 4.6)
	if err != mogi.ErrUnstubbed {
		t.Error("err should be ErrUnstubbed but is", err)
	}
}

func TestSelectCount(t *testing.T) {
	defer mogi.Reset()
	db := openDB()

	mogi.Select("COUNT(abc)", "COUNT(*)").StubCSV("")
	_, err := db.Query("SELECT COUNT(abc), COUNT(*) FROM beer")
	checkNil(t, err)
}

func TestSelectWhereIn(t *testing.T) {
	defer mogi.Reset()
	db := openDB()

	mogi.Select().Where("pct", 5.4, 10.2).StubCSV("2")
	_, err := db.Query("SELECT COUNT(*) FROM beer WHERE pct IN (5.4, ?)", 10.2)
	checkNil(t, err)
}

func TestSelectStar(t *testing.T) {
	defer mogi.Reset()
	db := openDB()

	mogi.Select("*").StubCSV("a,b,c")
	_, err := db.Query("SELECT * FROM beer")
	checkNil(t, err)
}

func TestDelete(t *testing.T) {
	defer mogi.Reset()
	db := openDB()

	mogi.Delete().StubRowsAffected(1)
	_, err := db.Exec("DELETE FROM beer WHERE id = ?", 42)
	checkNil(t, err)

	mogi.Reset()
	mogi.Delete().Table("beer").StubRowsAffected(1)
	_, err = db.Exec("DELETE FROM beer WHERE id = ?", 42)
	checkNil(t, err)

	mogi.Reset()
	mogi.Delete().Table("beer").Where("id", 42).StubRowsAffected(1)
	_, err = db.Exec("DELETE FROM beer WHERE id = ?", 42)
	checkNil(t, err)

	mogi.Reset()
	mogi.Delete().Table("beer").Where("id", 50).StubRowsAffected(1)
	_, err = db.Exec("DELETE FROM beer WHERE id = ?", 42)
	if err != mogi.ErrUnstubbed {
		t.Error("err should be ErrUnstubbed but is", err)
	}
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
