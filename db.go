package carbon

import (
	"sync"
	"time"
)

// DB struct
type DB struct {
	name   string
	table  map[string]*value // storage map
	keeper map[string]*timer // cleaning map
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
			for key, t := range db.keeper {
				// if true delete value from storage map
				// delete value from keeper map
				if now.UnixNano()-t.at.UnixNano() > int64(t.ttl) {
					delete(db.table, key)
					delete(db.keeper, key)
				}
			}
			db.mu.Unlock()
			//
		case <-db.stop:
			// stop ticker
			// exist the loop
			db.ticker.Stop()
			return
		}
	}
}

// value struct
type value struct {
	data []byte
	time *timer
}

// timeer struct
type timer struct {
	ttl time.Duration
	at  time.Time
}

// Set method
func (db *DB) Set(key string, data []byte, ttl time.Duration) {
	db.mu.Lock()
	// timer
	t := &timer{
		ttl: ttl,
		at:  time.Now(),
	}
	// value
	val := &value{
		data: data,
		time: t,
	}
	//
	db.table[key] = val // set value to DB
	db.keeper[key] = t  // set key and timer to keeper DB
	db.mu.Unlock()
}

// Get method
func (db *DB) Get(key string) ([]byte, bool) {
	db.mu.RLock()
	val, ok := db.table[key]
	db.mu.RUnlock()
	//
	if !ok {
		return nil, ok
	}
	return val.data, ok
}

// reset table & keeper
func (db *DB) reset() {
	db.mu.Lock()
	db.table = map[string]*value{}
	db.keeper = map[string]*timer{}
	db.mu.Unlock()
}

// Stop method
func (db *DB) shutdown() {
	db.stop <- true
	close(db.stop)
}

// stats method
func (db *DB) size() (size int) {
	db.mu.Lock()
	size = len(db.table)
	db.mu.Unlock()
	return
}
