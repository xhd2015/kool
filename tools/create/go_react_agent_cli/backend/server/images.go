//go:build ignore

package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/file/idalloc"
)

// maxImageBytes caps one stored image. The HTTP handler reads at most this
// much, so an oversized upload fails at the door instead of filling the disk.
const maxImageBytes = 8 << 20

// allowedImageTypes maps a sniffed media type to the stored file extension.
// SVG is deliberately absent: an uploaded SVG served from the app origin can
// execute script.
var allowedImageTypes = map[string]string{
	"image/png":  "png",
	"image/jpeg": "jpg",
	"image/gif":  "gif",
	"image/webp": "webp",
}

// Problem kinds: the machine-readable reason an image is not usable, so a
// caller labels a row from data instead of matching prose.
const (
	problemNotAnImage    = "not-an-image"
	problemMissingBlob   = "missing-blob"
	problemMissingSource = "missing-source"
	problemNoSource      = "no-source"
)

// ImageMeta is the persisted identity of one library image. The bytes live
// beside it as image.<ext>; meta.json is written last and is what makes the
// record complete. There is no images.json catalog: the directory listing is
// the index, and a catalog would be a second source of truth that drifts.
type ImageMeta struct {
	ID               string `json:"id"`
	MD5              string `json:"md5,omitempty"`
	Name             string `json:"name,omitempty"`
	Source           string `json:"source,omitempty"`
	OriginalFilename string `json:"original_filename,omitempty"`
	// Ext is the SNIFFED format. An empty Ext means the record is an external
	// reference that lives at Source and has no local bytes: a supported kind,
	// not a defect.
	Ext       string `json:"ext,omitempty"`
	Mime      string `json:"mime,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	Caption   string `json:"caption,omitempty"`
}

// PublicImage is ImageMeta plus the derived fields the API publishes. Path is
// the absolute path of the stored blob, so a caller never has to guess where
// the bytes are; it is empty for an external record. Reused reports that
// identical bytes were already stored, so the id is not new.
type PublicImage struct {
	ImageMeta
	URL    string `json:"url"`
	Path   string `json:"path,omitempty"`
	Size   int64  `json:"bytes"`
	Reused bool   `json:"reused,omitempty"`
}

// ImageWrite is the input for creating (or reusing) a library image.
type ImageWrite struct {
	Data             []byte
	Name             string
	Source           string
	OriginalFilename string
	Caption          string
	// AllowInvalid keeps bytes that are not a recognizable image. Uploads
	// leave this false so the library cannot take in an error page; an import
	// or migration sets it so one bad legacy file cannot abort the run. Such a
	// record is reported as not-an-image by the audit.
	AllowInvalid bool
}

// LibraryImage is one library image plus the two facts the audit exists to
// surface: whether its bytes are really a picture, and where it is used.
//
// Valid and Problem are derived on every read rather than stored, so a stale
// flag can never claim a corrupt file is fine. Usages is why deleting needs a
// guard: removing an image a page shows leaves that page with a broken
// picture.
type LibraryImage struct {
	ImageMeta
	URL     string `json:"url"`
	Path    string `json:"path,omitempty"`
	Bytes   int64  `json:"bytes"`
	Valid   bool   `json:"valid"`
	Problem string `json:"problem,omitempty"`
	// ProblemKind is the machine-readable half of Problem.
	ProblemKind string `json:"problem_kind,omitempty"`
	// External marks an image that lives at its Source URL instead of a local
	// blob.
	External bool         `json:"external,omitempty"`
	Usages   []ImageUsage `json:"usages"`
}

// ImageUsage names one place an image is referenced from, so a refusal can say
// exactly what to detach first.
type ImageUsage struct {
	Kind     string `json:"kind"`
	PagePath string `json:"page_path"`
	// Detachable is false when a reference is found in a file this build
	// cannot rewrite. The delete then refuses instead of leaving a dangling id
	// behind.
	Detachable bool `json:"detachable"`
}

// validationError is a caller mistake. The HTTP layer answers 400 for it and
// 500 for anything else, so a bad upload is never reported as a server fault.
type validationError struct{ msg string }

func (e *validationError) Error() string { return e.msg }

// imageStore is the unified image library under <dataDir>/images. One store
// per data dir: the process-wide instance is shared, so the per-directory lock
// is shared too.
type imageStore struct {
	dir string
	mu  sync.RWMutex
}

var (
	imageStoresMu sync.Mutex
	imageStores   = map[string]*imageStore{}
)

// imageStoreFor returns the shared library for one data dir.
func imageStoreFor(dir string) *imageStore {
	imageStoresMu.Lock()
	defer imageStoresMu.Unlock()
	if store, ok := imageStores[dir]; ok {
		return store
	}
	store := &imageStore{dir: dir}
	imageStores[dir] = store
	return store
}

// defaultImageStore returns the library under the user's data dir.
func defaultImageStore() (*imageStore, error) {
	dir, err := dataDir()
	if err != nil {
		return nil, err
	}
	return imageStoreFor(dir), nil
}

func (s *imageStore) imagesDir() string           { return filepath.Join(s.dir, "images") }
func (s *imageStore) imageDir(id string) string   { return filepath.Join(s.imagesDir(), id) }
func (s *imageStore) metaPath(id string) string   { return filepath.Join(s.imageDir(id), "meta.json") }
func (s *imageStore) blobPath(id, ext string) string {
	return filepath.Join(s.imageDir(id), "image."+ext)
}

// allocateID hands out the next id from the app-wide sequence, raised past
// every image id already on disk so a lost id.json cannot re-issue one.
func (s *imageStore) allocateID() (string, error) {
	n, err := idalloc.New(filepath.Join(s.dir, "id.json")).Next(s.maxImageID())
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(n, 10), nil
}

// maxImageID is the largest image id on disk. Images live outside any single
// document, so the allocator floor has to include them.
func (s *imageStore) maxImageID() int64 {
	entries, err := os.ReadDir(s.imagesDir())
	if err != nil {
		return 0
	}
	var max int64
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if n, err := strconv.ParseInt(entry.Name(), 10, 64); err == nil && n > max {
			max = n
		}
	}
	return max
}

// Create stores data as a library image. It reports whether an existing record
// was reused because the same bytes were already stored: identical bytes are
// one record, so an id's bytes never change and its URL is immutable.
func (s *imageStore) Create(in ImageWrite) (ImageMeta, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.createLocked(in)
}

func (s *imageStore) createLocked(in ImageWrite) (ImageMeta, bool, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Source = strings.TrimSpace(in.Source)
	in.OriginalFilename = strings.TrimSpace(in.OriginalFilename)
	in.Caption = strings.TrimSpace(in.Caption)
	if len(in.Data) == 0 {
		return ImageMeta{}, false, &validationError{"image bytes required"}
	}
	if len(in.Data) > maxImageBytes {
		return ImageMeta{}, false, &validationError{fmt.Sprintf("image exceeds %d bytes", maxImageBytes)}
	}
	sum := md5Hex(in.Data)
	if existing, ok, err := s.findByMD5Locked(sum); err != nil {
		return ImageMeta{}, false, err
	} else if ok {
		return existing, true, nil
	}
	ext, mime, err := sniffImageType(in.OriginalFilename, in.Data, in.AllowInvalid)
	if err != nil {
		return ImageMeta{}, false, err
	}
	id, err := s.allocateID()
	if err != nil {
		return ImageMeta{}, false, err
	}
	meta := ImageMeta{
		ID:               id,
		MD5:              sum,
		Name:             firstNonEmpty(in.Name, in.Caption, filenameStem(in.OriginalFilename), "image"),
		Source:           in.Source,
		OriginalFilename: firstNonEmpty(in.OriginalFilename, "image."+ext),
		Ext:              ext,
		Mime:             mime,
		CreatedAt:        time.Now().Format(time.RFC3339),
		Caption:          in.Caption,
	}
	dir := s.imageDir(id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return ImageMeta{}, false, err
	}
	if err := os.WriteFile(s.blobPath(id, ext), in.Data, 0644); err != nil {
		_ = os.RemoveAll(dir)
		return ImageMeta{}, false, err
	}
	// meta.json LAST: its presence is what makes the directory a complete
	// record, so a partial upload is ignored by every reader.
	if err := writeJSONFile(s.metaPath(id), meta); err != nil {
		_ = os.RemoveAll(dir)
		return ImageMeta{}, false, err
	}
	return meta, false, nil
}

// findByMD5Locked returns the record already holding these bytes, if any.
func (s *imageStore) findByMD5Locked(sum string) (ImageMeta, bool, error) {
	if sum == "" {
		return ImageMeta{}, false, nil
	}
	entries, err := os.ReadDir(s.imagesDir())
	if os.IsNotExist(err) {
		return ImageMeta{}, false, nil
	}
	if err != nil {
		return ImageMeta{}, false, err
	}
	for _, entry := range entries {
		if !entry.IsDir() || !isDecimalID(entry.Name()) {
			continue
		}
		meta, err := s.loadLocked(entry.Name())
		if err != nil {
			continue
		}
		if meta.MD5 == sum {
			return meta, true, nil
		}
	}
	return ImageMeta{}, false, nil
}

// Get returns one record by id.
func (s *imageStore) Get(id string) (ImageMeta, error) {
	id = strings.TrimSpace(id)
	if !isDecimalID(id) {
		return ImageMeta{}, &validationError{"invalid image id"}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.loadLocked(id)
}

func (s *imageStore) loadLocked(id string) (ImageMeta, error) {
	var meta ImageMeta
	path := s.metaPath(id)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return ImageMeta{}, &validationError{fmt.Sprintf("image %s not found", id)}
		}
		return ImageMeta{}, err
	}
	if err := readJSONFile(path, &meta); err != nil {
		return ImageMeta{}, err
	}
	meta.ID = id
	return meta, nil
}

// blobBytes reads the stored blob for a record, if there is one.
func (s *imageStore) blobBytes(meta ImageMeta) []byte {
	ext := strings.TrimSpace(meta.Ext)
	if ext == "" {
		return nil
	}
	data, err := os.ReadFile(s.blobPath(meta.ID, ext))
	if err != nil {
		return nil
	}
	return data
}

// imageURL is the address the bytes are served at, or the Source URL for an
// external record that has no local blob.
func (s *imageStore) imageURL(meta ImageMeta) string {
	id, ext := strings.TrimSpace(meta.ID), strings.TrimSpace(meta.Ext)
	if id != "" && ext != "" {
		return "/api/data/images/" + id + "/image." + ext
	}
	return strings.TrimSpace(meta.Source)
}

// publicImage is what an upload answers with: the record plus its address, the
// absolute path of its bytes, and their size.
func (s *imageStore) publicImage(meta ImageMeta) PublicImage {
	pub := PublicImage{ImageMeta: meta, URL: s.imageURL(meta)}
	if strings.TrimSpace(meta.Ext) != "" {
		pub.Path = s.blobPath(meta.ID, meta.Ext)
	}
	pub.Size = int64(len(s.blobBytes(meta)))
	return pub
}
