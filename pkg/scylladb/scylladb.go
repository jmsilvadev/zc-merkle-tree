package scylladb

import (
	"time"

	"github.com/gocql/gocql"
)

// DB is the structure that represents the database
type DB struct {
	session Session
}

// New initializes a new ScyllaDB database connection
func New(hosts []string, keyspace string) (*DB, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = 10 * time.Second

	adminSession, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	defer adminSession.Close()

	err = adminSession.Query(`
		CREATE KEYSPACE IF NOT EXISTS ` + keyspace + ` WITH replication = {
			'class': 'SimpleStrategy',
			'replication_factor': '1'
		}`).Exec()
	if err != nil {
		return nil, err
	}

	cluster.Keyspace = keyspace
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	err = session.Query(`
		CREATE TABLE IF NOT EXISTS kv (
			key text PRIMARY KEY,
			value blob
		)`).Exec()
	if err != nil {
		session.Close()
		return nil, err
	}

	return &DB{session: &gocqlSessionAdapter{session}}, nil
}

// Get returns the value associated with the given key
func (d *DB) Get(key string) ([]byte, error) {
	var value []byte
	err := d.session.Query(`SELECT value FROM kv WHERE key = ?`, key).Scan(&value)
	if err != nil {
		return nil, err
	}
	return value, nil
}

// Put inserts a key-value pair into the database
func (d *DB) Put(key string, value []byte) error {
	err := d.session.Query(`INSERT INTO kv (key, value) VALUES (?, ?)`, key, value).Exec()
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a key-value pair from the database
func (d *DB) Delete(key string) error {
	err := d.session.Query(`DELETE FROM kv WHERE key = ?`, key).Exec()
	if err != nil {
		return err
	}
	return nil
}

// GetByPrefix returns all key-value pairs that have the given prefix
func (d *DB) GetByPrefix(prefix string) (map[string][]byte, error) {
	result := make(map[string][]byte)

	query := `SELECT key, value FROM kv WHERE key >= ? AND key < ? ALLOW FILTERING`
	iter := d.session.Query(query, prefix, prefix+"\uFFFF").Iter()

	var key string
	var value []byte

	for iter.Scan(&key, &value) {
		v := make([]byte, len(value))
		copy(v, value)
		result[key] = v
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteByPrefix deletes all key-value pairs that have the given prefix
func (d *DB) DeleteByPrefix(prefix string) error {
	keys, err := d.GetByPrefix(prefix)
	if err != nil {
		return err
	}

	for k := range keys {
		query := `DELETE FROM kv WHERE key = ?`
		err := d.session.Query(query, k).Exec()
		if err != nil {
			return err
		}
	}
	return nil
}

// Close closes the database connection
func (d *DB) Close() error {
	d.session.Close()
	return nil
}
