package mogi_test

import (
	"testing"

	"github.com/guregu/mogi"
)

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
	mogi.Delete().Table("beer").WhereOp("id", "=", 42).StubRowsAffected(1)
	_, err = db.Exec("DELETE FROM beer WHERE id = ?", 42)
	checkNil(t, err)

	mogi.Reset()
	mogi.Delete().Table("beer").Where("id", 50).StubRowsAffected(1)
	_, err = db.Exec("DELETE FROM beer WHERE id = ?", 42)
	if err != mogi.ErrUnstubbed {
		t.Error("err should be ErrUnstubbed but is", err)
	}
}
