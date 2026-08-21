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

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/dot-pkgs/go-pkgs/computer-use/macos/space"
)

func installEmptyLiveCollectorForRestoreTest(t *testing.T) {
	t.Helper()
	InstallPhasedFixtureCollectorForTest(t, PhasedFixtureOpts{
		ITermRunning: true,
		Hostname:     "testhost",
	})
}

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
	// Empty fixture so the already-running scan does not hit live AppleScript.
	installEmptyLiveCollectorForRestoreTest(t)

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

	// real restore with mocked AS + Space backend (no live Mission Control / CGS)
	SetSpaceBackendForTest(&space.MockBackend{Desktops: []int{1}})
	t.Cleanup(func() { SetSpaceBackendForTest(nil) })
	SetCurrentSpaceIndexForTest(func() (int, error) { return 0, nil }) // ckpt space 0 → skip Switch
	t.Cleanup(func() { SetCurrentSpaceIndexForTest(nil) })

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
	if !strings.Contains(stderr.String(), "--force") {
		t.Fatal(stderr.String())
	}

	// --force permits inspection or reapplication of a consumed checkpoint;
	// it still runs the live-session duplicate check.
	stdout.Reset()
	stderr.Reset()
	if err := runSessions([]string{"restore", "--force", "--dry-run"}, &stdout, &stderr); err != nil {
		t.Fatalf("forced dry-run: %v stderr=%s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Would restore") {
		t.Fatal(stdout.String())
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
	if !strings.Contains(stdout.String(), "restored_at") || !strings.Contains(stdout.String(), "--force") {
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
	installEmptyLiveCollectorForRestoreTest(t)
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
	SetCurrentSpaceIndexForTest(func() (int, error) { return 0, nil })
	t.Cleanup(func() { SetCurrentSpaceIndexForTest(nil) })

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
	installEmptyLiveCollectorForRestoreTest(t)
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
	SetCurrentSpaceIndexForTest(func() (int, error) { return 0, nil })
	t.Cleanup(func() { SetCurrentSpaceIndexForTest(nil) })

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

func TestSortSaveWindowsBySpace(t *testing.T) {
	windows := []SaveWindow{
		{Name: "c", Space: 12, App: CanonicalITermAppHome, ItermWindowID: 3},
		{Name: "a", Space: 10, App: CanonicalITermAppHome, ItermWindowID: 1},
		{Name: "b", Space: 11, App: CanonicalITermAppSystem, ItermWindowID: 2},
		{Name: "a2", Space: 10, App: CanonicalITermAppSystem, ItermWindowID: 4},
	}
	sortSaveWindowsBySpace(windows)
	// space asc, then app, then name
	wantSpaces := []int{10, 10, 11, 12}
	wantNames := []string{"a2", "a", "b", "c"} // app system < home for space 10? "/" vs "~" — "/" < "~"
	// CanonicalITermAppSystem = "/Applications/..." < CanonicalITermAppHome = "~/..."
	for i, sp := range wantSpaces {
		if windows[i].Space != sp {
			t.Fatalf("i=%d space=%d want %d windows=%+v", i, windows[i].Space, sp, windows)
		}
	}
	if windows[0].Name != "a2" || windows[0].App != CanonicalITermAppSystem {
		t.Fatalf("space 10 first should be system a2: %+v", windows[0])
	}
	if windows[1].Name != "a" {
		t.Fatalf("space 10 second home a: %+v", windows[1])
	}
	_ = wantNames
	for i := range windows {
		if windows[i].SourceIndex != i+1 {
			t.Fatalf("source_index not renumbered: %+v", windows)
		}
	}
}

func TestFormatSaveWindowBlock_NoLeadingBlankOnFirst(t *testing.T) {
	win := SaveWindow{
		SourceIndex: 1,
		Name:        "work",
		Space:       1,
		Tabs: []SaveTab{{
			Kind: "grok", SessionID: "sess-1", Cwd: "/proj", ResumeCmd: "grok --resume sess-1",
		}},
	}
	var buf bytes.Buffer
	formatSaveWindowBlock(&buf, win, false, false)
	out := buf.String()
	if strings.HasPrefix(out, "\n") {
		t.Fatalf("first block must not start with blank line: %q", out)
	}
	if !strings.HasPrefix(out, "  W1") {
		t.Fatalf("want indent + W1; got %q", out)
	}

	buf.Reset()
	formatSaveWindowBlock(&buf, win, false, true)
	out2 := buf.String()
	if !strings.HasPrefix(out2, "\n  W1") {
		t.Fatalf("subsequent block must start with blank line: %q", out2)
	}
}

func TestFormatSavePlan_ColorTokens(t *testing.T) {
	doc := &SaveDocument{
		Summary: SaveSummary{Windows: 1, Tabs: 1, Sessions: 1, ByKind: map[string]int{"grok": 1}},
		Windows: []SaveWindow{{
			SourceIndex: 1, Name: "work", Space: 0,
			Tabs: []SaveTab{{Kind: "grok", SessionID: "s1", Cwd: "/a", ResumeCmd: "grok --resume s1"}},
		}},
	}
	var buf bytes.Buffer
	formatSavePlan(&buf, doc, "/tmp/x.json", true, true)
	out := buf.String()
	if strings.HasPrefix(out, "\n") {
		t.Fatalf("leading blank: %q", out)
	}
	for _, code := range []string{"\x1b[32m", "\x1b[1m", "\x1b[90m"} {
		if !strings.Contains(out, code) {
			t.Fatalf("missing SGR %q in:\n%q", code, out)
		}
	}
	if !strings.Contains(out, "Would save") {
		t.Fatal(out)
	}
}

func TestParseSpacesList(t *testing.T) {
	got, err := parseSpacesList("2,0,2, 1")
	if err != nil {
		t.Fatal(err)
	}
	want := []int{0, 1, 2}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
	if _, err := parseSpacesList(""); err == nil {
		t.Fatal("expected empty error")
	}
	if _, err := parseSpacesList("16"); err == nil {
		t.Fatal("expected range error")
	}
	if _, err := parseSpacesList("-1"); err == nil {
		t.Fatal("expected negative error")
	}
	if _, err := parseSpacesList("x"); err == nil {
		t.Fatal("expected non-int error")
	}
}

func TestFilterSaveDocumentBySpaces(t *testing.T) {
	doc := &SaveDocument{
		Windows: []SaveWindow{
			{SourceIndex: 1, Space: 0, Tabs: []SaveTab{{Kind: "grok", Cwd: "/a", ResumeCmd: "g"}}},
			{SourceIndex: 2, Space: 2, Tabs: []SaveTab{{Kind: "mark", Cwd: "/b", ResumeCmd: "m"}}},
			{SourceIndex: 3, Space: 1, Tabs: []SaveTab{{Kind: "codex", Cwd: "/c", ResumeCmd: "c"}}},
		},
	}
	recomputeSaveSummary(doc)
	skipped := filterSaveDocumentBySpaces(doc, []int{0, 2})
	if skipped != 1 {
		t.Fatalf("skipped=%d", skipped)
	}
	if doc.Summary.Windows != 2 || doc.Summary.Sessions != 2 {
		t.Fatalf("summary=%+v", doc.Summary)
	}
	if len(doc.Windows) != 2 || doc.Windows[0].Space != 0 || doc.Windows[1].Space != 2 {
		t.Fatalf("windows=%+v", doc.Windows)
	}
}

func TestSessionsSave_SpacesFilter_WriteAndSkipWarn(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "spaces.json")
	prevPath := sessionsSavePathForTest
	sessionsSavePathForTest = path
	t.Cleanup(func() { sessionsSavePathForTest = prevPath })

	const (
		wid0 = uint64(1001)
		wid2 = uint64(1002)
	)
	space0, space2 := 0, 2
	InstallPhasedFixtureCollectorForTest(t, PhasedFixtureOpts{
		ITermRunning: true,
		Windows: []SnapshotWindow{
			{
				Index: 1, Name: "On-0", WindowID: wid0, FixedSpace: &space0,
				Tabs: []SnapshotTab{{Index: 1, Sessions: []SnapshotSession{
					{Index: 1, ID: "A", TTY: "/dev/ttys001", Profile: "Default"},
				}}},
			},
			{
				Index: 2, Name: "On-2", WindowID: wid2, FixedSpace: &space2,
				Tabs: []SnapshotTab{{Index: 1, Sessions: []SnapshotSession{
					{Index: 1, ID: "B", TTY: "/dev/ttys002", Profile: "Default"},
				}}},
			},
		},
		BusyTTYs: []string{"ttys001", "ttys002"},
		CwdByTTY: map[string]string{"ttys001": "/proj/a", "ttys002": "/proj/b"},
		AgentResolveByTTY: map[string]AgentResolveFixture{
			"ttys001": {Kind: "grok", SessionID: "sess-space-0"},
			"ttys002": {Kind: "codex", SessionID: "sess-space-2"},
		},
		Hostname: "testhost",
	})

	var stdout, stderr bytes.Buffer
	if err := runSessions([]string{"save", "--spaces", "0"}, &stdout, &stderr); err != nil {
		t.Fatalf("%v stderr=%s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Saved") {
		t.Fatalf("stdout:\n%s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "skipped 1 windows not matching --spaces 0") {
		t.Fatalf("stderr skip warn:\n%s", stderr.String())
	}
	got, err := ReadSaveDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Filter == nil || len(got.Filter.Spaces) != 1 || got.Filter.Spaces[0] != 0 {
		t.Fatalf("filter=%+v", got.Filter)
	}
	if got.Summary.Windows != 1 || got.Summary.Sessions != 1 {
		t.Fatalf("summary=%+v", got.Summary)
	}
	if got.Windows[0].Space != 0 {
		t.Fatalf("space=%d", got.Windows[0].Space)
	}
}

func TestSessionsSave_SpacesConflictIgnore(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := runSessions([]string{"save", "--spaces", "0", "--ignore-macos-space"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected conflict error")
	}
	if !strings.Contains(stderr.String(), "--spaces cannot be used with --ignore-macos-space") {
		t.Fatalf("stderr:\n%s", stderr.String())
	}
}

func TestSessionsSave_SpacesInvalid(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := runSessions([]string{"save", "--spaces", "16"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected invalid space error")
	}
	if !strings.Contains(stderr.String(), "invalid space index 16") {
		t.Fatalf("stderr:\n%s", stderr.String())
	}
}

func TestListTabsAndSessionsAppleScript_UsesStableIndexes(t *testing.T) {
	script := listTabsAndSessionsAppleScript(4, "/tmp/iTerm.app")
	for _, want := range []string{
		"set tabCount to count of tabs of window target",
		"session si of tab ti of window target",
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("missing %q:\n%s", want, script)
		}
	}
	if strings.Contains(script, "repeat with t in tabs") {
		t.Fatalf("nested tab references are unstable for path-targeted apps:\n%s", script)
	}
}

func TestMatchCheckpointSkips_ConsumesLiveHitsAndPrefersExactITermSession(t *testing.T) {
	message := "same task"
	snap := &Snapshot{Windows: []SnapshotWindow{{
		Index: 1,
		Tabs: []SnapshotTab{{Index: 1, Sessions: []SnapshotSession{
			{Index: 1, ID: "LIVE-A", Cwd: strPtr("/a"), Processes: []SnapshotProc{{PID: 11, Command: "mark " + message}}},
			{Index: 2, ID: "LIVE-B", Cwd: strPtr("/b"), Processes: []SnapshotProc{{PID: 22, Command: "mark " + message}}},
		}}},
	}}}
	doc := &SaveDocument{Windows: []SaveWindow{{Tabs: []SaveTab{
		{Kind: "mark", Message: message, ItermSessionID: "live-b"},
		{Kind: "mark", Message: message},
		{Kind: "mark", Message: message},
		{Kind: "grok", SessionID: "different", ItermSessionID: "LIVE-A"},
	}}}}

	var stderr bytes.Buffer
	skipped, count := matchCheckpointSkips(doc, indexLiveCritical(snap), &stderr)
	if count != 2 {
		t.Fatalf("skipped=%d want 2; flags=%v stderr=%s", count, skipped, stderr.String())
	}
	want := []bool{true, true, false, false}
	for i, v := range want {
		if skipped[0][i] != v {
			t.Fatalf("tab %d skipped=%v want %v; all=%v", i, skipped[0][i], v, skipped[0])
		}
	}
	lines := strings.Split(strings.TrimSpace(stderr.String()), "\n")
	if len(lines) < 1 || !strings.Contains(lines[0], "pid 22") {
		t.Fatalf("exact iTerm session should consume LIVE-B first:\n%s", stderr.String())
	}
}

func TestCaptureSnapshotAcrossApps_FindsHomeWhenBareHasNoWindows(t *testing.T) {
	release := holdTestCollector()
	t.Cleanup(release)

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	homeApp := filepath.Join(home, "Applications", "iTerm.app")
	failHome := false
	collector := defaultCollector()
	collector.ITermRunning = func() bool { return true }
	collector.RunAppleScript = func(script string) (string, error) {
		if !strings.Contains(script, homeApp) {
			return "", nil
		}
		if failHome {
			return "", fmt.Errorf("home capture unavailable")
		}
		if strings.Contains(script, "set tabCount to count of tabs") {
			return "###T###1###Default (mark)\n###S###1###/dev/ttys099###true###Default###LIVE-HOME###Default (mark)\n", nil
		}
		return "###W###1###Home###0\n", nil
	}
	collector.ListProcs = func(string) ([]rawProc, error) {
		return []rawProc{
			{PID: 100, PPID: 0, Stat: "Ss", Command: "login"},
			{PID: 101, PPID: 100, Stat: "S", Command: "-zsh"},
			{PID: 102, PPID: 101, Stat: "S+", Command: "mark home task"},
		}, nil
	}
	collector.ListCwds = func(pids []int) (map[int]string, error) {
		return map[int]string{100: "/work", 101: "/work", 102: "/work"}, nil
	}
	SetSnapshotCollectorForTest(collector)

	previousPreflight := currentMultiAppPreflightFn()
	SetMultiAppPreflightForTest(func() (MultiAppPreflight, error) {
		return MultiAppPreflight{
			AsApp:       CanonicalITermAppSystem,
			RunningApps: []string{CanonicalITermAppSystem, CanonicalITermAppHome},
		}, nil
	})
	t.Cleanup(func() { SetMultiAppPreflightForTest(previousPreflight) })

	snap, warnings, err := CaptureSnapshotAcrossApps(CaptureOpts{NoEnrich: true})
	if err != nil {
		t.Fatalf("capture: %v warnings=%v", err, warnings)
	}
	if len(snap.Windows) != 1 || snap.Windows[0].App != CanonicalITermAppHome {
		t.Fatalf("windows=%+v warnings=%v", snap.Windows, warnings)
	}
	live := indexLiveCritical(snap)
	if _, ok := live.take(SaveTab{Kind: "mark", Message: "home task", ItermSessionID: "LIVE-HOME"}); !ok {
		t.Fatalf("home-app live mark was not indexed: snap=%+v", snap)
	}

	failHome = true
	if _, _, err := CaptureSnapshotAcrossAppsStrict(CaptureOpts{NoEnrich: true}); err == nil {
		t.Fatal("strict multi-app capture must fail when the home surface cannot be scanned")
	}
}

func TestSessionsRestore_LiveScanFailureAbortsWithoutStamp(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scan-fail.json")
	doc := &SaveDocument{
		Version: sessionsSaveVersion,
		SavedAt: "2026-08-20T21:09:03+0800",
		Host:    "testhost",
		Source:  sessionsSaveSource,
		Summary: SaveSummary{Windows: 1, Tabs: 1, Sessions: 1, ByKind: map[string]int{"grok": 1}},
		Windows: []SaveWindow{{Tabs: []SaveTab{{Cwd: "/work", Kind: "grok", SessionID: "abc", ResumeCmd: "grok --resume abc"}}}},
	}
	if err := WriteSaveDocument(path, doc); err != nil {
		t.Fatal(err)
	}
	previousPath := sessionsSavePathForTest
	sessionsSavePathForTest = path
	t.Cleanup(func() { sessionsSavePathForTest = previousPath })
	InstallPhasedFixtureCollectorForTest(t, PhasedFixtureOpts{ITermRunning: false, Hostname: "testhost"})

	var stdout, stderr bytes.Buffer
	if err := runSessions([]string{"restore"}, &stdout, &stderr); err == nil {
		t.Fatalf("expected safe abort; stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "could not safely check already-running sessions") {
		t.Fatalf("stderr:\n%s", stderr.String())
	}
	got, err := ReadSaveDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.IsConsumed() {
		t.Fatal("failed live scan must not stamp restored_at")
	}
}

func TestParseLiveCriticalPanes_FailsWhenITermChangesDuringScan(t *testing.T) {
	_, err := parseLiveCriticalPanes("###E###window 2, tab 3 changed while scanning\n")
	if err == nil || !strings.Contains(err.Error(), "changed while scanning") {
		t.Fatalf("err=%v", err)
	}
}

func TestScanLiveCriticalAcrossApps_UsesOneProcessListingAndKeepsPanesInOneWindow(t *testing.T) {
	release := holdTestCollector()
	t.Cleanup(release)

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	homeApp := filepath.Join(home, "Applications", "iTerm.app")
	var appleScriptCalls, processCalls, resolveCalls int
	collector := defaultCollector()
	collector.ITermRunning = func() bool { return true }
	collector.RunAppleScript = func(script string) (string, error) {
		appleScriptCalls++
		if !strings.Contains(script, "###P###") {
			t.Fatalf("restore scan must use its single-pane AppleScript: %s", script)
		}
		if strings.Contains(script, homeApp) {
			return "###P###202###/dev/ttys003###IDLE-HOME###idle\n", nil
		}
		return strings.Join([]string{
			"###P###101###/dev/ttys001###GROK-LIVE###grok tab",
			"###P###101###/dev/ttys002###MARK-LIVE###mark tab",
			"###P###303###/dev/ttys004###UNRELATED-GROK###unrelated grok tab",
		}, "\n") + "\n", nil
	}
	collector.ListTTYProcs = func(ttys []string) ([]liveTTYProc, error) {
		processCalls++
		if got, want := strings.Join(ttys, ","), "ttys001,ttys002,ttys003,ttys004"; got != want {
			t.Fatalf("TTY filter=%q, want %q", got, want)
		}
		return []liveTTYProc{
			{rawProc: rawProc{PID: 1, PPID: 0, Stat: "Ss", Command: "login"}, TTY: "ttys001"},
			{rawProc: rawProc{PID: 2, PPID: 1, Stat: "S", Command: "-zsh"}, TTY: "ttys001"},
			{rawProc: rawProc{PID: 3, PPID: 2, Stat: "S+", Command: "grok"}, TTY: "ttys001"},
			{rawProc: rawProc{PID: 4, PPID: 0, Stat: "Ss", Command: "login"}, TTY: "ttys002"},
			{rawProc: rawProc{PID: 5, PPID: 4, Stat: "S", Command: "-zsh"}, TTY: "ttys002"},
			{rawProc: rawProc{PID: 6, PPID: 5, Stat: "S+", Command: "mark waiting for CI"}, TTY: "ttys002"},
			{rawProc: rawProc{PID: 7, PPID: 0, Stat: "Ss", Command: "login"}, TTY: "ttys004"},
			{rawProc: rawProc{PID: 8, PPID: 7, Stat: "S", Command: "-zsh"}, TTY: "ttys004"},
			{rawProc: rawProc{PID: 9, PPID: 8, Stat: "S+", Command: "grok"}, TTY: "ttys004"},
		}, nil
	}
	collector.ListProcs = func(string) ([]rawProc, error) {
		t.Fatal("restore scan must not list processes per TTY")
		return nil, nil
	}
	collector.ListCwds = func([]int) (map[int]string, error) {
		t.Fatal("restore scan must not query current directories")
		return nil, nil
	}
	collector.ResolveFromPID = func(pid int) (*procresolve.Result, error) {
		resolveCalls++
		if pid == 3 {
			return &procresolve.Result{Kind: "grok", SessionID: "session-1", RunnerPID: 3}, nil
		}
		if pid == 9 {
			t.Fatal("unrelated agent pane should not be resolved after exact match succeeds")
		}
		return nil, nil
	}
	SetSnapshotCollectorForTest(collector)

	previousPreflight := currentMultiAppPreflightFn()
	SetMultiAppPreflightForTest(func() (MultiAppPreflight, error) {
		return MultiAppPreflight{
			AsApp:       CanonicalITermAppSystem,
			RunningApps: []string{CanonicalITermAppSystem, CanonicalITermAppHome},
		}, nil
	})
	t.Cleanup(func() { SetMultiAppPreflightForTest(previousPreflight) })

	doc := &SaveDocument{Windows: []SaveWindow{{Tabs: []SaveTab{
		{Kind: "grok", SessionID: "session-1", ItermSessionID: "GROK-LIVE"},
		{Kind: "mark", Message: "waiting for CI", ItermSessionID: "MARK-LIVE"},
	}}}}
	live, warnings, err := scanLiveCriticalAcrossApps(doc, true)
	if err != nil {
		t.Fatalf("scan: %v; warnings=%v", err, warnings)
	}
	if appleScriptCalls != 2 {
		t.Fatalf("AppleScript calls=%d, want one per running installation", appleScriptCalls)
	}
	if processCalls != 1 {
		t.Fatalf("process listings=%d, want 1", processCalls)
	}
	if resolveCalls != 1 {
		t.Fatalf("agent resolutions=%d, want one candidate pane", resolveCalls)
	}
	for _, tab := range doc.Windows[0].Tabs {
		if _, ok := live.take(tab); !ok {
			t.Fatalf("missing live match for %+v", tab)
		}
	}
}

func TestFormatRestorePlan_PreservesSkippedLayoutAndColorsCommands(t *testing.T) {
	doc := &SaveDocument{
		SavedAt: "2026-08-20T21:09:03+0800",
		Windows: []SaveWindow{{Space: 13, Tabs: []SaveTab{
			{Cwd: "/work", Kind: "grok", ResumeCmd: "grok --resume abc"},
			{Cwd: "/work", Kind: "mark", ResumeCmd: "mark 'fix picture'"},
		}}},
	}
	var out bytes.Buffer
	formatRestorePlan(&out, doc, "/tmp/save.json", true, true, false,
		[][]bool{{true, true}}, 2, 0, 0,
		restorePlanOpts{Global: RestoreAppTarget{Canonical: CanonicalITermAppHome}})
	got := out.String()
	for _, want := range []string{
		"Would restore", "0 windows / 0 tabs", "saved layout shown below",
		"saved window (would not create — all tabs already running)",
		"already running — would skip", "cd", "grok", "mark",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q:\n%s", want, got)
		}
	}
	if strings.Count(got, "already running — would skip") != 2 {
		t.Fatalf("expected both skipped tabs in layout:\n%s", got)
	}
	for _, command := range []string{"cd", "grok", "mark"} {
		if !strings.Contains(got, ansiGreen+command+ansiReset) {
			t.Fatalf("command %q is not green:\n%q", command, got)
		}
	}
}
