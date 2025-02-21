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
	Initialized bool
	FilePath    string
	Content     []byte
	Etag        string

	processes []func(content []byte) ([]byte, error)
}

func NewFileHandler(filePath string) (*FileHandler, error) {
	f := &FileHandler{
		FilePath: filePath,
	}
	err := f.Refresh()
	if err != nil {
		return nil, err
	}
	return f, nil
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
		Initialized: c.Initialized,
		FilePath:    c.FilePath,
		Content:     c.Content,
		Etag:        c.Etag,
		processes:   processes,
	}
}

func (c *FileHandler) Refresh() error {
	content, err := os.ReadFile(c.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.Initialized = false
			return nil
		}
		return fmt.Errorf("reading template file: %w", err)
	}

	// Calculate etag from content
	md5sum := md5.Sum(content)
	etag := hex.EncodeToString(md5sum[:])
	c.Initialized = true
	c.Etag = etag
	c.Content = content
	return nil
}

func (h *FileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.Initialized {
		// try refresh on demand
		err := h.Refresh()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !h.Initialized {
			http.Error(w, fmt.Sprintf("%s not found", h.FilePath), http.StatusNotFound)
			return
		}
	}

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
	_, err := w.Write(content)
	if err != nil {
		fmt.Fprintf(os.Stderr, "write content error: %v", err)
	}
}
