package tarus_store

import (
	bolt "go.etcd.io/bbolt"
)

// DB represents a metadata database backed by a bolt
// database. The database is fully namespaced and stores
// image, container, namespace, snapshot, and content data
// while proxying data shared across namespaces to backend
// datastores for content and snapshots.
type DB struct {
	db *bolt.DB

	// wlock is used to protect access to the data structures during garbage
	// collection. While the wlock is held no writable transactions can be
	// opened, preventing changes from occurring between the mark and
	// sweep phases without preventing read transactions.
	// wlock sync.RWMutex
}

// NewDB creates a new metadata database using the provided
// bolt database.
func NewDB(db *bolt.DB) *DB {
	m := &DB{
		db: db,
	}

	return m
}

// View runs a readonly transaction on the metadata store.
func (m *DB) View(fn func(*bolt.Tx) error) error {
	return m.db.View(fn)
}

// Update runs a writable transaction on the metadata store.
func (m *DB) Update(fn func(*bolt.Tx) error) error {
	return m.db.Update(fn)
}
