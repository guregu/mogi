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

// Pass multiple arguments to Where() for IN clauses. 
mogi.Select().Where("id", 10, 42).StubCSV("Apex\nWestvleteren XII")
rows, err = db.Query("SELECT name FROM beer WHERE id IN (?, ?)", 10, 42)

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

// Filter by the table used in the query
mogi.Insert().Into("beer").StubResult(1, 1)

// Filter by the args passed to the query (the things replacing the ?s)
mogi.Insert().Args("Yona Yona Ale", "Yo-Ho Brewing", 5.5).StubResult(1, 1)

// Filter by the values used in the query
mogi.Insert().Value("name", "Yona Yona Ale").Value("brewery", "Yo-Ho Brewing").StubResult(1, 1)
// Use ValueAt when you are inserting multiple rows. The first argument is the row #, starting with 0.
// Parameters are interpolated for you.
mogi.Insert().
	ValueAt(0, "brewery", "Mikkeller").ValueAt(0, "pct", 4.6).
	ValueAt(1, "brewery", "BrewDog").ValueAt(1, "pct", 18.2).
	StubResult(4, 2)
result, err = db.Exec(`INSERT INTO beer (name, brewery, pct) VALUES (?, "Mikkeller", 4.6), (?, ?, ?)`,
	"Mikkel’s Dream",
	"Tokyo*", "BrewDog", 18.2,
)
```

#### Stubbing UPDATE queries
```go
// Stub any UPDATE query
// UPDATE stubs work the same as INSERT stubs
// This stubs all UPDATE queries to return 10 rows affected
mogi.Update().StubResult(-1, 10)
// This does the same thing
mogi.Update().StubRowsAffected(10)

// Filter by the columns used in the SET clause
mogi.Update("name", "brewery", "pct").StubRowsAffected(1)
_, err := db.Exec(`UPDATE beer
				   SET name = "Mikkel’s Dream", brewery = "Mikkeller", pct = 4.6
				   WHERE id = ? AND moon = ?`, 3, "full")

// Filter by values set by the SET clause
mogi.Update().Value("name", "Mikkel’s Dream").Value("brewery", "Mikkeller").StubRowsAffected(1)

// Filter by args (? placeholder values)
mogi.Update().Args(3, "full").StubRowsAffected(1)

// Filter by the table being updated
mogi.Update().Table("beer").StubRowsAffected(1)

// Filter by WHERE clause params
mogi.Update().Where("id", 3).Where("moon", "full").StubRowsAffected(1)
```

#### Other stuff

##### Reset
You can remove all the stubs you've set with `mogi.Reset()`.

##### Verbose
`mogi.Verbose(true)` will enable verbose mode, logging unstubbed queries. 

##### Parse time
Set the time layout with `mogi.ParseTime()`. CSV values matching that layout will be converted to time.Time. 
You can also stub time.Time directly using the `Stub()` method. 
```go
mogi.ParseTime(time.RFC3339)
mogi.Select("release").
		From("beer").
		Where("id", 42).
		StubCSV(`2014-06-30T12:00:00Z`)
```

##### Dump stubs
Dump all the stubs with `mogi.Dump()`. It will print something like this:
```
Query stubs: (1 total)
#1     [3]     SELECT (any)        [+1]
               FROM device_tokens  [+1]
               WHERE user_id ≈ 42  [+1]
Exec stubs: (3 total)
#1     [3]     INSERT (any)                             [+1]
               TABLE device_tokens                      [+1]
               VALUE device_type ≈ gunosy_lite (row 0)  [+1]
#2     [2]     INSERT (any)                             [+1]
               TABLE device_tokens                      [+1]
#3     [-995]  INSERT a, b, c                           [+4]
               PRIORITY                                 [-999]
```
This is helpful when you're debugging and need to double-check the priorities and conditions you've stubbed. 
The numbers in [brackets] are the priorities. 
You can also add `Dump()` to a stub condition chain. It will dump lots of information about the query when matched. 

### License
BSD