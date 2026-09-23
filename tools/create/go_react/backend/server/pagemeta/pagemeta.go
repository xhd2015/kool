//go:build ignore

// Package pagemeta holds the meta every __PROJECT_NAME__ card shows: the card
// title, the hint under it, and the sentence shown when the card has no rows.
//
// The parts under parts/ are the single source of truth. The server embeds them
// (//go:embed) and serves them with the page content, so an agent reading
// GET /api/pages/home learns what each card is for -- empty cards included --
// without opening a single .tsx. The React card imports the same JSON through
// the @pagemeta vite alias. Never retype a hint in TSX or in a Go string: add
// the word here and read it from both sides.
package pagemeta

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

//go:embed parts/*.json
var partsFS embed.FS

// Meta is one card's server-owned meta.
type Meta struct {
	Title string `json:"title"`
	Hint  string `json:"hint"`
	Empty string `json:"empty"`
}

// sectionDef registers one card: the key the page document uses and the part
// file holding its meta. Page documents are built from this registry only, so a
// card cannot ship without meta.
type sectionDef struct {
	Key  string
	Part string
}

var sections = []sectionDef{
	{Key: "counter", Part: "CounterCard"},
}

var (
	loadOnce sync.Once
	metas    map[string]Meta
	loadErr  error
)

// Sections returns the registered card keys in page order.
func Sections() []string {
	keys := make([]string, 0, len(sections))
	for _, s := range sections {
		keys = append(keys, s.Key)
	}
	return keys
}

// Section returns one card's meta. An unregistered key is an error, so a card
// never renders without its brief.
func Section(key string) (Meta, error) {
	all, err := Catalog()
	if err != nil {
		return Meta{}, err
	}
	meta, ok := all[key]
	if !ok {
		return Meta{}, fmt.Errorf("pagemeta: no such section %q", key)
	}
	return meta, nil
}

// Catalog returns every registered card's meta, keyed by section.
func Catalog() (map[string]Meta, error) {
	loadOnce.Do(load)
	if loadErr != nil {
		return nil, loadErr
	}
	out := make(map[string]Meta, len(metas))
	for key, meta := range metas {
		out[key] = meta
	}
	return out, nil
}

// Validate reports a registered card that is missing a visible word. The
// package test runs it, so a rename or a dropped hint fails the build instead of
// rendering a card without its brief.
func Validate() error {
	if _, err := Catalog(); err != nil {
		return err
	}
	for _, s := range sections {
		meta := metas[s.Key]
		if strings.TrimSpace(meta.Title) == "" {
			return fmt.Errorf("pagemeta: %s has no title in parts/%s.json", s.Key, s.Part)
		}
		if strings.TrimSpace(meta.Hint) == "" {
			return fmt.Errorf("pagemeta: %s has no hint in parts/%s.json", s.Key, s.Part)
		}
		if strings.TrimSpace(meta.Empty) == "" {
			return fmt.Errorf("pagemeta: %s has no empty state in parts/%s.json", s.Key, s.Part)
		}
	}
	return nil
}

func load() {
	metas = map[string]Meta{}
	for _, s := range sections {
		name := "parts/" + s.Part + ".json"
		raw, err := partsFS.ReadFile(name)
		if err != nil {
			loadErr = fmt.Errorf("pagemeta: read %s: %w", name, err)
			return
		}
		var meta Meta
		if err := json.Unmarshal(raw, &meta); err != nil {
			loadErr = fmt.Errorf("pagemeta: parse %s: %w", name, err)
			return
		}
		metas[s.Key] = meta
	}
}
