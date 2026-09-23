package iterm2

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func installListFixture(t *testing.T, cwdGrok, cwdOther string) {
	t.Helper()
	InstallPhasedFixtureCollectorForTest(t, PhasedFixtureOpts{
		ITermRunning: true,
		Windows: []SnapshotWindow{
			{
				Index: 1,
				Name:  "W1",
				Tabs: []SnapshotTab{
					{
						Index: 1,
						Name:  "T1",
						Sessions: []SnapshotSession{
							{
								Index: 1, ID: "AAAAAAAA-1111-1111-1111-111111111111",
								Name: "grok-pane", TTY: "/dev/ttys010", Profile: "Default",
							},
							{
								Index: 2, ID: "BBBBBBBB-2222-2222-2222-222222222222",
								Name: "idle-pane", TTY: "/dev/ttys011", Profile: "Default",
							},
							{
								Index: 3, ID: "CCCCCCCC-3333-3333-3333-333333333333",
								Name: "codex-pane", TTY: "/dev/ttys012", Profile: "Default",
							},
						},
					},
				},
			},
		},
		BusyTTYs: []string{"ttys010", "ttys012"},
		IdleTTYs: []string{"ttys011"},
		CwdByTTY: map[string]string{
			"ttys010": cwdGrok,
			"ttys011": cwdOther,
			"ttys012": cwdOther,
		},
		AgentResolveByTTY: map[string]AgentResolveFixture{
			"ttys010": {Kind: "grok", SessionID: "019fabcdef-1234-5678-9abc-def012345678", Title: "t"},
			"ttys012": {Kind: "codex", SessionID: "codex-sess-99"},
		},
		Now:      time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
		Hostname: "testhost",
	})
}

func TestSessionListHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := runSession([]string{"list", "-h"}, &stdout, &stderr, TestRun{}); err != nil {
		t.Fatal(err)
	}
	out := stdout.String()
	for _, want := range []string{"--grok", "--only-cwd", "--json", "session list"} {
		if !strings.Contains(out, want) {
			t.Fatalf("help missing %q:\n%s", want, out)
		}
	}
}

func TestSessionListTable(t *testing.T) {
	dir := t.TempDir()
	other := filepath.Join(dir, "other")
	if err := os.MkdirAll(other, 0o755); err != nil {
		t.Fatal(err)
	}
	installListFixture(t, dir, other)

	var stdout, stderr bytes.Buffer
	if err := runSession([]string{"list", "--no-color"}, &stdout, &stderr, TestRun{}); err != nil {
		t.Fatalf("err=%v stderr=%s", err, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "AAAAAAAA") || !strings.Contains(out, "grok") {
		t.Fatalf("missing grok row:\n%s", out)
	}
	if !strings.Contains(out, "019fabcdef-1234-5678-9abc-def012345678") {
		t.Fatalf("missing agent id:\n%s", out)
	}
	if !strings.Contains(out, "codex") {
		t.Fatalf("missing codex row:\n%s", out)
	}
	if !strings.Contains(out, "3 sessions") {
		t.Fatalf("footer:\n%s", out)
	}
}

func TestSessionListGrokOnly(t *testing.T) {
	dir := t.TempDir()
	other := filepath.Join(dir, "other")
	_ = os.MkdirAll(other, 0o755)
	installListFixture(t, dir, other)

	var stdout, stderr bytes.Buffer
	if err := runSession([]string{"list", "--grok", "--no-color"}, &stdout, &stderr, TestRun{}); err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "grok") || !strings.Contains(out, "1 session") {
		t.Fatalf("expected one grok:\n%s", out)
	}
	if strings.Contains(out, "codex") {
		t.Fatalf("codex should be filtered:\n%s", out)
	}
}

func TestSessionListGrokOnlyCwd(t *testing.T) {
	dir := t.TempDir()
	// Process cwd for --only-cwd is WorkingDir of the test process; chdir into dir.
	prev, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })

	other := filepath.Join(dir, "other")
	_ = os.MkdirAll(other, 0o755)
	// Fixture cwd for grok must match process cwd (dir).
	installListFixture(t, dir, other)

	var stdout, stderr bytes.Buffer
	if err := runSession([]string{"list", "--grok", "--only-cwd", "--no-color"}, &stdout, &stderr, TestRun{}); err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "1 session") || !strings.Contains(out, "grok") {
		t.Fatalf("expected one matching grok:\n%s", out)
	}
}

func TestSessionListJSON(t *testing.T) {
	dir := t.TempDir()
	other := filepath.Join(dir, "other")
	_ = os.MkdirAll(other, 0o755)
	installListFixture(t, dir, other)

	var stdout, stderr bytes.Buffer
	if err := runSession([]string{"list", "--json"}, &stdout, &stderr, TestRun{}); err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	if strings.Contains(stdout.String(), "\x1b[") {
		t.Fatal("ANSI in json")
	}
	var rows []SessionListRow
	if err := json.Unmarshal(stdout.Bytes(), &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 {
		t.Fatalf("rows=%d", len(rows))
	}
	var foundGrok bool
	for _, r := range rows {
		if r.Kind == "grok" && r.AgentSession != "" {
			foundGrok = true
		}
	}
	if !foundGrok {
		t.Fatalf("no grok row: %+v", rows)
	}
}

func TestFilterSessionList(t *testing.T) {
	rows := []SessionListRow{
		{ItermID: "a", Kind: "grok", AgentSession: "g1", Cwd: "/proj"},
		{ItermID: "b", Kind: "codex", AgentSession: "c1", Cwd: "/proj"},
		{ItermID: "c", Kind: "grok", AgentSession: "g2", Cwd: "/other"},
		{ItermID: "d", Kind: "", Cwd: "/proj"},
	}
	got := FilterSessionList(rows, true, true, "/proj")
	if len(got) != 1 || got[0].ItermID != "a" {
		t.Fatalf("%+v", got)
	}
}

func TestSessionListEmpty(t *testing.T) {
	InstallPhasedFixtureCollectorForTest(t, PhasedFixtureOpts{
		ITermRunning: true,
		Windows:      []SnapshotWindow{},
		Hostname:     "h",
	})
	var stdout, stderr bytes.Buffer
	if err := runSession([]string{"list", "--no-color"}, &stdout, &stderr, TestRun{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "0 sessions") {
		t.Fatal(stdout.String())
	}
}
