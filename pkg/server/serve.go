package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/aifoundry-org/storage-manager/pkg/cache"
	downloadparser "github.com/aifoundry-org/storage-manager/pkg/download/parser"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// Server a server to listen for API requests
type Server struct {
	addr   string
	cache  cache.Cache
	logger *log.Logger
}

// New create a new server instance with the provided configuration
func New(addr string, cache cache.Cache, logger *log.Logger) *Server {
	if logger == nil {
		logger = log.New()
	}
	return &Server{
		addr:   addr,
		cache:  cache,
		logger: logger,
	}
}

// Start start the server, runs continually, returning only when stopped or an error occurs.
func (s *Server) Start() error {
	r := mux.NewRouter()

	// Check if provided URL source exists in the cache or not. URL is base64 encoded and part of the query.
	r.HandleFunc("/content/{urlencoded}", s.contentGetHandler).Methods("GET")
	// Delete the provided URL source from the cache, if it exists. If not, return 200 OK.
	r.HandleFunc("/content/{urlencoded}", s.contentDeleteHandler).Methods("DELETE")
	// Ensure that the provided content is in the cache. If not, download it and store it in the cache.
	// URL and possible credentials are in the body of the request.
	r.HandleFunc("/content/", s.contentPostHandler).Methods("POST")

	server := &http.Server{
		Addr:    s.addr,
		Handler: r,
	}

	// Start HTTPS server with TLS configuration
	s.logger.Infof("Starting server on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		s.logger.Fatalf("Failed to listen and serve: %v", err)
	}
	return nil
}

// contentGetHandler check if the content is available in the cache
func (s *Server) contentGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	urlencoded := vars["urlencoded"]
	s.logger.Debugf("GET /content/%s", urlencoded)
	u, err := base64.StdEncoding.DecodeString(urlencoded)
	if err != nil {
		log.Debugf("GET /content/%s %v", urlencoded, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.logger.Debugf("GET %s", u)
	key, err := s.cache.Resolve(string(u))
	if err != nil {
		log.Debugf("cache resolve %s %v", u, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if key == "" {
		log.Debugf("key not found %s", u)
		http.Error(w, fmt.Sprintf("content not found %s %s", urlencoded, u), http.StatusNotFound)
		return
	}
	log.Debugf("found %s", u)
	w.WriteHeader(http.StatusOK)
}

// contentDeleteHandler remove the selected content from the cache
func (s *Server) contentDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	urlencoded := vars["urlencoded"]
	s.logger.Debugf("DELETE /content/%s", urlencoded)
	u, err := base64.StdEncoding.DecodeString(urlencoded)
	if err != nil {
		s.logger.Debugf("DELETE /content/%s %v", urlencoded, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	key, err := s.cache.Resolve(string(u))
	if err != nil {
		s.logger.Debugf("cache resolve %s %v", u, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if key == "" {
		s.logger.Debugf("key not present %s", u)
		w.WriteHeader(http.StatusOK)
		return
	}
	if err := s.cache.Unname(string(u)); err != nil {
		s.logger.Debugf("cache unname %s %v", u, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// and now need to clean up any unreferenced content in the cache
	if err := s.cache.GC(); err != nil {
		s.logger.Debugf("cache GC %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.logger.Debugf("cache delete %s OK", u)
	w.WriteHeader(http.StatusOK)
}

// contentPostHandler ensure that the provided content is in the cache
func (s *Server) contentPostHandler(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("POST /content/")
	// read the body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Debugf("POST /content read body %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var content PostRequest
	if err := json.Unmarshal(body, &content); err != nil {
		s.logger.Debugf("POST /content json unmarshal %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// check if the content is in the cache
	exists, err := s.cache.Exists(content.URL)
	if err != nil {
		s.logger.Debugf("POST /content error checking if content %s exists %v", content.URL, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if exists {
		s.logger.Debugf("POST /content %s already exists", content.URL)
		w.WriteHeader(http.StatusOK)
		return
	}
	// it does not, so download it
	downloader, err := downloadparser.Parse(content.URL)
	if err != nil {
		s.logger.Debugf("POST /content error getting downloader for %s %v", content.URL, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	downloadReaders, err := downloader.Download()
	if err != nil {
		s.logger.Debugf("POST /content error getting readers for content %s %v", content.URL, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, downloadReader := range downloadReaders {
		defer downloadReader.Reader.Close()
		exists, err := s.cache.Exists(downloadReader.Key)
		if err != nil {
			s.logger.Debugf("POST /content error checking if key %s exists %v", downloadReader.Key, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if exists {
			s.logger.Debugf("POST /content key %s already exists", downloadReader.Key)
			continue
		}
		s.logger.Debugf("POST /content putting into cache key %s", downloadReader.Key)
		if err := s.cache.Put(downloadReader.Key, downloadReader.Size, downloadReader.Reader); err != nil {
			s.logger.Debugf("POST /content error putting into cache key %s %v", downloadReader.Key, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if err := s.cache.Name(downloadReaders[0].Key, content.URL); err != nil {
		s.logger.Debugf("POST /content error tagging root %s %v", content.URL, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.logger.Debugf("POST /content success %s", content.URL)
	w.WriteHeader(http.StatusOK)
}
