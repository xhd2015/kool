//go:build ignore

package server

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/file/idalloc"
)

// jpegBytes is a tiny JPEG: enough magic for the sniffer.
func jpegBytes(body string) []byte {
	return append([]byte{0xff, 0xd8, 0xff, 0xe0}, []byte(body)...)
}

// pngBytes is a tiny PNG: enough magic for the sniffer.
func pngBytes(body string) []byte {
	return append([]byte("\x89PNG\r\n\x1a\n"), []byte(body)...)
}

// xmlErrorPage is what a failed download stores: an error document. It is not
// a picture no matter what the file name says.
func xmlErrorPage() []byte {
	return []byte(`<?xml version='1.0' encoding='utf-8' ?><Error><Code>NoSuchKey</Code></Error>`)
}

// newTestLibrary returns a library in a fresh data dir.
func newTestLibrary(t *testing.T) *imageStore {
	t.Helper()
	return imageStoreFor(t.TempDir())
}

func mustCreate(t *testing.T, lib *imageStore, in ImageWrite) ImageMeta {
	t.Helper()
	meta, _, err := lib.Create(in)
	if err != nil {
		t.Fatalf("create image: %v", err)
	}
	return meta
}

func TestImageUploadStoresAndLists(t *testing.T) {
	lib := newTestLibrary(t)
	meta := mustCreate(t, lib, ImageWrite{
		Data:             jpegBytes("body"),
		Name:             "成都",
		OriginalFilename: "chengdu.jpg",
	})

	if !isDecimalID(meta.ID) {
		t.Fatalf("image id %q is not a decimal id from the shared sequence", meta.ID)
	}
	if meta.Ext != "jpg" || meta.Mime != "image/jpeg" {
		t.Fatalf("sniffed format = %q/%q, want jpg/image/jpeg", meta.Ext, meta.Mime)
	}
	blob := filepath.Join(lib.dir, "images", meta.ID, "image.jpg")
	if _, err := os.Stat(blob); err != nil {
		t.Fatalf("blob not written: %v", err)
	}
	if _, err := os.Stat(lib.metaPath(meta.ID)); err != nil {
		t.Fatalf("meta.json (the commit marker) not written: %v", err)
	}

	images, err := lib.List(true)
	if err != nil {
		t.Fatal(err)
	}
	if len(images) != 1 || !images[0].Valid {
		t.Fatalf("library = %+v, want one valid image", images)
	}
	// The audit reports the address an agent can open, not a 200.
	if images[0].Path != blob {
		t.Fatalf("audit path = %q, want %q", images[0].Path, blob)
	}
	if images[0].Bytes != int64(len(jpegBytes("body"))) {
		t.Fatalf("audit bytes = %d", images[0].Bytes)
	}
	if len(images[0].Usages) != 0 {
		t.Fatalf("a fresh image is unused, got usages %+v", images[0].Usages)
	}
}

func TestImageUploadDedupesIdenticalBytes(t *testing.T) {
	lib := newTestLibrary(t)
	first := mustCreate(t, lib, ImageWrite{Data: jpegBytes("same"), OriginalFilename: "a.jpg"})
	again, reused, err := lib.Create(ImageWrite{Data: jpegBytes("same"), OriginalFilename: "b.jpg"})
	if err != nil {
		t.Fatal(err)
	}
	if !reused || again.ID != first.ID {
		t.Fatalf("identical bytes must reuse the record: got id %q reused=%v, want %q", again.ID, reused, first.ID)
	}
	entries, err := os.ReadDir(lib.imagesDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("dedup wrote %d records, want 1", len(entries))
	}
}

func TestImageUploadRejectsNonImage(t *testing.T) {
	lib := newTestLibrary(t)
	_, _, err := lib.Create(ImageWrite{Data: xmlErrorPage(), OriginalFilename: "error.jpg"})
	if err == nil {
		t.Fatal("a file that is not a picture must be rejected")
	}
	// The message must name what the bytes are, so a wrong-format photo is
	// distinguishable from a downloaded error page.
	for _, want := range []string{"not an image", "text/xml", ".jpg"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("rejection %q missing %q", err, want)
		}
	}
	if entries, _ := os.ReadDir(lib.imagesDir()); len(entries) != 0 {
		t.Fatalf("a rejected upload must write nothing, found %d entries", len(entries))
	}
}

func TestImageAllowInvalidKeepsUnrecognizedBytes(t *testing.T) {
	lib := newTestLibrary(t)
	meta, _, err := lib.Create(ImageWrite{
		Data: xmlErrorPage(), OriginalFilename: "error.jpg", AllowInvalid: true,
	})
	if err != nil {
		t.Fatalf("the import escape hatch must keep the bytes: %v", err)
	}
	item, err := lib.GetLibraryImage(meta.ID)
	if err != nil {
		t.Fatal(err)
	}
	if item.Valid || item.ProblemKind != problemNotAnImage {
		t.Fatalf("escaped-in bytes must be reported, got valid=%v kind=%q", item.Valid, item.ProblemKind)
	}
}

func TestImageAuditReportsUsagesAndSkippedByteCheck(t *testing.T) {
	lib := newTestLibrary(t)
	meta := mustCreate(t, lib, ImageWrite{Data: jpegBytes("used"), OriginalFilename: "u.jpg"})
	if err := lib.SaveGallery(Gallery{ImageIDs: []string{meta.ID}}); err != nil {
		t.Fatal(err)
	}

	item, err := lib.GetLibraryImage(meta.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(item.Usages) != 1 || item.Usages[0].Kind != "gallery" || item.Usages[0].PagePath != "/api/gallery" {
		t.Fatalf("audit usages = %+v, want the gallery page", item.Usages)
	}
	if !item.Usages[0].Detachable {
		t.Fatal("a registered container with a detach must be detachable")
	}

	// A planted non-picture: the audit's verdict comes from the bytes, not
	// from the record existing.
	corrupt := mustCreate(t, lib, ImageWrite{
		Data: xmlErrorPage(), OriginalFilename: "bad.jpg", AllowInvalid: true,
	})
	images, err := lib.List(true)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, img := range images {
		if img.ID != corrupt.ID {
			continue
		}
		found = true
		if img.Valid || img.ProblemKind != problemNotAnImage || !strings.Contains(img.Problem, "text/xml") {
			t.Fatalf("corrupt image = valid:%v kind:%q problem:%q", img.Valid, img.ProblemKind, img.Problem)
		}
	}
	if !found {
		t.Fatal("the corrupt image is missing from the audit")
	}

	// verify=0 drops the byte verdict but keeps the references.
	unverified, err := lib.List(false)
	if err != nil {
		t.Fatal(err)
	}
	for _, img := range unverified {
		if !img.Valid || img.Problem != "" {
			t.Fatalf("verify=0 must not report a byte verdict, got %+v", img)
		}
		if img.ID == meta.ID && len(img.Usages) != 1 {
			t.Fatalf("verify=0 must still list references, got %+v", img.Usages)
		}
	}
}

func TestImageDeleteGuardRefusesReferencedImage(t *testing.T) {
	lib := newTestLibrary(t)
	meta := mustCreate(t, lib, ImageWrite{Data: jpegBytes("kept"), OriginalFilename: "k.jpg"})
	if err := lib.SaveGallery(Gallery{ImageIDs: []string{meta.ID}}); err != nil {
		t.Fatal(err)
	}

	detached, err := lib.Delete(meta.ID, false)
	if err == nil {
		t.Fatal("deleting an image a page shows must be refused")
	}
	if detached != 0 {
		t.Fatalf("a refused delete detaches nothing, got %d", detached)
	}
	// The refusal must name the page and the way out.
	for _, want := range []string{meta.ID, "/api/gallery", "--force"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("refusal %q missing %q", err, want)
		}
	}
	if _, err := os.Stat(lib.metaPath(meta.ID)); err != nil {
		t.Fatalf("the refused delete removed the image anyway: %v", err)
	}
}

func TestImageDeleteForceDetachesAndRemoves(t *testing.T) {
	lib := newTestLibrary(t)
	meta := mustCreate(t, lib, ImageWrite{Data: jpegBytes("gone"), OriginalFilename: "g.jpg"})
	other := mustCreate(t, lib, ImageWrite{Data: pngBytes("stays"), OriginalFilename: "s.png"})
	if err := lib.SaveGallery(Gallery{ImageIDs: []string{meta.ID, other.ID}}); err != nil {
		t.Fatal(err)
	}

	detached, err := lib.Delete(meta.ID, true)
	if err != nil {
		t.Fatalf("forced delete: %v", err)
	}
	if detached != 1 {
		t.Fatalf("forced delete detached %d documents, want 1", detached)
	}
	if _, err := os.Stat(lib.imageDir(meta.ID)); !os.IsNotExist(err) {
		t.Fatalf("the image directory survived a forced delete: %v", err)
	}
	// No dangling reference: the gallery no longer carries the deleted id.
	doc, err := lib.LoadGallery()
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.ImageIDs) != 1 || doc.ImageIDs[0] != other.ID {
		t.Fatalf("gallery after forced delete = %+v, want only %q", doc.ImageIDs, other.ID)
	}
}

func TestImageDeleteRefusesUnregisteredContainer(t *testing.T) {
	lib := newTestLibrary(t)
	meta := mustCreate(t, lib, ImageWrite{Data: jpegBytes("hidden"), OriginalFilename: "h.jpg"})
	// A file this build does not know about still references the image. The
	// delete must refuse rather than leave a dangling id behind, even --force.
	planted := `{"image_ids":["` + meta.ID + `"]}`
	if err := os.WriteFile(filepath.Join(lib.dir, "notes.json"), []byte(planted), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := lib.Delete(meta.ID, false); err == nil {
		t.Fatal("an unregistered reference must refuse the delete")
	}
	if _, err := lib.Delete(meta.ID, true); err == nil {
		t.Fatal("--force must not rewrite a container this build does not know")
	} else if !strings.Contains(err.Error(), "cannot rewrite") {
		t.Fatalf("refusal should say why: %v", err)
	}
	if _, err := os.Stat(lib.metaPath(meta.ID)); err != nil {
		t.Fatalf("the refused delete removed the image: %v", err)
	}
}

func TestImageAndCounterShareOneSequence(t *testing.T) {
	lib := newTestLibrary(t)
	meta := mustCreate(t, lib, ImageWrite{Data: jpegBytes("one"), OriginalFilename: "1.jpg"})
	// The same id.json the counter and every future entity draws from.
	next, err := idalloc.New(filepath.Join(lib.dir, "id.json")).Next(0)
	if err != nil {
		t.Fatal(err)
	}
	first, err := strconv.Atoi(meta.ID)
	if err != nil {
		t.Fatal(err)
	}
	if next != int64(first)+1 {
		t.Fatalf("next id after image %s = %d, want %d: one shared sequence", meta.ID, next, first+1)
	}
}

func TestImageFloorReseedSurvivesLostStateFile(t *testing.T) {
	lib := newTestLibrary(t)
	first := mustCreate(t, lib, ImageWrite{Data: jpegBytes("a"), OriginalFilename: "a.jpg"})
	second := mustCreate(t, lib, ImageWrite{Data: pngBytes("b"), OriginalFilename: "b.png"})
	// A lost state file must reseed from the ids already in use, never reuse.
	if err := os.Remove(filepath.Join(lib.dir, "id.json")); err != nil {
		t.Fatal(err)
	}
	third := mustCreate(t, lib, ImageWrite{Data: pngBytes("c"), OriginalFilename: "c.png"})
	for _, previous := range []string{first.ID, second.ID} {
		if third.ID == previous {
			t.Fatalf("id %s was reused after id.json was lost", third.ID)
		}
	}
	if third.ID <= second.ID {
		t.Fatalf("reseeded id %s must exceed the highest image id %s", third.ID, second.ID)
	}
}

// apiHarness points the data dir at a temp HOME and mounts the API.
func apiHarness(t *testing.T) (*http.ServeMux, string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	mux := http.NewServeMux()
	if err := RegisterAPI(mux); err != nil {
		t.Fatal(err)
	}
	dir, err := dataDir()
	if err != nil {
		t.Fatal(err)
	}
	return mux, dir
}

// uploadBody builds the multipart body of one upload.
func uploadBody(t *testing.T, data []byte, filename string, fields map[string]string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return &buf, writer.FormDataContentType()
}

func TestImagesAPIUploadReadDelete(t *testing.T) {
	mux, dir := apiHarness(t)

	body, contentType := uploadBody(t, jpegBytes("api"), "api.jpg", map[string]string{"name": "接口图"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/images", body)
	req.Header.Set("Content-Type", contentType)
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("upload status = %d: %s", rec.Code, rec.Body.String())
	}
	var uploaded PublicImage
	if err := json.Unmarshal(rec.Body.Bytes(), &uploaded); err != nil {
		t.Fatalf("parse upload: %v\n%s", err, rec.Body.String())
	}
	if uploaded.Path != filepath.Join(dir, "images", uploaded.ID, "image.jpg") {
		t.Fatalf("upload path = %q, want the stored blob under %s", uploaded.Path, dir)
	}
	if _, err := os.Stat(uploaded.Path); err != nil {
		t.Fatalf("the reported path does not exist: %v", err)
	}

	// The same bytes again: the same record, no new id.
	body, contentType = uploadBody(t, jpegBytes("api"), "again.jpg", nil)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/images", body)
	req.Header.Set("Content-Type", contentType)
	mux.ServeHTTP(rec, req)
	var reused PublicImage
	if err := json.Unmarshal(rec.Body.Bytes(), &reused); err != nil {
		t.Fatal(err)
	}
	if !reused.Reused || reused.ID != uploaded.ID {
		t.Fatalf("re-upload = id %q reused %v, want id %q reused true", reused.ID, reused.Reused, uploaded.ID)
	}

	// One record, with the path an agent can open.
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/images/"+uploaded.ID, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("get image status = %d: %s", rec.Code, rec.Body.String())
	}
	var item LibraryImage
	if err := json.Unmarshal(rec.Body.Bytes(), &item); err != nil {
		t.Fatal(err)
	}
	if !item.Valid || item.Path == "" {
		t.Fatalf("get image = %+v, want valid with a path", item)
	}

	// The bytes are served at the record's URL.
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, uploaded.URL, nil))
	if rec.Code != http.StatusOK || !strings.HasPrefix(rec.Header().Get("Content-Type"), "image/jpeg") {
		t.Fatalf("serve bytes = %d %q", rec.Code, rec.Header().Get("Content-Type"))
	}
	if got := rec.Header().Get("Cache-Control"); !strings.Contains(got, "immutable") {
		t.Fatalf("an id's bytes never change, so the cache header must be immutable, got %q", got)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/images/"+uploaded.ID, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("delete unreferenced image = %d: %s", rec.Code, rec.Body.String())
	}
}

func TestImagesAPIRejectsNonImageAndBadGalleryReference(t *testing.T) {
	mux, _ := apiHarness(t)

	body, contentType := uploadBody(t, xmlErrorPage(), "error.jpg", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/images", body)
	req.Header.Set("Content-Type", contentType)
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("a file that is not a picture must be a 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "text/xml") {
		t.Fatalf("the rejection must say what the bytes are: %s", rec.Body.String())
	}

	// A reference to an image that does not exist fails at author time.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/gallery", strings.NewReader(`{"image_ids":["999999"]}`))
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "unknown image") {
		t.Fatalf("gallery write = %d %s, want 400 unknown image", rec.Code, rec.Body.String())
	}
}

func TestIDsAPIHandsOutOneSequence(t *testing.T) {
	mux, _ := apiHarness(t)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/ids", strings.NewReader(`{"count":2}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("allocate ids = %d: %s", rec.Code, rec.Body.String())
	}
	var allocated struct {
		IDs []string `json:"ids"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &allocated); err != nil {
		t.Fatal(err)
	}
	if len(allocated.IDs) != 2 {
		t.Fatalf("allocated %d ids, want 2", len(allocated.IDs))
	}
	first, err := strconv.Atoi(allocated.IDs[0])
	if err != nil {
		t.Fatal(err)
	}
	second, err := strconv.Atoi(allocated.IDs[1])
	if err != nil {
		t.Fatal(err)
	}
	if second != first+1 {
		t.Fatalf("ids %v are not one sequence", allocated.IDs)
	}

	// The count is capped, so one call cannot drain the sequence.
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/ids", strings.NewReader(`{"count":100000}`)))
	var capped struct {
		IDs []string `json:"ids"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &capped); err != nil {
		t.Fatal(err)
	}
	if len(capped.IDs) != maxAllocatedIDs {
		t.Fatalf("allocation cap = %d ids, want %d", len(capped.IDs), maxAllocatedIDs)
	}
}
