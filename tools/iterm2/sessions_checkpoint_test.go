package iterm2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/computer-use/macos/space"
)

func TestBuildSaveDocument_FiltersAndKinds(t *testing.T) {
	now := time.Date(2026, 7, 25, 18, 0, 0, 0, time.FixedZone("CST", 8*3600))
	snap := &Snapshot{
		Windows: []SnapshotWindow{
			{
				Index: 1,
				Name:  "Win-A",
				Tabs: []SnapshotTab{
					{
						Index: 1,
						Name:  "idle",
						Sessions: []SnapshotSession{
							{Index: 1, ID: "IDLE-1", Cwd: strPtr("/idle"), Idle: boolPtr(true), Command: strPtr("zsh")},
						},
					},
					{
						Index: 2,
						Name:  "grok-tab",
						Sessions: []SnapshotSession{
							{
								Index: 1, ID: "GROK-1", Cwd: strPtr("/proj/a"),
								Command: strPtr("grok"),
								Agent:   &SessionAgent{Kind: "grok", SessionID: "sess-grok-1", Title: "t1"},
							},
						},
					},
					{
						Index: 3,
						Name:  "busy-no-agent",
						Sessions: []SnapshotSession{
							{Index: 1, ID: "BUSY-1", Cwd: strPtr("/proj/b"), Idle: boolPtr(false), Command: strPtr("python")},
						},
					},
				},
			},
			{
				Index: 2,
				Name:  "Win-B",
				Tabs: []SnapshotTab{
					{
						Index: 1,
						Sessions: []SnapshotSession{
							{
								Index: 1, ID: "CODEX-1", Cwd: strPtr("/proj/c"),
								Agent: &SessionAgent{Kind: "codex", SessionID: "sess-codex-1"},
							},
						},
					},
					{
						Index: 2,
						Sessions: []SnapshotSession{
							{
								Index: 1, ID: "MARK-1", Cwd: strPtr("/proj/d"),
								Command:     strPtr("mark"),
								CommandLine: strPtr("mark still waiting for CI"),
								Processes: []SnapshotProc{
									{PID: 1, Command: "login"},
									{PID: 2, Command: "-zsh"},
									{PID: 3, Command: "mark still waiting for CI"},
								},
							},
						},
					},
					{
						Index: 3,
						Sessions: []SnapshotSession{
							{
								// agent preferred over mark on same pane
								Index: 1, ID: "BOTH-1", Cwd: strPtr("/proj/e"),
								Command:     strPtr("mark"),
								CommandLine: strPtr("mark ignored"),
								Agent:       &SessionAgent{Kind: "grok", SessionID: "sess-prefer"},
							},
						},
					},
					{
						Index: 4,
						Sessions: []SnapshotSession{
							// empty cwd skip
							{
								Index: 1, ID: "NOCWD-1",
								Agent: &SessionAgent{Kind: "grok", SessionID: "sess-nocwd"},
							},
						},
					},
				},
			},
		},
	}

	doc, warns := BuildSaveDocument(snap, now, "testhost")
	if doc.Version != 1 {
		t.Fatalf("version %d", doc.Version)
	}
	if doc.Host != "testhost" {
		t.Fatal(doc.Host)
	}
	if doc.RestoredAt != nil {
		t.Fatal("restored_at should be null")
	}
	if doc.Summary.Sessions != 4 {
		t.Fatalf("sessions=%d want 4; windows=%+v warns=%v", doc.Summary.Sessions, doc.Windows, warns)
	}
	if doc.Summary.ByKind["grok"] != 2 || doc.Summary.ByKind["codex"] != 1 || doc.Summary.ByKind["mark"] != 1 {
		t.Fatalf("by_kind=%v", doc.Summary.ByKind)
	}
	if len(warns) == 0 || !strings.Contains(warns[0], "empty cwd") {
		t.Fatalf("expected empty cwd warning: %v", warns)
	}

	// find mark tab
	var mark *SaveTab
	var prefer *SaveTab
	for _, w := range doc.Windows {
		for i := range w.Tabs {
			tab := &w.Tabs[i]
			if tab.Kind == "mark" {
				mark = tab
			}
			if tab.SessionID == "sess-prefer" {
				prefer = tab
			}
		}
	}
	if mark == nil || mark.Message != "still waiting for CI" {
		t.Fatalf("mark tab: %+v", mark)
	}
	if !strings.HasPrefix(mark.ResumeCmd, "mark ") {
		t.Fatal(mark.ResumeCmd)
	}
	if prefer == nil || prefer.Kind != "grok" {
		t.Fatalf("prefer agent: %+v", prefer)
	}
	if prefer.ResumeCmd != "grok --resume sess-prefer" {
		t.Fatal(prefer.ResumeCmd)
	}
}

func TestResumeCmdShapes(t *testing.T) {
	if got := resumeCmdForAgent("grok", "abc"); got != "grok --resume abc" {
		t.Fatal(got)
	}
	if got := resumeCmdForAgent("codex", "xyz"); got != "codex resume xyz" {
		t.Fatal(got)
	}
	if got := resumeCmdForMark(""); got != "mark" {
		t.Fatal(got)
	}
	if got := resumeCmdForMark("I'm waiting"); got != `mark 'I'\''m waiting'` {
		t.Fatal(got)
	}
}

func TestWriteReadSaveDocument_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sessions-save.json")
	doc := &SaveDocument{
		Version: 1,
		SavedAt: "2026-07-25T18:00:00+0800",
		Host:    "h",
		Source:  sessionsSaveSource,
		Summary: SaveSummary{Windows: 1, Tabs: 1, Sessions: 1, ByKind: map[string]int{"mark": 1}},
		Windows: []SaveWindow{{
			SourceIndex: 1,
			Tabs: []SaveTab{{
				Cwd: "/tmp", Kind: "mark", Message: "hi", ResumeCmd: "mark 'hi'",
			}},
		}},
	}
	if err := WriteSaveDocument(path, doc); err != nil {
		t.Fatal(err)
	}
	got, err := ReadSaveDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.IsConsumed() {
		t.Fatal("should not be consumed")
	}
	if got.Windows[0].Tabs[0].Kind != "mark" {
		t.Fatal(got.Windows[0].Tabs[0])
	}
	// stamp restored
	ts := "2026-07-25T20:00:00+0800"
	got.RestoredAt = &ts
	if err := WriteSaveDocument(path, got); err != nil {
		t.Fatal(err)
	}
	got2, err := ReadSaveDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	if !got2.IsConsumed() {
		t.Fatal("expected consumed")
	}
	// JSON null for restored_at when nil
	doc.RestoredAt = nil
	if err := WriteSaveDocument(path, doc); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	if !bytes.Contains(raw, []byte(`"restored_at": null`)) {
		t.Fatalf("want null restored_at:\n%s", raw)
	}
}

func TestSessionsSave_CLI_DryRunAndWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "save.json")
	prevPath := sessionsSavePathForTest
	sessionsSavePathForTest = path
	t.Cleanup(func() { sessionsSavePathForTest = prevPath })

	prevTTY := sessionsIsStdinTTY
	sessionsIsStdinTTY = func() bool { return false }
	t.Cleanup(func() { sessionsIsStdinTTY = prevTTY })

	InstallPhasedFixtureCollectorForTest(t, PhasedFixtureOpts{
		ITermRunning: true,
		Windows: []SnapshotWindow{
			{
				Index: 1, Name: "W",
				Tabs: []SnapshotTab{
					{Index: 1, Sessions: []SnapshotSession{
						{Index: 1, ID: "A-1", TTY: "/dev/ttys010", Profile: "Default"},
					}},
					{Index: 2, Sessions: []SnapshotSession{
						{Index: 1, ID: "M-1", TTY: "/dev/ttys011", Profile: "Default"},
					}},
					{Index: 3, Sessions: []SnapshotSession{
						{Index: 1, ID: "I-1", TTY: "/dev/ttys012", Profile: "Default"},
					}},
				},
			},
		},
		BusyTTYs: []string{"ttys010", "ttys011"},
		IdleTTYs: []string{"ttys012"},
		BusyLeafByTTY: map[string]string{
			"ttys011": "mark still waiting",
		},
		CwdByTTY: map[string]string{
			"ttys010": "/proj/grok",
			"ttys011": "/proj/mark",
		},
		AgentResolveByTTY: map[string]AgentResolveFixture{
			"ttys010": {Kind: "grok", SessionID: "g-sess-1", Title: "t"},
		},
		Now:      time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
		Hostname: "testhost",
	})

	var stdout, stderr bytes.Buffer
	if err := runSessions([]string{"save", "--dry-run"}, &stdout, &stderr); err != nil {
		t.Fatalf("dry-run: %v stderr=%s", err, stderr.String())
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("dry-run must not write")
	}
	out := stdout.String()
	if !strings.Contains(out, "Would save") || !strings.Contains(out, "g-sess-1") {
		t.Fatalf("dry-run stdout:\n%s", out)
	}
	if !strings.Contains(out, "mark") {
		t.Fatalf("dry-run missing mark:\n%s", out)
	}

	stdout.Reset()
	stderr.Reset()
	if err := runSessions([]string{"save"}, &stdout, &stderr); err != nil {
		t.Fatalf("save: %v stderr=%s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Saved") {
		t.Fatal(stdout.String())
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var doc SaveDocument
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Summary.Sessions != 2 {
		t.Fatalf("sessions=%d doc=%s", doc.Summary.Sessions, raw)
	}
	if doc.IsConsumed() {
		t.Fatal("should not be consumed after save")
	}
}

func TestSessionsSave_OverwriteNonTTY_Pending(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "save.json")
	// pending existing
	pending := &SaveDocument{
		Version: 1, SavedAt: "old", Host: "h", Source: sessionsSaveSource,
		Summary: SaveSummary{Sessions: 1, Tabs: 1, Windows: 1, ByKind: map[string]int{"grok": 1}},
		Windows: []SaveWindow{{SourceIndex: 1, Tabs: []SaveTab{{Cwd: "/x", Kind: "grok", SessionID: "old", ResumeCmd: "grok --resume old"}}}},
	}
	if err := WriteSaveDocument(path, pending); err != nil {
		t.Fatal(err)
	}

	prevPath := sessionsSavePathForTest
	sessionsSavePathForTest = path
	t.Cleanup(func() { sessionsSavePathForTest = prevPath })
	prevTTY := sessionsIsStdinTTY
	sessionsIsStdinTTY = func() bool { return false }
	t.Cleanup(func() { sessionsIsStdinTTY = prevTTY })

	InstallPhasedFixtureCollectorForTest(t, PhasedFixtureOpts{
		ITermRunning: true,
		Windows: []SnapshotWindow{{
			Index: 1,
			Tabs: []SnapshotTab{{Index: 1, Sessions: []SnapshotSession{
				{Index: 1, ID: "A", TTY: "/dev/ttys020", Profile: "Default"},
			}}},
		}},
		BusyTTYs: []string{"ttys020"},
		AgentResolveByTTY: map[string]AgentResolveFixture{
			"ttys020": {Kind: "grok", SessionID: "new-sess"},
		},
		Hostname: "testhost",
	})

	var stdout, stderr bytes.Buffer
	err := runSessions([]string{"save"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected non-TTY overwrite error")
	}
	if !strings.Contains(stderr.String(), "not restored") && !strings.Contains(stderr.String(), "TTY") {
		t.Fatal(stderr.String())
	}
	// file unchanged
	got, _ := ReadSaveDocument(path)
	if got.Windows[0].Tabs[0].SessionID != "old" {
		t.Fatal("file was overwritten")
	}
}

func TestSessionsSave_OverwriteAlreadyRestored(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "save.json")
	ts := "2026-07-25T10:00:00+0800"
	old := &SaveDocument{
		Version: 1, SavedAt: "old", RestoredAt: &ts, Host: "h", Source: sessionsSaveSource,
		Summary: SaveSummary{Sessions: 1, Tabs: 1, Windows: 1, ByKind: map[string]int{"grok": 1}},
		Windows: []SaveWindow{{SourceIndex: 1, Tabs: []SaveTab{{Cwd: "/x", Kind: "grok", SessionID: "old", ResumeCmd: "grok --resume old"}}}},
	}
	if err := WriteSaveDocument(path, old); err != nil {
		t.Fatal(err)
	}

	prevPath := sessionsSavePathForTest
	sessionsSavePathForTest = path
	t.Cleanup(func() { sessionsSavePathForTest = prevPath })
	prevTTY := sessionsIsStdinTTY
	sessionsIsStdinTTY = func() bool { return false } // no prompt expected
	t.Cleanup(func() { sessionsIsStdinTTY = prevTTY })

	InstallPhasedFixtureCollectorForTest(t, PhasedFixtureOpts{
		ITermRunning: true,
		Windows: []SnapshotWindow{{
			Index: 1,
			Tabs: []SnapshotTab{{Index: 1, Sessions: []SnapshotSession{
				{Index: 1, ID: "A", TTY: "/dev/ttys021", Profile: "Default"},
			}}},
		}},
		BusyTTYs: []string{"ttys021"},
		AgentResolveByTTY: map[string]AgentResolveFixture{
			"ttys021": {Kind: "codex", SessionID: "fresh-codex"},
		},
		Hostname: "testhost",
	})

	var stdout, stderr bytes.Buffer
	if err := runSessions([]string{"save"}, &stdout, &stderr); err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	got, err := ReadSaveDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.IsConsumed() {
		t.Fatal("restored_at should be cleared on new save")
	}
	if got.Windows[0].Tabs[0].SessionID != "fresh-codex" {
		t.Fatal(got.Windows[0].Tabs[0])
	}
}

func TestSessionsRestore_ConsumedAndDryRun(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "save.json")
	doc := &SaveDocument{
		Version: 1, SavedAt: "2026-07-25T18:00:00+0800", Host: "testhost", Source: sessionsSaveSource,
		Summary: SaveSummary{Windows: 1, Tabs: 2, Sessions: 2, ByKind: map[string]int{"grok": 1, "mark": 1}},
		Windows: []SaveWindow{{
			SourceIndex: 1, Name: "N",
			Tabs: []SaveTab{
				{Cwd: "/a", Kind: "grok", SessionID: "g1", ResumeCmd: "grok --resume g1"},
				{Cwd: "/b", Kind: "mark", Message: "wait", ResumeCmd: "mark 'wait'"},
			},
		}},
	}
	if err := WriteSaveDocument(path, doc); err != nil {
		t.Fatal(err)
	}

	prevPath := sessionsSavePathForTest
	sessionsSavePathForTest = path
	t.Cleanup(func() { sessionsSavePathForTest = prevPath })

	var stdout, stderr bytes.Buffer
	if err := runSessions([]string{"restore", "--dry-run"}, &stdout, &stderr); err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "Would restore") || !strings.Contains(out, "grok --resume g1") {
		t.Fatal(out)
	}
	if !strings.Contains(out, "mark 'wait'") {
		t.Fatal(out)
	}
	// not stamped
	got, _ := ReadSaveDocument(path)
	if got.IsConsumed() {
		t.Fatal("dry-run must not stamp")
	}

	// real restore with mocked AS + Space backend (no live Mission Control)
	SetSpaceBackendForTest(&space.MockBackend{Desktops: []int{1}})
	t.Cleanup(func() { SetSpaceBackendForTest(nil) })

	var scripts []string
	prevAS := sessionsRunRestoreAS
	sessionsRunRestoreAS = func(script string) (string, error) {
		scripts = append(scripts, script)
		return "ok", nil
	}
	t.Cleanup(func() { sessionsRunRestoreAS = prevAS })

	stdout.Reset()
	stderr.Reset()
	if err := runSessions([]string{"restore"}, &stdout, &stderr); err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	if len(scripts) != 1 {
		t.Fatalf("scripts=%d", len(scripts))
	}
	sc := scripts[0]
	if !strings.Contains(sc, "create window") || !strings.Contains(sc, "create tab") {
		t.Fatal(sc)
	}
	if !strings.Contains(sc, "grok --resume g1") || !strings.Contains(sc, "mark 'wait'") {
		t.Fatal(sc)
	}
	if !strings.Contains(stdout.String(), "Restored") {
		t.Fatal(stdout.String())
	}
	got, err := ReadSaveDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	if !got.IsConsumed() {
		t.Fatal("expected restored_at")
	}

	// second restore errors
	stdout.Reset()
	stderr.Reset()
	err = runSessions([]string{"restore"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected consumed error")
	}
	if !strings.Contains(stderr.String(), "consumed") && !strings.Contains(stderr.String(), "restored_at") {
		t.Fatal(stderr.String())
	}
}

func TestSessionsRestore_MissingFile(t *testing.T) {
	prevPath := sessionsSavePathForTest
	sessionsSavePathForTest = filepath.Join(t.TempDir(), "nope.json")
	t.Cleanup(func() { sessionsSavePathForTest = prevPath })
	var stdout, stderr bytes.Buffer
	err := runSessions([]string{"restore"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr.String(), "not found") {
		t.Fatal(stderr.String())
	}
}

func TestSessionsSave_ZeroCritical(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "save.json")
	prevPath := sessionsSavePathForTest
	sessionsSavePathForTest = path
	t.Cleanup(func() { sessionsSavePathForTest = prevPath })

	InstallPhasedFixtureCollectorForTest(t, PhasedFixtureOpts{
		ITermRunning: true,
		Windows: []SnapshotWindow{{
			Index: 1,
			Tabs: []SnapshotTab{{Index: 1, Sessions: []SnapshotSession{
				{Index: 1, ID: "I", TTY: "/dev/ttys030", Profile: "Default"},
			}}},
		}},
		IdleTTYs: []string{"ttys030"},
		Hostname: "testhost",
	})
	var stdout, stderr bytes.Buffer
	if err := runSessions([]string{"save"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "0 critical") {
		t.Fatal(stdout.String())
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("must not write empty")
	}
}

func TestSessionsHelp_MentionsSaveRestore(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := runSessions([]string{"-h"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	out := stdout.String()
	for _, want := range []string{"save", "restore", "snapshot"} {
		if !strings.Contains(out, want) {
			t.Fatalf("help missing %q:\n%s", want, out)
		}
	}
	stdout.Reset()
	if err := runSessions([]string{"save", "-h"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "--dry-run") || !strings.Contains(stdout.String(), "--file") {
		t.Fatal(stdout.String())
	}
	stdout.Reset()
	if err := runSessions([]string{"restore", "-h"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "restored_at") {
		t.Fatal(stdout.String())
	}
}

func TestBuildSessionsRestoreScript_TwoTabs(t *testing.T) {
	doc := &SaveDocument{
		Windows: []SaveWindow{{
			Name: "Win",
			Tabs: []SaveTab{
				{Cwd: "/a", ResumeCmd: "grok --resume x"},
				{Cwd: "/b", ResumeCmd: "mark 'm'"},
			},
		}},
	}
	sc := BuildSessionsRestoreScript(doc)
	if !strings.Contains(sc, "create window with default profile") {
		t.Fatal(sc)
	}
	if !strings.Contains(sc, "create tab with default profile") {
		t.Fatal(sc)
	}
	if strings.Count(sc, "write text") < 4 {
		t.Fatalf("expected cd+cmd ×2 write texts:\n%s", sc)
	}
	// Title set is best-effort: try/on error so restore continues.
	if !strings.Contains(sc, "set titleWarnings to \"\"") {
		t.Fatalf("missing titleWarnings init:\n%s", sc)
	}
	if !strings.Contains(sc, "try") || !strings.Contains(sc, "on error errMsg") || !strings.Contains(sc, "end try") {
		t.Fatalf("set name must be wrapped in try/on error:\n%s", sc)
	}
	if !strings.Contains(sc, `set name of newWindow to desiredTitle`) {
		t.Fatalf("missing set name via desiredTitle:\n%s", sc)
	}
	if !strings.Contains(sc, "return titleWarnings") {
		t.Fatalf("missing return titleWarnings:\n%s", sc)
	}
	if !strings.Contains(sc, `set desiredTitle to "Win"`) {
		t.Fatalf("missing desiredTitle for named window:\n%s", sc)
	}
}

func TestBuildSessionsRestoreScript_EmptyNameSkipsSetName(t *testing.T) {
	doc := &SaveDocument{
		Windows: []SaveWindow{{
			Name: "",
			Tabs: []SaveTab{{Cwd: "/a", ResumeCmd: "grok --resume x"}},
		}},
	}
	sc := BuildSessionsRestoreScript(doc)
	if strings.Contains(sc, "set name of newWindow") || strings.Contains(sc, "desiredTitle") {
		t.Fatalf("empty Name must not emit set name:\n%s", sc)
	}
	// Still initializes and returns titleWarnings for a uniform contract.
	if !strings.Contains(sc, "return titleWarnings") {
		t.Fatalf("missing return titleWarnings:\n%s", sc)
	}
}

func TestSessionsRestore_TitleWarningStillStamps(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "save.json")
	doc := &SaveDocument{
		Version:    sessionsSaveVersion,
		SavedAt:    "2026-07-25T12:00:00+0800",
		RestoredAt: nil,
		Host:       "testhost",
		Source:     sessionsSaveSource,
		Summary:    SaveSummary{Windows: 1, Tabs: 1, Sessions: 1, ByKind: map[string]int{"grok": 1}},
		Windows: []SaveWindow{{
			SourceIndex: 1,
			Name:        `Bounding walk.jsonl Size… - grok`,
			Tabs: []SaveTab{{
				Cwd: "/proj", Kind: "grok", SessionID: "g1",
				ResumeCmd: "grok --resume g1",
			}},
		}},
	}
	if err := WriteSaveDocument(path, doc); err != nil {
		t.Fatal(err)
	}

	prevPath := sessionsSavePathForTest
	sessionsSavePathForTest = path
	t.Cleanup(func() { sessionsSavePathForTest = prevPath })

	SetSpaceBackendForTest(&space.MockBackend{Desktops: []int{1}})
	t.Cleanup(func() { SetSpaceBackendForTest(nil) })

	warnLine := `could not set window title "Bounding walk.jsonl Size… - grok": Can't set name of window`
	prevAS := sessionsRunRestoreAS
	sessionsRunRestoreAS = func(script string) (string, error) {
		if !strings.Contains(script, "on error errMsg") {
			t.Fatalf("script missing soft-fail title:\n%s", script)
		}
		return warnLine + "\n", nil
	}
	t.Cleanup(func() { sessionsRunRestoreAS = prevAS })

	var stdout, stderr bytes.Buffer
	if err := runSessions([]string{"restore"}, &stdout, &stderr); err != nil {
		t.Fatalf("title failure must not fail restore: %v stderr=%s", err, stderr.String())
	}
	if !strings.Contains(stderr.String(), "warning:") {
		t.Fatalf("expected warning on stderr:\n%s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "could not set window title") {
		t.Fatalf("stderr:\n%s", stderr.String())
	}
	if !strings.Contains(stdout.String(), "Restored") {
		t.Fatalf("stdout:\n%s", stdout.String())
	}
	got, err := ReadSaveDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	if !got.IsConsumed() {
		t.Fatal("expected restored_at stamped after title-only failure")
	}
}

func TestSessionsRestore_AppleScriptHardErrorNoStamp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "save.json")
	doc := &SaveDocument{
		Version:    sessionsSaveVersion,
		SavedAt:    "2026-07-25T12:00:00+0800",
		RestoredAt: nil,
		Host:       "testhost",
		Source:     sessionsSaveSource,
		Summary:    SaveSummary{Windows: 1, Tabs: 1, Sessions: 1, ByKind: map[string]int{"grok": 1}},
		Windows: []SaveWindow{{
			SourceIndex: 1,
			Name:        "Win",
			Tabs: []SaveTab{{
				Cwd: "/proj", Kind: "grok", SessionID: "g1",
				ResumeCmd: "grok --resume g1",
			}},
		}},
	}
	if err := WriteSaveDocument(path, doc); err != nil {
		t.Fatal(err)
	}

	prevPath := sessionsSavePathForTest
	sessionsSavePathForTest = path
	t.Cleanup(func() { sessionsSavePathForTest = prevPath })

	SetSpaceBackendForTest(&space.MockBackend{Desktops: []int{1}})
	t.Cleanup(func() { SetSpaceBackendForTest(nil) })

	prevAS := sessionsRunRestoreAS
	sessionsRunRestoreAS = func(script string) (string, error) {
		return "", fmt.Errorf("iTerm not running")
	}
	t.Cleanup(func() { sessionsRunRestoreAS = prevAS })

	var stdout, stderr bytes.Buffer
	err := runSessions([]string{"restore"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected hard error")
	}
	if !strings.Contains(stderr.String(), "AppleScript failed") {
		t.Fatalf("stderr:\n%s", stderr.String())
	}
	got, _ := ReadSaveDocument(path)
	if got.IsConsumed() {
		t.Fatal("must not stamp restored_at on hard AS failure")
	}
}
