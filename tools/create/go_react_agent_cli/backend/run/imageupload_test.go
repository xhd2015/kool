//go:build ignore

package run

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderImagePutsThePathOnTheFirstLine(t *testing.T) {
	dir := t.TempDir()
	blob := filepath.Join(dir, "image.jpg")
	if err := os.WriteFile(blob, []byte("bytes"), 0644); err != nil {
		t.Fatal(err)
	}
	// One image is one JSON object; the collection is the array. Mixing the
	// two shapes is what this test exists to catch.
	payload := `{"id":"7","name":"成都","ext":"jpg","url":"/api/data/images/7/image.jpg",` +
		`"path":"` + blob + `","bytes":5,"valid":true,"usages":[{"kind":"gallery","page_path":"/api/gallery"}]}`
	var stdout, stderr bytes.Buffer
	if err := renderImage("7", []byte(payload), &stdout, &stderr); err != nil {
		t.Fatalf("renderImage: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if lines[0] != blob {
		t.Fatalf("first line = %q, want the image path %q:\n%s", lines[0], blob, stdout.String())
	}
	for _, want := range []string{"name: 成都", "format: jpg · 5 B", "state: ok", "used by: /api/gallery"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("output missing %q:\n%s", want, stdout.String())
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("a healthy image must not warn: %s", stderr.String())
	}
}

func TestRenderImageWarnsWithoutFailing(t *testing.T) {
	payload := `{"id":"9","ext":"jpg","url":"/api/data/images/9/image.jpg","bytes":4,` +
		`"valid":false,"problem":"not an image (detected text/xml)","usages":[]}`
	var stdout, stderr bytes.Buffer
	if err := renderImage("9", []byte(payload), &stdout, &stderr); err != nil {
		t.Fatalf("a file that is not a picture is data, not a failed read: %v", err)
	}
	if !strings.Contains(stdout.String(), "state: NOT AN IMAGE (not an image (detected text/xml))") {
		t.Fatalf("state line must carry the verdict:\n%s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "warning: image 9 is not an image") {
		t.Fatalf("stderr must warn a script:\n%s", stderr.String())
	}
}

func TestImageAddressNeverDoublesAnOrigin(t *testing.T) {
	// External record: no local blob, an absolute source URL. Joining the
	// server origin onto it would yield http://localhost:8080https://…
	external := imageLibraryItem{ID: "3", URL: "https://example.com/a.jpg"}
	if got := imageAddress(external); got != "https://example.com/a.jpg" {
		t.Fatalf("imageAddress = %q, want the absolute URL unchanged", got)
	}
	// A reported path wins while it exists, and a stale one falls back.
	dir := t.TempDir()
	blob := filepath.Join(dir, "image.png")
	if err := os.WriteFile(blob, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	local := imageLibraryItem{ID: "4", Path: blob, URL: "/api/data/images/4/image.png"}
	if got := imageAddress(local); got != blob {
		t.Fatalf("imageAddress = %q, want %q", got, blob)
	}
	local.Path = filepath.Join(dir, "gone.png")
	if got := imageAddress(local); got != "/api/data/images/4/image.png" {
		t.Fatalf("a stale path must fall back to the URL, got %q", got)
	}
}

func TestRenderImageLibrarySummarizesProblems(t *testing.T) {
	payload := `[
	  {"id":"1","ext":"jpg","bytes":2048,"valid":true,"usages":[{"kind":"gallery","page_path":"/api/gallery"}]},
	  {"id":"2","ext":"jpg","bytes":0,"valid":false,"problem":"not an image (detected text/xml)","usages":[]},
	  {"id":"3","ext":"","valid":true,"external":true,"url":"https://example.com/b.jpg","usages":[]}
	]`
	var stdout, stderr bytes.Buffer
	if err := renderImageLibrary([]byte(payload), &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"ID", "Used by", "State", "2.0 KB", "external", "ok",
		"NOT AN IMAGE (not an image (detected text/xml))", "3 images · 1 not an image"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("table missing %q:\n%s", want, stdout.String())
		}
	}
	if !strings.Contains(stderr.String(), "warning: 1 of 3 library images are not pictures") {
		t.Fatalf("summary warning missing:\n%s", stderr.String())
	}
	// An external record has no local bytes, so it is not "unused": it is a
	// supported kind.
	if strings.Contains(stdout.String(), "3 images · 1 not an image · 2 unused") {
		t.Fatalf("external records must not count as unused:\n%s", stdout.String())
	}
}

func TestHandleImageUploadRejectsWrongTarget(t *testing.T) {
	// --file is only meaningful for the library collection.
	err := handleImageUpload("POST", "http://localhost:8080/api/gallery", "/tmp/x.jpg", "", "", false, requestOptions{})
	if err == nil || !strings.Contains(err.Error(), "--file uploads to /api/images") {
		t.Fatalf("wrong target error = %v", err)
	}
	// A JSON body and a file are two different uploads.
	err = handleImageUpload("POST", "http://localhost:8080/api/images", "/tmp/x.jpg", "", "", true, requestOptions{})
	if err == nil || !strings.Contains(err.Error(), "drop the JSON body") {
		t.Fatalf("body+file error = %v", err)
	}
	// put is not the upload verb.
	err = handleImageUpload("PUT", "http://localhost:8080/api/images", "/tmp/x.jpg", "", "", false, requestOptions{})
	if err == nil || !strings.Contains(err.Error(), "with post") {
		t.Fatalf("wrong verb error = %v", err)
	}
	// A missing file fails before any request is sent.
	err = handleImageUpload("POST", "http://localhost:8080/api/images", "/tmp/does-not-exist.jpg", "", "", false, requestOptions{})
	if err == nil || !strings.Contains(err.Error(), "open image") {
		t.Fatalf("missing file error = %v", err)
	}
}

func TestImageLibraryPathRecognizesCollectionAndItem(t *testing.T) {
	cases := []struct {
		target     string
		collection bool
		id         string
		ok         bool
	}{
		{target: "/api/images", collection: true, ok: true},
		{target: "http://localhost:8080/api/images", collection: true, ok: true},
		{target: "/api/images/", collection: true, ok: true},
		{target: "/api/images/12", id: "12", ok: true},
		{target: "http://localhost:8080/api/images/12?verify=0", id: "12", ok: true},
		{target: "/api/images/12/image.jpg", ok: false},
		{target: "/api/gallery", ok: false},
		{target: "/api/data/images/12/image.jpg", ok: false},
	}
	for _, tc := range cases {
		collection, id, ok := imageLibraryPath(tc.target)
		if ok != tc.ok || collection != tc.collection || id != tc.id {
			t.Fatalf("imageLibraryPath(%q) = (%v, %q, %v), want (%v, %q, %v)",
				tc.target, collection, id, ok, tc.collection, tc.id, tc.ok)
		}
	}
}

func TestWithQueryAddsForceAndVerify(t *testing.T) {
	got, err := withQuery("http://localhost:8080/api/images/3", "force", "1")
	if err != nil {
		t.Fatal(err)
	}
	if got != "http://localhost:8080/api/images/3?force=1" {
		t.Fatalf("withQuery = %q", got)
	}
	got, err = withQuery("http://localhost:8080/api/images?verify=0", "force", "1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "verify=0") || !strings.Contains(got, "force=1") {
		t.Fatalf("withQuery must keep the existing query: %q", got)
	}
}

func TestErrorDetailPrefersTheServerSentence(t *testing.T) {
	got := errorDetail([]byte(`{"error":"image 7 is used by /api/gallery"}`))
	if got != "image 7 is used by /api/gallery" {
		t.Fatalf("errorDetail = %q, want the server sentence", got)
	}
	// A non-JSON body still yields something readable.
	if got := errorDetail([]byte("  <html>nope</html>  ")); got != "<html>nope</html>" {
		t.Fatalf("errorDetail fallback = %q", got)
	}
}
