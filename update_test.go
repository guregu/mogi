package mogi_test

import (
	"testing"
	"time"

	"github.com/guregu/mogi"
)

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

	// time.Time
	mogi.Reset()
	mogi.ParseTime(time.RFC3339)
	now := time.Now()
	mogi.Update().Value("updated_at", now).Value("brewery", "Mikkeller").Value("pct", 4.6).StubRowsAffected(1)
	_, err = db.Exec(`UPDATE beer
					   SET updated_at = ?, brewery = "Mikkeller", pct = ?
					   WHERE id = 3`, now, 4.6)
	checkNil(t, err)

	// boolean as 1 vs true
	mogi.Reset()
	mogi.Update().Value("awesome", true).StubRowsAffected(1)
	_, err = db.Exec(`UPDATE beer
					   SET awesome = 1
					   WHERE id = 3`)
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
