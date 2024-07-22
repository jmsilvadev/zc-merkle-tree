package client

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmsilvadev/zc/pkg/mkt"
	"github.com/stretchr/testify/assert"
)

func TestUploadFiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/upload", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var files [][]byte
		err := json.NewDecoder(r.Body).Decode(&files)
		assert.NoError(t, err)
		assert.NotEmpty(t, files)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("upload successful"))
	}))
	defer server.Close()

	client := NewClient(server.URL)

	files := [][]byte{[]byte("file1"), []byte("file2")}
	resp, err := client.UploadFiles(files)
	assert.NoError(t, err)
	assert.Equal(t, "upload successful", resp)
}

func TestDownloadFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		files := [][]byte{[]byte("file1"), []byte("file2")}
		hash1 := sha256.Sum256(files[0])
		hash2 := sha256.Sum256(files[1])
		hashes := []string{hex.EncodeToString(hash1[:]), hex.EncodeToString(hash2[:])}

		m := mkt.NewMerkleTree(hashes)
		expectedRootHash := m.Root.Hash

		assert.Equal(t, "/download/"+expectedRootHash+"/0", r.URL.Path)

		response := struct {
			File  []byte     `json:"file"`
			Proof *mkt.Proof `json:"proof"`
		}{
			File:  []byte("file1"),
			Proof: &mkt.Proof{},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	files := [][]byte{[]byte("file1"), []byte("file2")}
	hash1 := sha256.Sum256(files[0])
	hash2 := sha256.Sum256(files[1])
	hashes := []string{hex.EncodeToString(hash1[:]), hex.EncodeToString(hash2[:])}

	m := mkt.NewMerkleTree(hashes)
	expectedRootHash := m.Root.Hash
	file, proof, err := client.DownloadFile(0, expectedRootHash)
	assert.NoError(t, err)
	assert.Equal(t, []byte("file1"), file)
	assert.NotNil(t, proof)
}

func TestUpdateFiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/update/newRootHash", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var files [][]byte
		err := json.NewDecoder(r.Body).Decode(&files)
		assert.NoError(t, err)
		assert.NotEmpty(t, files)

		response := struct {
			RootHash string `json:"root_hash"`
		}{
			RootHash: "newRootHash",
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	tempDir := t.TempDir()

	rootHashPath := filepath.Join(tempDir, ".rootHash")
	err := os.WriteFile(rootHashPath, []byte("newRootHash"), 0644)
	assert.NoError(t, err)

	files := [][]byte{[]byte("file1"), []byte("file2")}
	newRootHash, err := client.UpdateFiles(files, tempDir)
	assert.NoError(t, err)
	assert.Equal(t, "newRootHash", newRootHash)
}

func TestVerifyProof(t *testing.T) {

	file := []byte("file1")
	hash1 := sha256.Sum256(file)
	hashStr := hex.EncodeToString(hash1[:])
	m := mkt.NewMerkleTree([]string{hashStr})
	proof, _ := m.GetProof(hashStr)
	hash := mkt.GetProofHash(hashStr, proof)

	client := NewClient("")
	client.VerifyProof(file, proof, hash)
}

func TestGetRootHash(t *testing.T) {
	client := NewClient("")

	files := [][]byte{[]byte("file1"), []byte("file2")}
	hash1 := sha256.Sum256(files[0])
	hash2 := sha256.Sum256(files[1])
	hashes := []string{hex.EncodeToString(hash1[:]), hex.EncodeToString(hash2[:])}

	m := mkt.NewMerkleTree(hashes)
	expectedRootHash := m.Root.Hash

	rootHash := client.GetRootHash(files)
	assert.Equal(t, expectedRootHash, rootHash)
}

func TestGetLocalRootHash(t *testing.T) {
	client := NewClient("")
	_, err := client.GetLocalRootHash(getDefaultConfigDir())
	assert.NoError(t, err)
}

func getDefaultConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		os.Exit(1)
	}
	// TODO: put this dirname as a config env
	return filepath.Join(homeDir, ".zc")
}
