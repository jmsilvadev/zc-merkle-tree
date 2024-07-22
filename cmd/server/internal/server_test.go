package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jmsilvadev/zc/pkg/config"
	"github.com/jmsilvadev/zc/pkg/mkt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDatabase is a mock implementation of the db.Database interface
type MockDatabase struct {
	mock.Mock
	data map[string][]byte
}

func (m *MockDatabase) Put(key string, value []byte) error {
	m.Called(key, value)
	m.data[key] = value
	return nil
}

func (m *MockDatabase) Get(key string) ([]byte, error) {
	args := m.Called(key)
	if value, ok := m.data[key]; ok {
		return value, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDatabase) GetByPrefix(prefix string) (map[string][]byte, error) {
	args := m.Called(prefix)
	result := make(map[string][]byte)
	for k, v := range m.data {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			result[k] = v
		}
	}
	return result, args.Error(1)
}

func (m *MockDatabase) Delete(key string) error {
	m.Called(key)
	delete(m.data, key)
	return nil
}

// Add the missing DeleteByPrefix method
func (m *MockDatabase) DeleteByPrefix(prefix string) error {
	m.Called(prefix)
	for k := range m.data {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(m.data, k)
		}
	}
	return nil
}

func (m *MockDatabase) Close() error {
	return nil
}

func TestUploadHandler(t *testing.T) {
	c := config.GetDefaultConfig()
	conf := &config.Config{
		ServerPort: ":5005",
		Logger:     c.Logger,
	}
	mockDB := &MockDatabase{data: make(map[string][]byte)}
	server := NewServer(conf, mockDB)

	files := [][]byte{[]byte("file1"), []byte("file2")}
	filesJSON, _ := json.Marshal(files)

	req := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewBuffer(filesJSON))
	w := httptest.NewRecorder()

	mockDB.On("Put", mock.Anything, mock.Anything).Return(nil)

	server.UploadHandler(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, "/upload/root", bytes.NewBuffer(filesJSON))
	w = httptest.NewRecorder()

	mockDB.On("Put", mock.Anything, mock.Anything).Return(nil)
	mockDB.On("DeleteByPrefix", fileKey+"root").Return(nil)
	mockDB.On("DeleteByPrefix", proofKey+"root").Return(nil)

	server.UploadHandler(w, req)

	resp = w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDownloadHandler(t *testing.T) {
	c := config.GetDefaultConfig()
	conf := &config.Config{
		ServerPort: ":5005",
		Logger:     c.Logger,
	}
	mockDB := &MockDatabase{data: make(map[string][]byte)}
	server := NewServer(conf, mockDB)

	root := "root"
	hash := "somehash"
	file := []byte("file1")
	proof := &mkt.Proof{}
	proofJSON, _ := json.Marshal(proof)

	mockDB.data[fileKey+root+hash] = file
	mockDB.data[proofKey+root+hash] = proofJSON
	mockDB.data[root+"0"] = []byte(hash)

	req := httptest.NewRequest(http.MethodGet, "/download/"+root+"/0", nil)
	w := httptest.NewRecorder()

	mockDB.On("Get", root+"0").Return([]byte(hash), nil)
	mockDB.On("Get", proofKey+root+hash).Return(proofJSON, nil)
	mockDB.On("Get", fileKey+root+hash).Return(file, nil)

	server.DownloadHandler(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		File  []byte     `json:"file"`
		Proof *mkt.Proof `json:"proof"`
	}
	err := json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, file, result.File)
	assert.Equal(t, proof, result.Proof)

	req = httptest.NewRequest(http.MethodGet, "/download/", nil)
	w = httptest.NewRecorder()

	server.DownloadHandler(w, req)

	resp = w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

}

func TestUpdatedHandler(t *testing.T) {
	c := config.GetDefaultConfig()
	conf := &config.Config{
		ServerPort: ":5005",
		Logger:     c.Logger,
	}
	mockDB := &MockDatabase{data: make(map[string][]byte)}
	server := NewServer(conf, mockDB)

	root := "root"
	files := [][]byte{[]byte("file1"), []byte("file2")}
	filesJSON, _ := json.Marshal(files)

	req := httptest.NewRequest(http.MethodPost, "/update/"+root, bytes.NewBuffer(filesJSON))
	w := httptest.NewRecorder()

	mockDB.On("GetByPrefix", fileKey+root).Return(map[string][]byte{}, nil)
	mockDB.On("Put", mock.Anything, mock.Anything).Return(nil)
	mockDB.On("Delete", mock.Anything).Return(nil)
	mockDB.On("DeleteByPrefix", fileKey+root).Return(nil)
	mockDB.On("DeleteByPrefix", proofKey+root).Return(nil)

	server.UpdatedHandler(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		RootHash string `json:"root_hash"`
	}
	err := json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.NotEmpty(t, result.RootHash)
}

func TestRoutes(t *testing.T) {
	c := config.GetDefaultConfig()
	conf := &config.Config{
		ServerPort: ":5005",
		Logger:     c.Logger,
	}
	mockDB := &MockDatabase{data: make(map[string][]byte)}
	server := NewServer(conf, mockDB)
	server.routes()
}
