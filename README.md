## mogi [![GoDoc](https://godoc.org/github.com/guregu/mogi?status.svg)](https://godoc.org/github.com/guregu/mogi) [![Coverage](http://gocover.io/_badge/github.com/guregu/mogi)](http://gocover.io/github.com/guregu/mogi)
`import "github.com/guregu/mogi"`

mogi is a fancy SQL mocking/stubbing library for Go. It uses the [vitess](https://github.com/youtube/vitess) SQL parser for maximum happiness. 

It's not finished yet. Complex queries will break it. Stay tuned! 

### Usage

#### Getting started
```go
import 	"github.com/guregu/mogi"
db, _ := sql.Open("mogi", "")
```

#### Stubbing SELECT queries
```go
// Stub any SELECT query
mogi.Select().StubCSV(`1,Yona Yona Ale,Yo-Ho Brewing,5.5`)
rows, err := db.Query("SELECT id, name, brewery, pct FROM beer")

// Reset to clear all stubs
mogi.Reset()

// Stub SELECT queries by columns selected
mogi.Select("id", "name", "brewery", "pct").StubCSV(`1,Yona Yona Ale,Yo-Ho Brewing,5.5`)
// Aliased columns should be given as they are aliased.
// Qualified columns should be given as they are qualified. 
// e.g. SELECT beer.name AS n, breweries.founded FROM beer JOIN breweries ON beer.brewery = breweries.name
mogi.Select("n", "breweries.founded").StubCSV(`Stone IPA,1996`)

// You can stub with driver.Values instead of CSV
mogi.Select("id", "deleted_at").Stub([][]driver.Value{{1, nil}})

// Filter by table name
mogi.Select().From("beer").StubCSV(`1,Yona Yona Ale,Yo-Ho Brewing,5.5`)

// You can supply multiple table names for JOIN queries
// e.g. SELECT beer.name, wine.name FROM beer JOIN wine ON beer.pct = wine.pct
// or   SELECT beer.name, wine.name FROM beer, wine WHERE beer.pct = wine.pct
mogi.Select().From("beer", "wine").StubCSV(`Westvleteren XII,Across the Pond Riesling`)

// Filter by WHERE clause params
mogi.Select().Where("id", 10).StubCSV(`10,Apex,Bear Republic Brewing Co.,8.95`)
mogi.Select().Where("id", 42).StubCSV(`42,Westvleteren XII,Brouwerij Westvleteren,10.2`)
rows, err := db.Query("SELECT id, name, brewery, pct FROM beer WHERE id = ?", 10)
...
rows, err = db.Query("SELECT id, name, brewery, pct FROM beer WHERE id = ?", 42)
...

// Stub an error while you're at it
mogi.Select().Where("id", 3).StubError(sql.ErrNoRows)
// FYI, unstubbed queries will return mogi.ErrUnstubbed

// Filter by args given 
mogi.Select().Args(1).StubCSV(`1,Yona Yona Ale,Yo-Ho Brewing,5.5`)
rows, err := db.Query("SELECT id, name, brewery, pct FROM beer WHERE id = ?", 1)

// Chain filters as much as you'd like
mogi.Select("id", "name", "brewery", "pct").From("beer").Where("id", 1).StubCSV(`1,Yona Yona Ale,Yo-Ho Brewing,5.5`)
```

### License
BSD