package scylladb

import "github.com/gocql/gocql"

// NOTE: Need to create this to be able to mock and run unit tests

// Session is an interface for gocql.Session
type Session interface {
	Query(stmt string, values ...interface{}) Query
	Close()
}

// Query is an interface for gocql.Query
type Query interface {
	Consistency(c gocql.Consistency) Query
	Scan(dest ...interface{}) error
	Iter() Iter
	Exec() error
}

// Iter is an interface for gocql.Iter
type Iter interface {
	Scan(dest ...interface{}) bool
	Close() error
}

// gocqlSessionAdapter wraps *gocql.Session to implement the Session interface
type gocqlSessionAdapter struct {
	*gocql.Session
}

func (s *gocqlSessionAdapter) Query(stmt string, values ...interface{}) Query {
	return &gocqlQueryAdapter{s.Session.Query(stmt, values...)}
}

// gocqlQueryAdapter wraps *gocql.Query to implement the Query interface
type gocqlQueryAdapter struct {
	*gocql.Query
}

func (q *gocqlQueryAdapter) Consistency(c gocql.Consistency) Query {
	q.Query = q.Query.Consistency(c)
	return q
}

func (q *gocqlQueryAdapter) Iter() Iter {
	return &gocqlIterAdapter{q.Query.Iter()}
}

// gocqlIterAdapter wraps *gocql.Iter to implement the Iter interface
type gocqlIterAdapter struct {
	*gocql.Iter
}

func (i *gocqlIterAdapter) Scan(dest ...interface{}) bool {
	return i.Iter.Scan(dest...)
}
