//go:build ignore

package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// galleryFile is the demo container: the images a page shows, in order.
const galleryFile = "gallery.json"

// Gallery is the demo container document. It exists to make the delete guard
// real: deleting an image the gallery lists is refused until the reference is
// detached. Replace it with your own domain container -- every file carrying an
// image_id / image_ids field belongs in imageContainers below.
type Gallery struct {
	ImageIDs []string `json:"image_ids"`
}

// imageContainer describes one file that can reference library images.
//
// The audit and the forced delete both drive off this table. They used to be
// two hand-written walks that disagreed: the guard detected references in one
// set of files while the detach rewrote another, so `delete --force` on a
// referenced image deleted the picture and left the id dangling -- the
// opposite of what its own refusal message promised. One table makes that
// asymmetry unrepresentable.
type imageContainer struct {
	// file is the base name that marks the container.
	file string
	// kind is the usage kind reported for a reference found here.
	kind string
	// pagePath is where a human sees the reference, for the refusal message.
	pagePath string
	// retired marks data the product no longer reads. It is skipped entirely:
	// refusing a delete because of a file nothing renders would be misleading.
	retired bool
	// detach drops every reference to one image. A container with no detach
	// cannot be half-supported: the delete refuses rather than leaving a
	// dangling id behind.
	detach func(s *imageStore, path, id string) (bool, error)
}

// imageContainers is the complete set of files that carry image references.
// Every type with an image_id / image_ids field must appear here. A JSON file
// that carries a reference but is not registered is reported as undetachable,
// so the delete refuses instead of leaving a dangling id behind.
var imageContainers = []imageContainer{
	{file: galleryFile, kind: "gallery", pagePath: "/api/gallery", detach: detachGalleryFile},
}

// imageContainerFor finds the container a file belongs to.
func imageContainerFor(name string) (imageContainer, bool) {
	for _, container := range imageContainers {
		if container.file == name {
			return container, true
		}
	}
	return imageContainer{}, false
}

// LoadGallery reads the demo container. A missing file is an empty gallery.
func (s *imageStore) LoadGallery() (Gallery, error) {
	var doc Gallery
	data, err := os.ReadFile(filepath.Join(s.dir, galleryFile))
	if err != nil {
		if os.IsNotExist(err) {
			return Gallery{ImageIDs: []string{}}, nil
		}
		return Gallery{}, err
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return Gallery{}, err
	}
	if doc.ImageIDs == nil {
		doc.ImageIDs = []string{}
	}
	return doc, nil
}

// SaveGallery replaces the demo container document.
func (s *imageStore) SaveGallery(doc Gallery) error {
	if doc.ImageIDs == nil {
		doc.ImageIDs = []string{}
	}
	return writeJSONFile(filepath.Join(s.dir, galleryFile), doc)
}

// detachGalleryFile drops one image from the gallery. The gallery itself
// survives: only its picture reference goes.
func detachGalleryFile(s *imageStore, path, id string) (bool, error) {
	doc, err := s.LoadGallery()
	if err != nil {
		return false, err
	}
	kept := make([]string, 0, len(doc.ImageIDs))
	for _, imageID := range doc.ImageIDs {
		if strings.TrimSpace(imageID) == id {
			continue
		}
		kept = append(kept, imageID)
	}
	if len(kept) == len(doc.ImageIDs) {
		return false, nil
	}
	doc.ImageIDs = kept
	if err := s.SaveGallery(doc); err != nil {
		return false, err
	}
	return true, nil
}
