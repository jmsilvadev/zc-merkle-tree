package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/jmsilvadev/zc/pkg/config"
	"github.com/jmsilvadev/zc/pkg/db"
	"github.com/jmsilvadev/zc/pkg/mkt"
)

const (
	fileKey  = "file_"
	proofKey = "proof_"

	errInternal   = "internal error, try again"
	errBadRequest = "invalid data sent"
	errNotFound   = "not found"
)

type Server struct {
	conf *config.Config
	db   db.Database
}

func NewServer(c *config.Config, db db.Database) *Server {
	return &Server{
		db:   db,
		conf: c,
	}
}

func (s *Server) Start() {
	defer s.db.Close()

	server := &http.Server{
		Addr:    s.conf.ServerPort,
		Handler: s.routes(),
	}

	listener := make(chan os.Signal, 1)
	signal.Notify(listener, os.Interrupt, syscall.SIGTERM)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		<-listener
		s.conf.Logger.Warn("Received shutdown signal")

		// TODO: put this timeout as a config env
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			s.conf.Logger.Error("Server forced to shutdown: " + err.Error())
		}

		wg.Done()
	}()

	s.conf.Logger.Info("Server is running on port " + s.conf.ServerPort)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		s.conf.Logger.Error("Server error: " + err.Error())
	}

	wg.Wait()
	s.conf.Logger.Warn("Server gracefully stopped")
}

// TODO: create a streaming to transfer faster, but the text says
// that the files are small so maybe dont do it now
func (s *Server) UploadHandler(w http.ResponseWriter, r *http.Request) {
	var files [][]byte
	err := json.NewDecoder(r.Body).Decode(&files)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) == 3 {
		root := pathParts[2]
		// Delete the oldFiles if they exist
		err = s.db.DeleteByPrefix(fileKey + root)
		if err != nil {
			s.conf.Logger.Error(err.Error())
			http.Error(w, errInternal, http.StatusInternalServerError)
			return
		}
		// delete the old proof
		err = s.db.DeleteByPrefix(proofKey + root)
		if err != nil {
			s.conf.Logger.Error(err.Error())
			http.Error(w, errInternal, http.StatusInternalServerError)
			return
		}
	}

	// TODO: This is not atomic, transform to atomic
	hashes := make([]string, len(files))
	for i, v := range files {
		hash := sha256.Sum256(v)
		hashStr := hex.EncodeToString(hash[:])
		hashes[i] = hashStr
	}

	m := mkt.NewMerkleTree(hashes)

	for i, h := range hashes {
		proof, err := m.GetProof(h)
		if err != nil {
			s.conf.Logger.Error(err.Error())
			http.Error(w, errInternal, http.StatusInternalServerError)
			return
		}
		proofByte, err := json.Marshal(proof)
		if err != nil {
			s.conf.Logger.Error(err.Error())
			http.Error(w, errInternal, http.StatusInternalServerError)
			return
		}

		err = s.db.Put(proofKey+m.Root.Hash+h, proofByte)
		if err != nil {
			s.conf.Logger.Error(err.Error())
			http.Error(w, errInternal, http.StatusInternalServerError)
			return
		}
		err = s.db.Put(m.Root.Hash+strconv.Itoa(i), []byte(h))
		if err != nil {
			s.conf.Logger.Error(err.Error())
			http.Error(w, errInternal, http.StatusInternalServerError)
			return
		}
		err = s.db.Put(fileKey+m.Root.Hash+h, files[i])
		if err != nil {
			s.conf.Logger.Error(err.Error())
			http.Error(w, errInternal, http.StatusInternalServerError)
			return
		}
	}

	// TODO: improve the responses with a helper
	w.WriteHeader(http.StatusOK)
}

func (s *Server) DownloadHandler(w http.ResponseWriter, r *http.Request) {
	// NOTE: /root/index
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, errBadRequest, http.StatusBadRequest)
		return
	}

	root := pathParts[2]
	indexStr := pathParts[3]

	i, err := strconv.Atoi(indexStr)
	if err != nil {
		s.conf.Logger.Error(err.Error())
		http.Error(w, errBadRequest, http.StatusBadRequest)
		return
	}

	hash, err := s.db.Get(root + strconv.Itoa(i))
	if err != nil {
		s.conf.Logger.Error(err.Error())
		http.Error(w, errBadRequest, http.StatusBadRequest)
		return
	}

	proof, err := s.db.Get(proofKey + root + string(hash))
	if err != nil {
		s.conf.Logger.Error(err.Error())
		http.Error(w, errBadRequest, http.StatusBadRequest)
		return
	}

	file, err := s.db.Get(fileKey + root + string(hash))
	if err != nil {
		s.conf.Logger.Error(err.Error())
		http.Error(w, errNotFound, http.StatusNotFound)
		return
	}

	var mktProof *mkt.Proof
	err = json.Unmarshal(proof, &mktProof)
	if err != nil {
		s.conf.Logger.Error(err.Error())
		http.Error(w, errInternal, http.StatusInternalServerError)
		return
	}

	// TODO: create an entity
	result := struct {
		File  []byte     `json:"file"`
		Proof *mkt.Proof `json:"proof"`
	}{
		File:  file,
		Proof: mktProof,
	}

	// TODO: improve the responses with a helper
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) UpdatedHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, errBadRequest, http.StatusBadRequest)
		return
	}

	root := pathParts[2]

	var files [][]byte
	err := json.NewDecoder(r.Body).Decode(&files)
	if err != nil {
		s.conf.Logger.Error(err.Error())
		http.Error(w, errBadRequest, http.StatusBadRequest)
		return
	}

	oldFiles, err := s.db.GetByPrefix(fileKey + root)
	if err != nil {
		s.conf.Logger.Error(err.Error())
		http.Error(w, errBadRequest, http.StatusBadRequest)
		return
	}

	for _, v := range files {
		hash := sha256.Sum256(v)
		hashStr := hex.EncodeToString(hash[:])
		oldFiles[hashStr] = v
	}

	i := 0
	hashes := make([]string, len(oldFiles))
	newFiles := make([][]byte, len(oldFiles))
	for _, v := range oldFiles {
		hash := sha256.Sum256(v)
		hashes[i] = hex.EncodeToString(hash[:])
		newFiles[i] = v
		i++
	}

	m := mkt.NewMerkleTree(hashes)

	for i, h := range hashes {
		proof, err := m.GetProof(h)
		if err != nil {
			s.conf.Logger.Error(err.Error())
			http.Error(w, errInternal, http.StatusInternalServerError)
			return
		}
		proofByte, err := json.Marshal(proof)
		if err != nil {
			s.conf.Logger.Error(err.Error())
			http.Error(w, errInternal, http.StatusInternalServerError)
			return
		}

		err = s.db.Put(proofKey+m.Root.Hash+h, proofByte)
		if err != nil {
			s.conf.Logger.Error(err.Error())
			http.Error(w, errInternal, http.StatusInternalServerError)
			return
		}
		err = s.db.Put(m.Root.Hash+strconv.Itoa(i), []byte(h))
		if err != nil {
			s.conf.Logger.Error(err.Error())
			http.Error(w, errInternal, http.StatusInternalServerError)
			return
		}
		err = s.db.Put(fileKey+m.Root.Hash+h, newFiles[i])
		if err != nil {
			s.conf.Logger.Error(err.Error())
			http.Error(w, errInternal, http.StatusInternalServerError)
			return
		}
	}

	// Delete the oldFiles
	err = s.db.DeleteByPrefix(fileKey + root)
	if err != nil {
		s.conf.Logger.Error(err.Error())
		http.Error(w, errInternal, http.StatusInternalServerError)
		return
	}
	// delete the old proof
	err = s.db.DeleteByPrefix(proofKey + root)
	if err != nil {
		s.conf.Logger.Error(err.Error())
		http.Error(w, errInternal, http.StatusInternalServerError)
		return
	}

	i = 0
	// Delete old index
	for range oldFiles {
		err = s.db.Delete(root + strconv.Itoa(i))
		if err != nil {
			s.conf.Logger.Error(err.Error())
			http.Error(w, errInternal, http.StatusInternalServerError)
			return
		}
		i++
	}

	// TODO: create an entity
	result := struct {
		RootHash string `json:"root_hash"`
	}{
		RootHash: m.Root.Hash,
	}

	// TODO: improve the responses with a helper
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) routes() *http.ServeMux {
	// TODO: create an OAS if I have time
	mux := http.NewServeMux()

	// uploads creates a new merkle tree
	mux.HandleFunc("/upload", s.UploadHandler)
	// uploads creates a new merkle tree but uses the existent one
	mux.HandleFunc("/update/", s.UpdatedHandler)
	mux.HandleFunc("/download/", s.DownloadHandler)
	return mux
}
