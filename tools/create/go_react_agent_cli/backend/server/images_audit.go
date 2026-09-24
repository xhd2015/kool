//go:build ignore

// Audit: what the library holds, whether each record's bytes are really a
// picture, and which documents reference it. A 200 response and an image/jpeg
// mime are not evidence of a picture, so every verdict here is derived from
// the bytes on each read rather than stored.
package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// List returns the whole library in id order, cross-referenced, and — when
// verify is set — with each record's bytes checked. This is the audit surface:
// a file that is not a picture, an unused image, or a reference to a missing
// image becomes visible here.
//
// verify=false skips reading the blobs, which is the only reason to turn it
// off: the reference walk still runs, so unused images stay visible.
func (s *imageStore) List(verify bool) ([]LibraryImage, error) {
	// The usage walk reads other files; it never runs under the image lock.
	usages, err := s.usages()
	if err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	entries, err := os.ReadDir(s.imagesDir())
	if os.IsNotExist(err) {
		return []LibraryImage{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := []LibraryImage{}
	for _, entry := range entries {
		if !entry.IsDir() || !isDecimalID(entry.Name()) {
			continue
		}
		meta, err := s.loadLocked(entry.Name())
		if err != nil {
			continue
		}
		out = append(out, s.libraryImage(meta, usages[meta.ID], verify))
	}
	sort.Slice(out, func(i, j int) bool {
		left, lerr := strconv.Atoi(out[i].ID)
		right, rerr := strconv.Atoi(out[j].ID)
		if lerr != nil || rerr != nil {
			return out[i].ID < out[j].ID
		}
		return left < right
	})
	return out, nil
}

// GetLibraryImage is List for one id, for `get /api/images/<id>`.
func (s *imageStore) GetLibraryImage(id string) (LibraryImage, error) {
	meta, err := s.Get(id)
	if err != nil {
		return LibraryImage{}, err
	}
	usages, err := s.Usages(id)
	if err != nil {
		return LibraryImage{}, err
	}
	return s.libraryImage(meta, usages, true), nil
}

// Usages returns where one image is referenced.
func (s *imageStore) Usages(id string) ([]ImageUsage, error) {
	usages, err := s.usages()
	if err != nil {
		return nil, err
	}
	if list := usages[strings.TrimSpace(id)]; list != nil {
		return list, nil
	}
	return []ImageUsage{}, nil
}

func (s *imageStore) libraryImage(meta ImageMeta, usages []ImageUsage, verify bool) LibraryImage {
	item := LibraryImage{
		ImageMeta: meta,
		URL:       s.imageURL(meta),
		Valid:     true,
		Usages:    usages,
	}
	if item.Usages == nil {
		item.Usages = []ImageUsage{}
	}
	data := s.blobBytes(meta)
	item.Bytes = int64(len(data))
	if strings.TrimSpace(meta.Ext) != "" {
		// The path is derived from the record, so it is reported even when the
		// byte check is skipped; a missing file shows up as missing-blob.
		item.Path = s.blobPath(meta.ID, meta.Ext)
	}
	if !verify {
		return item
	}
	problem, kind, external := s.imageProblem(meta, data)
	item.External = external
	if problem != "" {
		item.Valid = false
		item.Problem = problem
		item.ProblemKind = kind
	}
	return item
}

// imageProblem describes what is wrong with an image, in the same words the
// upload path uses.
//
// No stored extension means the record was created from a Source URL and never
// had a blob: that is by design, so it is external rather than broken. A
// 200 response and an image/jpeg mime are not evidence of a picture, so the
// verdict comes from the bytes.
func (s *imageStore) imageProblem(meta ImageMeta, data []byte) (problem, kind string, external bool) {
	if strings.TrimSpace(meta.Ext) == "" {
		if strings.TrimSpace(meta.Source) == "" {
			return "no local file and no source", problemNoSource, false
		}
		return "", "", true
	}
	if len(data) == 0 {
		return "image file is missing", problemMissingBlob, false
	}
	if _, _, ok := sniffImageMagic(data); ok {
		return "", "", false
	}
	if _, _, ok := sniffContentType(data); ok {
		return "", "", false
	}
	detected := strings.TrimSpace(strings.SplitN(http.DetectContentType(data), ";", 2)[0])
	return "not an image (detected " + detected + ")", problemNotAnImage, false
}

// usages maps image id to every reference found in the data tree.
//
// References are collected from the JSON documents that carry them by walking
// each decoded document for the image_id / image_ids keys. Matching by key
// name rather than by one typed loader per container keeps this correct when a
// new container starts carrying images, which is the failure mode a usage
// guard must not have.
func (s *imageStore) usages() (map[string][]ImageUsage, error) {
	out := map[string][]ImageUsage{}
	root := s.dir
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // an unreadable subtree must not fail the audit
		}
		if d.IsDir() {
			// images/<id>/meta.json describes an image, it does not use one.
			if path != root && d.Name() == "images" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}
		rel, rerr := filepath.Rel(root, path)
		if rerr != nil {
			return nil
		}
		container, registered := imageContainerFor(d.Name())
		if registered && container.retired {
			// Dead data: nothing renders it, so it must not refuse a delete.
			return nil
		}
		owner := ImageUsage{
			Kind: strings.TrimSuffix(d.Name(), ".json"),
			// A file no container claims cannot be rewritten by this build, so
			// the reference is reported as undetachable and blocks the delete.
			PagePath:   filepath.ToSlash(rel),
			Detachable: registered && container.detach != nil,
		}
		if registered {
			owner.Kind = container.kind
			owner.PagePath = container.pagePath
		}
		data, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil
		}
		var doc any
		if json.Unmarshal(data, &doc) != nil {
			return nil
		}
		for _, id := range collectImageRefs(doc, nil) {
			out[id] = appendUsage(out[id], owner)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// collectImageRefs walks a decoded document for image_id / image_ids.
func collectImageRefs(node any, out []string) []string {
	switch typed := node.(type) {
	case map[string]any:
		for key, value := range typed {
			switch key {
			case "image_id":
				if id := strings.TrimSpace(stringValue(value)); id != "" {
					out = append(out, id)
				}
			case "image_ids":
				for _, id := range stringList(value) {
					if id = strings.TrimSpace(id); id != "" {
						out = append(out, id)
					}
				}
			}
		}
		for _, value := range typed {
			out = collectImageRefs(value, out)
		}
	case []any:
		for _, value := range typed {
			out = collectImageRefs(value, out)
		}
	}
	return out
}

func stringValue(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func stringList(v any) []string {
	items, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
	}
	return out
}
