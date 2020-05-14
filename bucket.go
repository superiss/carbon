package carbon

import (
	"errors"
	"sync"
	"time"
)

// Errors
var (
	errFound    error = errors.New("database already exists")
	errNotFound error = errors.New("no database found")
)

// Bucket struct
type Bucket struct {
	pool map[string]*DB
}

// NewBucket init a pool of *DB
func NewBucket() *Bucket {
	return &Bucket{
		pool: make(map[string]*DB),
	}
}

// CreateDB := create a new database
// with given "name" and "duration"
// interval to call cleaner func
func (b *Bucket) CreateDB(name string, duration int) (*DB, error) {
	// check if database already exists
	_, ok := b.pool[name]

	// if found, must choose a different name
	if ok {
		return nil, errFound
	}
	// convert from nanoseconds to seconds
	interval := time.Duration(duration * 1e+9)
	//
	db := &DB{
		name:   name,
		table:  map[string]*value{},
		keeper: map[string]*timer{},
		stop:   make(chan bool, 1),
		ticker: time.NewTicker(interval),
		mu:     &sync.RWMutex{},
	}
	// add DB to pool
	b.pool[name] = db
	//
	go db.clean()
	//
	return db, nil
}

/*
FindDB := find a database in pool by "name".
*/
func (b *Bucket) FindDB(name string) (*DB, error) {
	db, ok := b.pool[name]
	if !ok {
		return nil, errNotFound
	}
	return db, nil
}

/*
EmptyDB := empty all data from a database.
*/
func (b *Bucket) EmptyDB(name string) error {
	db, ok := b.pool[name]
	if !ok {
		return errNotFound
	}
	// reset values
	db.reset()
	//
	return nil
}

/*
RemoveDB := completely remove database from pool.
*/
func (b *Bucket) RemoveDB(name string) error {
	db, ok := b.pool[name]
	if !ok {
		return errNotFound
	}

	// stop DB and it cleaner
	db.shutdown()
	// delete it found
	delete(b.pool, name)

	//
	return nil
}

// Stop method
func (b *Bucket) Stop() {
	for name, db := range b.pool {
		// stop all DB and it cleaners.
		db.shutdown()
		delete(b.pool, name)
	}
}

// BucketStats struct
type BucketStats struct {
	TotalDB   int
	TotalSize int
	SizePerDB map[string]int
}

// Stats := show current statistic
func (b *Bucket) Stats() *BucketStats {
	var (
		ts      = 0
		details = map[string]int{}
	)
	//
	for name, db := range b.pool {
		size := db.size()
		ts = ts + size
		details[name] = size
	}
	//
	return &BucketStats{
		TotalDB:   len(b.pool),
		TotalSize: ts,
		SizePerDB: details,
	}
}
