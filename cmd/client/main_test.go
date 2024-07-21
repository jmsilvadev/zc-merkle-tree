package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunUpload(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "testfile.txt")
	err := os.WriteFile(tempFile, []byte("test content"), 0644)
	assert.NoError(t, err)

	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	args := []string{"-operation", "upload", "-dir", tempDir, "-host", "http://localhost:5000"}
	err = run(flagSet, args)
	assert.NoError(t, err)

	rootHashPath := filepath.Join(tempDir, ".rootHash")
	err = os.WriteFile(rootHashPath, []byte("6ae8a75555209fd6c44157c0aed8016e763ff435a19cf186f76863140143ff72"), 0644)
	assert.NoError(t, err)

	flagSet = flag.NewFlagSet("test", flag.ContinueOnError)
	args = []string{"-operation", "download", "-index", "0", "-config-dir", tempDir, "-host", "http://localhost:5000"}
	err = run(flagSet, args)
	assert.NoError(t, err)

	tempFile = filepath.Join(tempDir, "testfile.txt")
	err = os.WriteFile(tempFile, []byte("test content"), 0644)
	assert.NoError(t, err)

	flagSet = flag.NewFlagSet("test", flag.ContinueOnError)
	args = []string{"-operation", "upload", "-files", tempDir + "/testfile.txt", "-host", "http://localhost:5000"}
	err = run(flagSet, args)
	assert.NoError(t, err)

	tempFile = filepath.Join(tempDir, "testfile2.txt")
	err = os.WriteFile(tempFile, []byte("test content2"), 0644)
	assert.NoError(t, err)

	flagSet = flag.NewFlagSet("test", flag.ContinueOnError)
	args = []string{"-operation", "update", "-files", tempDir + "/testfile2.txt", "-host", "http://localhost:5000"}
	err = run(flagSet, args)
	assert.NoError(t, err)
}

func TestRunInvalidOperation(t *testing.T) {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	args := []string{"-operation", "invalid"}
	err := run(flagSet, args)
	assert.Error(t, err)
	assert.Equal(t, "invalid operation. Please specify 'upload' or 'update' or 'download' using the -operation parameter", err.Error())
}

func TestRunMissingIndex(t *testing.T) {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	args := []string{"-operation", "download", "-host", "http://localhost:5000"}
	err := run(flagSet, args)
	assert.Error(t, err)
	assert.Equal(t, "please provide the index and configDir parameters for the download operation", err.Error())
}

func TestRunMissingRootHash(t *testing.T) {
	tempDir := t.TempDir()
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	args := []string{"-operation", "download", "-index", "0", "-config-dir", tempDir, "-host", "http://localhost:5000"}
	err := run(flagSet, args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error fetching the rootHash")
}
