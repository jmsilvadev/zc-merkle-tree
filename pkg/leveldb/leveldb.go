package leveldb

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// DB is the structure that represents the database
type DB struct {
	db *leveldb.DB
}

// New initializes a new LevelDB database at the given path
func New(path string) (*DB, error) {
	db, err := leveldb.OpenFile(path, &opt.Options{
		ErrorIfMissing: false, // Create the database if it does not exist
	})
	if err != nil {
		return nil, err
	}
	return &DB{db: db}, nil
}

// Get returns the value associated with the given key
func (d *DB) Get(key string) ([]byte, error) {
	data, err := d.db.Get([]byte(key), nil)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Put inserts a key-value pair into the database
func (d *DB) Put(key string, value []byte) error {
	err := d.db.Put([]byte(key), value, nil)
	if err != nil {
		return err
	}
	return nil
}

// GetByPrefix returns all key-value pairs that have the given prefix
func (d *DB) GetByPrefix(prefix string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	iter := d.db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	defer iter.Release()

	for iter.Next() {
		key := string(iter.Key())
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value()) // need to copy due leveldb internal use
		result[key] = value
	}

	if err := iter.Error(); err != nil {
		return nil, err
	}
	return result, nil
}

// Delete deletes a key-value pair from the database
func (d *DB) Delete(key string) error {
	err := d.db.Delete([]byte(key), nil)
	if err != nil {
		return err
	}
	return nil
}

// DeleteByPrefix deletes all key-value pairs that have the given prefix
func (d *DB) DeleteByPrefix(prefix string) error {
	iter := d.db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	defer iter.Release()

	for iter.Next() {
		key := iter.Key()
		err := d.db.Delete(key, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

// Close closes the database
func (d *DB) Close() error {
	err := d.db.Close()
	if err != nil {
		return err
	}
	return nil
}
