package scylladb

import (
	"testing"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
)

func TestGocqlSessionAdapter_Query(t *testing.T) {
	mockSession := new(MockSession)
	mockQuery := new(MockQuery)

	mockSession.On("Query", "SELECT * FROM table WHERE id = ?", []interface{}{1}).Return(mockQuery)

	adapter := &gocqlSessionAdapter{Session: &gocql.Session{}}
	query := adapter.Query("SELECT * FROM table WHERE id = ?", 1)

	assert.IsType(t, &gocqlQueryAdapter{}, query)
}

func TestGocqlQueryAdapter_Consistency(t *testing.T) {
	mockQuery := new(MockQuery)

	mockQuery.On("Consistency", gocql.One).Return(mockQuery)

	adapter := &gocqlQueryAdapter{Query: &gocql.Query{}}
	query := adapter.Consistency(gocql.One)

	assert.IsType(t, &gocqlQueryAdapter{}, query)
}

func TestGocqlIterAdapter_Scan(t *testing.T) {
	mockIter := new(MockIter)
	dest := []interface{}{new(string)}

	mockIter.On("Scan", dest).Return(true)

	adapter := &gocqlIterAdapter{Iter: &gocql.Iter{}}
	result := adapter.Scan(dest...)

	assert.False(t, result)
}

func TestGocqlIterAdapter_Close(t *testing.T) {
	mockIter := new(MockIter)

	mockIter.On("Close").Return(nil)

	adapter := &gocqlIterAdapter{Iter: &gocql.Iter{}}
	err := adapter.Close()

	assert.NoError(t, err)
}
