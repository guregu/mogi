package mogi_test

import (
	"testing"

	"github.com/guregu/mogi"
)

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
