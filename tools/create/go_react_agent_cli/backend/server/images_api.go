//go:build ignore

package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/file/idalloc"
)

// maxAllocatedIDs caps one POST /api/ids request: ids are handed out per row
// the web adds, not in bulk by a script importing a dataset.
const maxAllocatedIDs = 100

// RegisterImagesAPI mounts the unified image library and the shared id
// sequence:
//
//	POST   /api/ids                            allocate ids from one sequence
//	GET    /api/ids                            the last allocated id
//	POST   /api/images                         upload (multipart: file, name, caption)
//	GET    /api/images                         audit the library (?verify=0 skips byte reads)
//	GET    /api/images/<id>                    one record, with path and usages
//	DELETE /api/images/<id>[?force=1]          guarded delete; force detaches first
//	GET/PUT /api/gallery                       the demo container that references images
//	GET    /api/data/images/<id>/image.<ext>   the stored bytes
//
// Every id it hands out comes from <dataDir>/id.json, the same sequence the
// rest of the app uses, and the bytes live under <dataDir>/images.
func RegisterImagesAPI(mux *http.ServeMux) {
	mux.HandleFunc("/api/ids", handleIDs)
	mux.HandleFunc("/api/images", handleImages)
	mux.HandleFunc("/api/images/", handleImageItem)
	mux.HandleFunc("/api/gallery", handleGallery)
	mux.Handle("/api/data/", http.StripPrefix("/api/data/", http.HandlerFunc(serveDataFile)))
}

// handleIDs serves the shared id sequence: POST {"count":n} -> {"ids":[...]},
// GET -> {"last":N}. A page asks for an id before it renders a new row,
// because the id is what that row's references point at.
func handleIDs(w http.ResponseWriter, r *http.Request) {
	dir, err := dataDir()
	if err != nil {
		writeStoreError(w, err)
		return
	}
	alloc := idalloc.New(filepath.Join(dir, "id.json"))
	switch r.Method {
	case http.MethodGet:
		last, err := alloc.Last()
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, map[string]int64{"last": last})
	case http.MethodPost:
		store := imageStoreFor(dir)
		// The floor is the highest image id on disk: images live outside the
		// documents, so a lost id.json must still not re-issue one.
		floor := store.maxImageID()
		count := requestedIDCount(r)
		ids := make([]string, 0, count)
		for i := 0; i < count; i++ {
			n, err := alloc.Next(floor)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			ids = append(ids, strconv.FormatInt(n, 10))
		}
		writeJSON(w, map[string]any{"ids": ids})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// requestedIDCount reads how many ids to allocate. An empty body, a missing
// count and a non-positive count all mean one; the cap keeps one call from
// draining the sequence into a client that never uses it.
func requestedIDCount(r *http.Request) int {
	count := 0
	if raw := strings.TrimSpace(r.URL.Query().Get("count")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			count = parsed
		}
	}
	if count == 0 && r.Body != nil {
		var body struct {
			Count int `json:"count"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil && body.Count > 0 {
			count = body.Count
		}
	}
	if count < 1 {
		count = 1
	}
	if count > maxAllocatedIDs {
		count = maxAllocatedIDs
	}
	return count
}

// handleImages serves the library collection: GET audits it, POST uploads.
func handleImages(w http.ResponseWriter, r *http.Request) {
	store, err := defaultImageStore()
	if err != nil {
		writeStoreError(w, err)
		return
	}
	switch r.Method {
	case http.MethodGet:
		// No id addresses the library itself: the whole list, cross
		// referenced, with each file's bytes checked unless verify=0.
		verify := r.URL.Query().Get("verify") != "0"
		images, err := store.List(verify)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, images)
	case http.MethodPost:
		meta, reused, err := createUploadedImage(r, store)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		pub := store.publicImage(meta)
		pub.Reused = reused
		if reused {
			// Identical bytes were already stored: the same record, so no new
			// id was allocated and nothing was written.
			writeJSON(w, pub)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(pub)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// createUploadedImage reads a multipart upload. The bytes are capped before
// they reach the store, so an oversized upload fails at the door.
func createUploadedImage(r *http.Request, store *imageStore) (ImageMeta, bool, error) {
	if err := r.ParseMultipartForm(maxImageBytes + (1 << 20)); err != nil {
		return ImageMeta{}, false, &validationError{"invalid multipart form"}
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		return ImageMeta{}, false, &validationError{"file required"}
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxImageBytes+1))
	if err != nil {
		return ImageMeta{}, false, &validationError{"failed to read file"}
	}
	if len(data) > maxImageBytes {
		return ImageMeta{}, false, &validationError{fmt.Sprintf("image exceeds %d bytes", maxImageBytes)}
	}
	filename := "image"
	if header != nil && strings.TrimSpace(header.Filename) != "" {
		filename = header.Filename
	}
	return store.Create(ImageWrite{
		Data:             data,
		Name:             r.FormValue("name"),
		Source:           r.FormValue("source"),
		Caption:          r.FormValue("caption"),
		OriginalFilename: filename,
	})
}

// handleImageItem serves one library image: /api/images/<id>.
func handleImageItem(w http.ResponseWriter, r *http.Request) {
	store, err := defaultImageStore()
	if err != nil {
		writeStoreError(w, err)
		return
	}
	id := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/api/images/"))
	if id == "" || strings.Contains(id, "/") {
		writeStoreError(w, &validationError{"invalid image id"})
		return
	}
	switch r.Method {
	case http.MethodGet:
		item, err := store.GetLibraryImage(id)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, item)
	case http.MethodDelete:
		force := r.URL.Query().Get("force") == "1"
		detached, err := store.Delete(id, force)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, map[string]any{"deleted": id, "detached": detached})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGallery serves the demo container. A write is validated against the
// library, so a reference to an image that does not exist fails at author time
// instead of showing a hole later.
func handleGallery(w http.ResponseWriter, r *http.Request) {
	store, err := defaultImageStore()
	if err != nil {
		writeStoreError(w, err)
		return
	}
	switch r.Method {
	case http.MethodGet:
		doc, err := store.LoadGallery()
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, doc)
	case http.MethodPut:
		var doc Gallery
		if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
			writeStoreError(w, &validationError{"invalid gallery body"})
			return
		}
		for _, id := range doc.ImageIDs {
			if _, err := store.Get(id); err != nil {
				writeStoreError(w, &validationError{fmt.Sprintf("unknown image: %s", strings.TrimSpace(id))})
				return
			}
		}
		if err := store.SaveGallery(doc); err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, doc)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// serveDataFile serves the data dir, which is how an image's bytes are fetched:
// /api/data/images/<id>/image.<ext>.
func serveDataFile(w http.ResponseWriter, r *http.Request) {
	dir, err := dataDir()
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if isImageBlobPath(r.URL.Path) {
		// Identical bytes reuse a record, so an id's bytes never change and
		// its address can be cached forever.
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		w.Header().Set("X-Content-Type-Options", "nosniff")
	}
	http.FileServer(http.Dir(dir)).ServeHTTP(w, r)
}

// isImageBlobPath reports whether a data path addresses a stored image blob.
func isImageBlobPath(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	return len(parts) == 3 && parts[0] == "images" && isDecimalID(parts[1]) &&
		strings.HasPrefix(parts[2], "image.")
}

// writeStoreError maps a store error to a status: a validation error is the
// caller's mistake (400), anything else is ours (500). The body carries the
// sentence so a CLI can show it instead of a status code.
func writeStoreError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	var invalid *validationError
	if errors.As(err, &invalid) {
		status = http.StatusBadRequest
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
