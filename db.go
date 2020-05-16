package carbon

import (
	"sync"
	"time"
)

// DB struct
type DB struct {
	name   string
	table  map[string]*value // storage map
	stop   chan bool
	ticker *time.Ticker
	mu     *sync.RWMutex
}

// clean method
func (db *DB) clean() {
	for {
		select {
		case now := <-db.ticker.C:
			db.mu.Lock()
			for key, t := range db.table {
				// if true delete value
				if now.UnixNano()-t.at.UnixNano() > int64(t.ttl) {
					delete(db.table, key)
				}
			}
			db.mu.Unlock()
			//
		case <-db.stop:
			// stop ticker, reset db and exist the loop
			db.ticker.Stop()
			db.reset()
			return
		}
	}
}

// value struct
type value struct {
	data []byte
	ttl  time.Duration
	at   time.Time
}

// Set method
func (db *DB) Set(key string, data []byte, ttl time.Duration) {
	db.mu.Lock()
	// value
	val := &value{
		data: data,
		ttl:  ttl,
		at:   time.Now(),
	}
	//
	db.table[key] = val // set value to DB
	db.mu.Unlock()
}

// Get method
func (db *DB) Get(key string) ([]byte, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	val, ok := db.table[key]
	//
	if !ok {
		return nil, ok
	}
	return val.data, ok
}

// size method
func (db *DB) size() (size int) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	//
	size = len(db.table)
	return
}

// reset table & keeper
func (db *DB) reset() {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.table = map[string]*value{}
}

// Stop method
func (db *DB) shutdown() {
	db.stop <- true
	close(db.stop)
}
