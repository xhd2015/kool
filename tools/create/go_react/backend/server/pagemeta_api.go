//go:build ignore

package server

import (
	"net/http"

	"__MODULE_NAME__/server/pagemeta"
)

// RegisterPageMetaAPI serves the server-owned card meta and the page document
// built from it. The meta travels with the card's rows: an agent that reads
// GET /api/pages/home learns what every card is for -- empty cards included --
// without opening a .tsx.
func RegisterPageMetaAPI(mux *http.ServeMux) {
	mux.HandleFunc("/api/page-meta", handlePageMeta)
	mux.HandleFunc("/api/pages/home", handleHomePage)
}

// pageSection is one card of a page document: its meta, whether it has rows,
// and the rows themselves.
type pageSection struct {
	Key   string        `json:"key"`
	Meta  pagemeta.Meta `json:"meta"`
	Empty bool          `json:"empty"`
	Items any           `json:"items,omitempty"`
}

// handlePageMeta reports the whole catalog, so a client can read every card's
// words without walking the pages.
func handlePageMeta(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	catalog, err := pagemeta.Catalog()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"sections": catalog})
}

// handleHomePage renders the Home page document: cards in web order, each with
// its meta and rows (or its empty state).
func handleHomePage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	last, err := CurrentCounter()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sections := make([]pageSection, 0, len(pagemeta.Sections()))
	for _, key := range pagemeta.Sections() {
		meta, err := pagemeta.Section(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		section := pageSection{Key: key, Meta: meta, Empty: true}
		switch key {
		case "counter":
			section.Items = map[string]any{"last": last}
			section.Empty = last == 0
		}
		sections = append(sections, section)
	}
	writeJSON(w, map[string]any{
		"page":     map[string]any{"title": "Home", "url": "http://" + r.Host + "/"},
		"sections": sections,
	})
}
