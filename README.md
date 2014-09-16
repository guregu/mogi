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

#### Stubbing INSERT queries
```go
// Stub any INSERT query
// You can use StubResult to easily stub a driver.Result. 
// You can pass -1 to StubResult to have it return an error for that particular bit.
// In this example, we have 1 row affected, but no LastInsertID. 
mogi.Insert().StubResult(-1, 1)
// If you have your own driver.Result you want to pass, just use Stub.
// You can also stub an error with StubError. 

// Filter by the columns used in the INSERT query
mogi.Insert("name", "brewery", "pct").StubResult(1, 1)
result, err := db.Exec("INSERT INTO beer (name, brewery, pct) VALUES (?, ?, ?)", "Yona Yona Ale", "Yo-Ho Brewing", 5.5)

// Filter by the args passed to the query (the things replacing the ?s)
mogi.Insert().Args("Yona Yona Ale", "Yo-Ho Brewing", 5.5).StubResult(1, 1)

// Filter by the values used in the query
mogi.Insert().Value("name", "Yona Yona Ale").Value("brewery", "Yo-Ho Brewing").StubResult(1, 1)
// Use ValueN when you are inserting multiple rows. The first argument is the row #, starting with 0.
// Parameters are interpolated for you.
mogi.Insert().
	ValueN(0, "brewery", "Mikkeller").ValueN(0, "pct", 4.6).
	ValueN(1, "brewery", "BrewDog").ValueN(1, "pct", 18.2).
	StubResult(4, 2)
result, err = db.Exec(`INSERT INTO beer (name, brewery, pct) VALUES (?, "Mikkeller", 4.6), (?, ?, ?)`,
	"Mikkel’s Dream",
	"Tokyo*", "BrewDog", 18.2,
)
```

### License
BSD