package route

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type FileHandler struct {
	FilePath  string
	Content   []byte
	Etag      string
	processes []func(content []byte) ([]byte, error)
}

func NewFileHandler(filePath string) (*FileHandler, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading template file: %w", err)
	}

	// Calculate etag from content
	md5sum := md5.Sum(content)
	etag := hex.EncodeToString(md5sum[:])

	return &FileHandler{
		FilePath: filePath,
		Content:  content,
		Etag:     etag,
	}, nil
}

func (c *FileHandler) AddProcess(process func(content []byte) ([]byte, error)) *FileHandler {
	c.processes = append(c.processes, process)
	return c
}

func (c *FileHandler) Clone() *FileHandler {
	var processes []func(content []byte) ([]byte, error)
	if c.processes != nil {
		processes = make([]func(content []byte) ([]byte, error), len(c.processes))
		copy(processes, c.processes)
	}
	return &FileHandler{
		Content:   c.Content,
		Etag:      c.Etag,
		processes: processes,
	}
}

func (h *FileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle cache properly
	w.Header().Set("Cache-Control", "no-cache, must-revalidate")

	// Check client etag
	clientEtag := r.Header.Get("If-None-Match")
	if clientEtag != "" && clientEtag == h.Etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// Set response cache header
	w.Header().Set("ETag", h.Etag)

	// set js
	contentType := "text/plain"
	if strings.HasSuffix(h.FilePath, ".js") {
		contentType = "text/javascript"
	} else if strings.HasSuffix(h.FilePath, ".css") {
		contentType = "text/css"
	} else if strings.HasSuffix(h.FilePath, ".html") {
		contentType = "text/html"
	}
	w.Header().Set("Content-Type", contentType+"; charset=utf-8")

	// Replace placeholders
	content := h.Content
	for _, process := range h.processes {
		var err error
		content, err = process(content)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.Write(content)
}
