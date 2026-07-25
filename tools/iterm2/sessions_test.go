package iterm2

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func installFakeSnapshot(t *testing.T) {
	t.Helper()
	InstallPhasedFixtureCollectorForTest(t, PhasedFixtureOpts{
		ITermRunning: true,
		Windows: []SnapshotWindow{
			{
				Index: 1,
				Name:  "Win",
				Tabs: []SnapshotTab{
					{
						Index: 1,
						Name:  "Tab",
						Sessions: []SnapshotSession{
							{
								Index:   1,
								ID:      "11111111-2222-3333-4444-555555555555",
								Name:    "name-a",
								TTY:     "/dev/ttys010",
								Profile: "Default",
							},
						},
					},
				},
			},
		},
		IdleTTYs: []string{"ttys010"},
		Now:      time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
		Hostname: "host",
	})
}

func TestSessionsSnapshotHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := runSessions([]string{"-h"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	out := stdout.String()
	for _, want := range []string{"snapshot", "--json", "--markdown", "--html", "-o"} {
		if !strings.Contains(out, want) {
			t.Fatalf("help missing %q:\n%s", want, out)
		}
	}
}

func TestSessionsSnapshotJSON(t *testing.T) {
	installFakeSnapshot(t)
	var stdout, stderr bytes.Buffer
	if err := runSessions([]string{"snapshot", "--json", "--no-color"}, &stdout, &stderr); err != nil {
		t.Fatalf("err=%v stderr=%s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"id": "11111111-2222-3333-4444-555555555555"`) {
		t.Fatal(stdout.String())
	}
	if strings.Contains(stdout.String(), "\x1b[") {
		t.Fatal("ANSI in json")
	}
}

func TestSessionsSnapshotOutputFile(t *testing.T) {
	installFakeSnapshot(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "snap.json")
	var stdout, stderr bytes.Buffer
	if err := runSessions([]string{"snapshot", "-o", path}, &stdout, &stderr); err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout should be empty, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "Wrote") {
		t.Fatal(stderr.String())
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"source": "iterm2"`) {
		t.Fatal(string(b))
	}
}

func TestSessionsSnapshotFormatConflict(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := runSessions([]string{"snapshot", "--json", "--html"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr.String(), "mutually exclusive") {
		t.Fatal(stderr.String())
	}
}

func TestSessionStatusByPrefix(t *testing.T) {
	installFakeSnapshot(t)
	var stdout, stderr bytes.Buffer
	if err := runSession([]string{"11111111", "status", "--no-color"}, &stdout, &stderr); err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "11111111-2222-3333-4444-555555555555") {
		t.Fatal(stdout.String())
	}
	if !strings.Contains(stdout.String(), "idle") {
		t.Fatal(stdout.String())
	}
}

func TestSessionStatusNotFound(t *testing.T) {
	installFakeSnapshot(t)
	var stdout, stderr bytes.Buffer
	err := runSession([]string{"deadbeef", "status"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr.String(), "not found") {
		t.Fatal(stderr.String())
	}
}

func TestSessionHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := runSession([]string{"-h"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "status") {
		t.Fatal(stdout.String())
	}
}
