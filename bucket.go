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
	pool     map[string]*DB
	interval time.Duration
	mu       *sync.Mutex
}

// NewBucket init a pool of *DB, w/ a "duration"
// interval to call cleaner func
func NewBucket(interval time.Duration) *Bucket {
	return &Bucket{
		pool:     make(map[string]*DB),
		interval: interval,
		mu:       &sync.Mutex{},
	}
}

// CreateDB := create a new database
// with given "name"
func (b *Bucket) CreateDB(name string) (*DB, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	// check if database already exists
	// if found, must choose a different name
	if _, ok := b.pool[name]; ok {
		return nil, errFound
	}
	//
	db := &DB{
		name:   name,
		table:  map[string]*value{},
		stop:   make(chan bool, 1),
		ticker: time.NewTicker(b.interval),
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
	b.mu.Lock()
	defer b.mu.Unlock()
	//
	db, ok := b.pool[name]
	if !ok {
		return nil, errNotFound
	}
	//
	return db, nil
}

/*
EmptyDB := empty all data from a database.
*/
func (b *Bucket) EmptyDB(name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	//
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
	b.mu.Lock()
	defer b.mu.Unlock()
	//
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
	b.mu.Lock()
	defer b.mu.Unlock()
	//
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
