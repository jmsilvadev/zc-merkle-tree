package scylladb

import (
	"testing"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSession is a mock of the Session interface
type MockSession struct {
	mock.Mock
}

func (m *MockSession) Query(stmt string, values ...interface{}) Query {
	args := m.Called(stmt, values)
	return args.Get(0).(Query)
}

func (m *MockSession) Close() {
	m.Called()
}

// MockQuery is a mock of the Query interface
type MockQuery struct {
	mock.Mock
}

func (m *MockQuery) Consistency(c gocql.Consistency) Query {
	m.Called(c)
	return m
}

func (m *MockQuery) Scan(dest ...interface{}) error {
	args := m.Called(dest)
	return args.Error(0)
}

func (m *MockQuery) Iter() Iter {
	args := m.Called()
	return args.Get(0).(Iter)
}

func (m *MockQuery) Exec() error {
	args := m.Called()
	return args.Error(0)
}

// MockIter is a mock of the Iter interface
type MockIter struct {
	mock.Mock
}

func (m *MockIter) Scan(dest ...interface{}) bool {
	args := m.Called(dest)
	return args.Bool(0)
}

func (m *MockIter) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNew(t *testing.T) {
	// To increase the coverage
	_, err := New(nil, "")
	require.Error(t, err)
}

func TestGet(t *testing.T) {
	mockSession := new(MockSession)
	mockQuery := new(MockQuery)

	mockSession.On("Query", `SELECT value FROM kv WHERE key = ?`, []interface{}{"test_key"}).Return(mockQuery)
	mockQuery.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(0).([]interface{})
		*arg[0].(*[]byte) = []byte("test_value")
	}).Return(nil)

	db := &DB{session: mockSession}
	value, err := db.Get("test_key")

	assert.NoError(t, err)
	assert.Equal(t, []byte("test_value"), value)

	mockSession.AssertExpectations(t)
	mockQuery.AssertExpectations(t)
}

func TestDelete(t *testing.T) {
	mockSession := new(MockSession)
	mockQuery := new(MockQuery)

	mockSession.On("Query", `DELETE FROM kv WHERE key = ?`, []interface{}{"test_key"}).Return(mockQuery)
	mockQuery.On("Exec").Return(nil)

	db := &DB{session: mockSession}
	err := db.Delete("test_key")

	assert.NoError(t, err)

	mockSession.AssertExpectations(t)
	mockQuery.AssertExpectations(t)
}

func TestPut(t *testing.T) {
	mockSession := new(MockSession)
	mockQuery := new(MockQuery)

	mockSession.On("Query", `INSERT INTO kv (key, value) VALUES (?, ?)`, []interface{}{"test_key", []byte("test_value")}).Return(mockQuery)
	mockQuery.On("Exec").Return(nil)

	db := &DB{session: mockSession}
	err := db.Put("test_key", []byte("test_value"))

	assert.NoError(t, err)

	mockSession.AssertExpectations(t)
	mockQuery.AssertExpectations(t)
}

func TestClose(t *testing.T) {
	mockSession := new(MockSession)

	mockSession.On("Close").Return()

	db := &DB{session: mockSession}
	err := db.Close()

	assert.NoError(t, err)

	mockSession.AssertExpectations(t)
}

func TestGetByPrefix(t *testing.T) {
	mockSession := new(MockSession)
	mockQuery := new(MockQuery)
	mockIter := new(MockIter)

	mockSession.On("Query", `SELECT key, value FROM kv WHERE key >= ? AND key < ? ALLOW FILTERING`, []interface{}{"test_prefix", "test_prefix\uFFFF"}).Return(mockQuery)
	mockQuery.On("Iter").Return(mockIter)
	mockIter.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(0).([]interface{})
		*arg[0].(*string) = "test_key"
		*arg[1].(*[]byte) = []byte("test_value")
	}).Return(true).Once()
	mockIter.On("Scan", mock.Anything).Return(false)
	mockIter.On("Close").Return(nil)

	db := &DB{session: mockSession}
	result, err := db.GetByPrefix("test_prefix")

	assert.NoError(t, err)
	assert.Equal(t, map[string][]byte{"test_key": []byte("test_value")}, result)

	mockSession.AssertExpectations(t)
	mockQuery.AssertExpectations(t)
	mockIter.AssertExpectations(t)
}

func TestDeleteByPrefix(t *testing.T) {
	mockSession := new(MockSession)
	mockQuery := new(MockQuery)
	mockIter := new(MockIter)

	mockSession.On("Query", `SELECT key, value FROM kv WHERE key >= ? AND key < ? ALLOW FILTERING`, []interface{}{"test_prefix", "test_prefix\uFFFF"}).Return(mockQuery)
	mockQuery.On("Iter").Return(mockIter)
	mockIter.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(0).([]interface{})
		*arg[0].(*string) = "test_key1"
		*arg[1].(*[]byte) = []byte("value1")
	}).Return(true).Once()
	mockIter.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(0).([]interface{})
		*arg[0].(*string) = "test_key2"
		*arg[1].(*[]byte) = []byte("value2")
	}).Return(true).Once()
	mockIter.On("Scan", mock.Anything).Return(false)
	mockIter.On("Close").Return(nil)

	mockSession.On("Query", `DELETE FROM kv WHERE key = ?`, []interface{}{"test_key1"}).Return(mockQuery)
	mockSession.On("Query", `DELETE FROM kv WHERE key = ?`, []interface{}{"test_key2"}).Return(mockQuery)
	mockQuery.On("Exec").Return(nil)

	db := &DB{session: mockSession}
	err := db.DeleteByPrefix("test_prefix")

	assert.NoError(t, err)

	mockSession.AssertExpectations(t)
	mockQuery.AssertExpectations(t)
	mockIter.AssertExpectations(t)
}
