package db

type Database interface {
	// Get returns the value associated with the given key
	Get(key string) ([]byte, error)
	// Put inserts a key-value pair into the database
	Put(key string, value []byte) error
	// Delete deletes a key-value pair from the database
	Delete(key string) error
	// DeleteByPrefix deletes all key-value pairs that have the given prefix
	DeleteByPrefix(prefix string) error
	// GetByPrefix returns all key-value pairs that have the given prefix
	GetByPrefix(prefix string) (map[string][]byte, error)
	// Close closes the database
	Close() error
}
