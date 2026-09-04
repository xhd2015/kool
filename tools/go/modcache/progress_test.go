package modcache

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestShouldHeartbeat(t *testing.T) {
	cases := []struct {
		i, n int
		want bool
	}{
		{0, 10, false},
		{1, 0, false},
		{1, 1, true},
		{1, 2, true},
		{2, 2, true},
		{2, 50, false},
		{25, 50, true},
		{26, 50, false},
		{50, 50, true},
		{51, 50, false},
	}
	for _, tc := range cases {
		if got := shouldHeartbeat(tc.i, tc.n); got != tc.want {
			t.Errorf("shouldHeartbeat(%d,%d)=%v want %v", tc.i, tc.n, got, tc.want)
		}
	}
}

func TestStageProgressMarkers(t *testing.T) {
	var buf bytes.Buffer
	p := newStageProgress(&buf, 3)
	p.start("extracted", "walking")
	p.line("extracted", "sizing 2 versions")
	p.detail("%d/%d  %s", 1, 2, "example.com/foo@v1.0.0")
	p.ok("extracted", "12B")
	p.start("download", "walking")
	p.ok("download", "0B")

	got := buf.String()
	wantLines := []string{
		"[1/3] extracted    walking",
		"[1/3] extracted    sizing 2 versions",
		"      1/2  example.com/foo@v1.0.0",
		"[1/3] extracted    ok  12B",
		"[2/3] download     walking",
		"[2/3] download     ok  0B",
	}
	for _, want := range wantLines {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	// Kind-aligned detail: indent equals len("[1/3] ").
	prefix := "[1/3] "
	indent := strings.Repeat(" ", len(prefix))
	if !strings.Contains(got, indent+"1/2  example.com/foo@v1.0.0\n") {
		t.Fatalf("detail indent want %q prefix; got:\n%s", indent, got)
	}
}

func TestInventoryCacheEmitsProgressDuringWalk(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "example.com", "foo@v1.0.0", "x.txt"), "old\n")
	writeFile(t, filepath.Join(root, "example.com", "foo@v1.2.0", "x.txt"), "new\n")

	var buf bytes.Buffer
	prog := newStageProgress(&buf, 3)
	if _, err := inventoryCache(root, prog); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "[1/3] extracted") {
		t.Fatalf("expected extracted stage; got:\n%s", got)
	}
	if !strings.Contains(got, "sizing 2 versions") {
		t.Fatalf("expected sizing count; got:\n%s", got)
	}
	if !strings.Contains(got, "[1/3] extracted    ok") {
		t.Fatalf("expected extracted ok; got:\n%s", got)
	}
	if !strings.Contains(got, "[2/3] download") || !strings.Contains(got, "[3/3] vcs") {
		t.Fatalf("expected download and vcs stages; got:\n%s", got)
	}
	// Heartbeats for 2 versions: first and last.
	if !strings.Contains(got, "1/2  example.com/foo@") || !strings.Contains(got, "2/2  example.com/foo@") {
		t.Fatalf("expected 1/2 and 2/2 heartbeats; got:\n%s", got)
	}
}
