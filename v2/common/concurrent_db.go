package common

import (
	"sync"
)

type ConcurrentDB struct {
	store DB
	lock  sync.RWMutex
}

func NewConcurrentDB(parent DB) DB {
	return &ConcurrentDB{
		store: parent,
	}
}

func (db *ConcurrentDB) Save(path string, v interface{}) {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.store.Save(path, v)
}

func (db *ConcurrentDB) Retrive(path string) interface{} {
	db.lock.Lock()
	defer db.lock.Unlock()

	v := db.store.Retrive(path)
	return v
}

func (db *ConcurrentDB) BatchSave(paths []string, dict interface{}) {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.store.BatchSave(paths, dict)
}

func (db *ConcurrentDB) Print() {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.store.Print()
}

func (db *ConcurrentDB) Clear() {}
