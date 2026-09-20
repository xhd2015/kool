//go:build ignore

package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/xhd2015/dot-pkgs/go-pkgs/file/idalloc"
)

// dataDir returns ~/.__PROJECT_NAME__/, the per-user directory for persistent
// state. Files are created with private permissions (dir 0700, files 0600).
func dataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".__PROJECT_NAME__"), nil
}

// RegisterCounterAPI demonstrates persistent server state with the shared
// dot-pkgs id allocator: ids come from ~/.__PROJECT_NAME__/id.json, survive
// restarts, and are serialized across processes via an flock on the inferred
// ~/.__PROJECT_NAME__/id.lock. Delete this call (and store.go) when you add
// your own storage.
func RegisterCounterAPI(mux *http.ServeMux) {
	mux.HandleFunc("/api/counter", handleCounter)
}

// handleCounter allocates the next sequential id on POST and reports the
// current counter on GET.
func handleCounter(w http.ResponseWriter, r *http.Request) {
	dir, err := dataDir()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	alloc := idalloc.New(filepath.Join(dir, "id.json"))
	switch r.Method {
	case http.MethodGet:
		last, err := alloc.Last()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]int64{"last": last})
	case http.MethodPost:
		next, err := alloc.Next(0)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]int64{"id": next})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
