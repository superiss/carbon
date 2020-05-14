# Carbon a Pool of key-value database
a Pool of key-value database with TTL, Fast & Concurrent Safe.

key(string)
value([]byte)

## Usage
To install:
`
go get github.com/superiss/carbon
`

## NewBucket:
create a new pool where database will be stored
```go
bucket := carbon.NewBucket() // new bucket
defer bucket.Stop() // always defer Stop()
//
db, err := bucket.CreateDB("cache_db", 1)
if err!=nil{
    // handle err
}
// set value with ttl
db.Set(s, []byte(s), 10*time.Minute)

// get value
value, ok := db.Get(s)
```
# (b *Bucket)Methods
## CreateDB(name string, duration int) (*DB, error)
Create a new DB w/ a given "name" and "duration" in seconds used for the cleaning interval 
```go
db, err := bucket.CreateDB("cache", 10) // the cleaner will clean the database every 10s
if err!=nil{
    // handle err
}
```

## FindDB(name string) (*DB, error)
find a DB by "name" and return *DB or Error if not found 
```go
db1, err := bucket.FindDB("cache")
if err!=nil{
    // handle err
}
```

## EmptyDB(name string) error
empty all data stored in a specifc DB, and returns an Error if not found
```go
if err := bucket.EmptyDB("cache"); err!=nil{
    // handle err
}
```

## RemoveDB(name string) error
remove a specifc DB completely, of return Error if not found
```go
if err := bucket.RemoveDB("cache"); err!=nil{
    // handle err
}
```

## Stop()
remove databases from the pool and stop all cleaners
```go
defer bucket.Stop()
```

# *DB Methods:
## Set(key string, data []byte, ttl time.Duration)
Set a new value into database with a time to expire (TTL)
```go
db.Set(s, []byte(s), 10*time.Minute)
```

## Get(key string) ([]byte, bool)
Get return value if found or return nil, false if not found
```go
value, ok := db.Get(s)
```

# Testing
## Benchmark test
```go
func BenchmarkBucketDB(b *testing.B) {
    bucket := carbon.NewBucket()
    defer bucket.Stop()
    //
    db, _ := bucket.CreateDB("cache", 1)
    for n := 0; n < b.N; n++ {
        s := fmt.Sprint(n)
        db.Set(s, []byte(s), 10*time.Minute)
        db.Get(s)
    }
    fmt.Println(bucket.Stats())
}
```
results on my PC

`
1000000              1606 ns/op             335 B/op          5 allocs/op
PASS
ok      carbon  1.680s
`

## Concurrent test
```go
bucket := carbon.NewBucket()
defer bucket.Stop()
//
db, _ := bucket.CreateDB("cache", 1)
//
for n := 0; n < 200000; n++ {
	s := fmt.Sprint(n)
	go func() {
		db.Set(s, []byte(s), 10*time.Minute)
		db.Get(s)
	}()
}
time.Sleep(3 * time.Second)
fmt.Println(bucket.Stats())
```

go run -race .

## Note
