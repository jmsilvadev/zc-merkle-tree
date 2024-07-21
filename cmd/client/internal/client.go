package client

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jmsilvadev/zc/pkg/mkt"
)

type Client struct {
	serverURL string
}

// NewClient creates a new client with the given server URL
func NewClient(serverURL string) *Client {
	return &Client{
		serverURL: serverURL,
	}
}

// UploadFiles uploads a list of files to the server and returns the server response
// TODO: create a streaming to transfer faster, but the text says
// that the files are small so maybe dont do it now
func (c *Client) UploadFiles(files [][]byte) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("invalid files")
	}

	data, err := json.Marshal(files)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(c.serverURL+"/upload", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode > 300 {
		return "", fmt.Errorf(string(body))
	}

	return string(body), nil
}

// UpdateFiles uploads a list of files to the server and includes
// in the existent list offiles in the server
// TODO: create a streaming to transfer faster, but the text says
// that the files are small so maybe dont do it now
func (c *Client) UpdateFiles(files [][]byte, configDir string) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("invalid files")
	}

	data, err := json.Marshal(files)
	if err != nil {
		return "", err
	}

	rootHash, err := c.GetLocalRootHash(configDir)
	if err != nil {
		return "", fmt.Errorf("error fetching the rootHash: %s", err)
	}

	resp, err := http.Post(fmt.Sprintf("%s/update/%s", c.serverURL, rootHash), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode > 300 {
		return "", fmt.Errorf(string(body))
	}

	// TODO: create an entity
	var result struct {
		RootHash string `json:"root_hash"`
	}

	err = json.NewDecoder(bytes.NewReader(body)).Decode(&result)
	if err != nil {
		return "", err
	}

	return result.RootHash, nil
}

// DownloadFile downloads a file from the server by its index and returns the file and its proof
func (c *Client) DownloadFile(index int, rootHash string) ([]byte, *mkt.Proof, error) {
	resp, err := http.Get(fmt.Sprintf("%s/download/%s/%d", c.serverURL, rootHash, index))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	// TODO: put this struct as entity
	var result struct {
		File  []byte     `json:"file"`
		Proof *mkt.Proof `json:"proof"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode > 300 {
		return nil, nil, fmt.Errorf(string(body))
	}

	err = json.NewDecoder(bytes.NewReader(body)).Decode(&result)
	if err != nil {
		return nil, nil, err
	}

	return result.File, result.Proof, nil
}

// GetRootHash calculates the root hash of a list of files using a Merkle tree
func (c *Client) GetRootHash(files [][]byte) string {
	hashes := make([]string, len(files))
	for i, v := range files {
		hash := sha256.Sum256(v)
		hashStr := hex.EncodeToString(hash[:])
		hashes[i] = hashStr
	}

	m := mkt.NewMerkleTree(hashes)
	return m.Root.Hash
}

func (c *Client) GetLocalRootHash(configDir string) (string, error) {
	// TODO: put this filename as a config env
	rootHashPath := filepath.Join(configDir, ".rootHash")
	rootHash, err := os.ReadFile(rootHashPath)
	if err != nil {
		return "", err
	}
	return string(rootHash), nil
}

// VerifyProof verifies the proof of a file against the given root hash
func (c *Client) VerifyProof(file []byte, proof *mkt.Proof, rootHash string) bool {
	hash := sha256.Sum256(file)
	hashStr := hex.EncodeToString(hash[:])

	return mkt.VerifyProof(hashStr, rootHash, proof)
}
