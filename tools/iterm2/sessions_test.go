package iterm2

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	lib "github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
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
	if err := runSession([]string{"11111111", "status", "--no-color"}, &stdout, &stderr, TestRun{}); err != nil {
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
	err := runSession([]string{"deadbeef", "status"}, &stdout, &stderr, TestRun{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr.String(), "not found") {
		t.Fatal(stderr.String())
	}
}

func TestSessionHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := runSession([]string{"-h"}, &stdout, &stderr, TestRun{}); err != nil {
		t.Fatal(err)
	}
	out := stdout.String()
	for _, want := range []string{"status", "send", "--no-submit", "--no-ctrl-u", "--focus"} {
		if !strings.Contains(out, want) {
			t.Fatalf("help missing %q:\n%s", want, out)
		}
	}
}

func sendListFixture() []lib.SessionRef {
	return []lib.SessionRef{
		{WindowID: "1", TabIndex: 1, SessionID: "11111111-2222-3333-4444-555555555555", TTY: "/dev/ttys010", Name: "a"},
		{WindowID: "1", TabIndex: 2, SessionID: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", TTY: "/dev/ttys011", Name: "b"},
	}
}

func TestSessionSendSuccess(t *testing.T) {
	var gotID, gotText string
	var gotOpts lib.SendTextOptions
	var listed bool
	var stdout, stderr bytes.Buffer
	err := runSession([]string{"11111111", "send", "--no-submit", "--no-ctrl-u", "echo hi"}, &stdout, &stderr, TestRun{
		ListSessions: func() ([]lib.SessionRef, error) {
			listed = true
			return sendListFixture(), nil
		},
		SendText: func(sessionID, text string, opts lib.SendTextOptions, cfg *lib.SendTextConfig) error {
			gotID, gotText, gotOpts = sessionID, text, opts
			return nil
		},
	})
	if err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	if !listed {
		t.Fatal("prefix resolve should use ListSessions")
	}
	if gotID != "11111111-2222-3333-4444-555555555555" {
		t.Fatalf("id=%q", gotID)
	}
	if gotText != "echo hi" {
		t.Fatalf("text=%q", gotText)
	}
	if !gotOpts.NoSubmit || !gotOpts.NoCtrlU || gotOpts.Focus {
		t.Fatalf("opts=%+v", gotOpts)
	}
	if stdout.String() != "sent to session 11111111\n" {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestSessionSendFullUUIDSkipsList(t *testing.T) {
	var gotID string
	var listed bool
	var stdout, stderr bytes.Buffer
	full := "11111111-2222-3333-4444-555555555555"
	err := runSession([]string{full, "send", "--focus", "ls"}, &stdout, &stderr, TestRun{
		ListSessions: func() ([]lib.SessionRef, error) {
			listed = true
			return sendListFixture(), nil
		},
		SendText: func(sessionID, text string, opts lib.SendTextOptions, cfg *lib.SendTextConfig) error {
			gotID = sessionID
			if !opts.Focus || opts.NoSubmit || opts.NoCtrlU {
				t.Fatalf("opts=%+v", opts)
			}
			return nil
		},
	})
	if err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	if listed {
		t.Fatal("full UUID must skip ListSessions")
	}
	if gotID != full {
		t.Fatalf("id=%q", gotID)
	}
	if stdout.String() != "sent to session "+full+"\n" {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestSessionSendMissingText(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := runSession([]string{"11111111", "send"}, &stdout, &stderr, TestRun{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr.String(), "missing text") {
		t.Fatal(stderr.String())
	}
}

func TestSessionSendNotFound(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := runSession([]string{"deadbeef", "send", "hi"}, &stdout, &stderr, TestRun{
		ListSessions: func() ([]lib.SessionRef, error) { return sendListFixture(), nil },
		SendText: func(sessionID, text string, opts lib.SendTextOptions, cfg *lib.SendTextConfig) error {
			t.Fatal("SendText should not be called")
			return nil
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr.String(), "not found") {
		t.Fatal(stderr.String())
	}
}

func sendTabStatusEnv(currentSessionID string) TestRun {
	refs := sendListFixture()
	return TestRun{
		CurrentStatus: &lib.CurrentStatusConfig{
			SessionID:      func() string { return currentSessionID },
			ListSessions:   func() ([]lib.SessionRef, error) { return refs, nil },
			ControllingTTY: func() string { return "" },
			AncestorTTYs:   func() []string { return nil },
		},
	}
}

func TestSessionSendFlags_TabNext(t *testing.T) {
	var gotID, gotText string
	env := sendTabStatusEnv("11111111-2222-3333-4444-555555555555")
	env.SendText = func(sessionID, text string, opts lib.SendTextOptions, cfg *lib.SendTextConfig) error {
		gotID, gotText = sessionID, text
		if opts.Focus || opts.NoSubmit || opts.NoCtrlU {
			t.Fatalf("opts=%+v", opts)
		}
		return nil
	}
	var stdout, stderr bytes.Buffer
	err := runSession([]string{"send", "--tab", "next", "from-next"}, &stdout, &stderr, env)
	if err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	want := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	if gotID != want || gotText != "from-next" {
		t.Fatalf("id=%q text=%q", gotID, gotText)
	}
	if stdout.String() != "sent to session "+want+"\n" {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestSessionSendFlags_TabIndex(t *testing.T) {
	var gotID string
	env := sendTabStatusEnv("11111111-2222-3333-4444-555555555555")
	env.SendText = func(sessionID, text string, opts lib.SendTextOptions, cfg *lib.SendTextConfig) error {
		gotID = sessionID
		return nil
	}
	var stdout, stderr bytes.Buffer
	err := runSession([]string{"send", "--tab-index", "0", "hi"}, &stdout, &stderr, env)
	if err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	if gotID != "11111111-2222-3333-4444-555555555555" {
		t.Fatalf("id=%q", gotID)
	}
}

func TestSessionSendFlags_SessionID(t *testing.T) {
	var gotID string
	var listed bool
	var stdout, stderr bytes.Buffer
	err := runSession([]string{"send", "--session-id", "aaaaaaaa", "--no-submit", "hi"}, &stdout, &stderr, TestRun{
		ListSessions: func() ([]lib.SessionRef, error) {
			listed = true
			return sendListFixture(), nil
		},
		SendText: func(sessionID, text string, opts lib.SendTextOptions, cfg *lib.SendTextConfig) error {
			gotID = sessionID
			if !opts.NoSubmit {
				t.Fatalf("opts=%+v", opts)
			}
			return nil
		},
	})
	if err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	if !listed {
		t.Fatal("prefix resolve should list")
	}
	if gotID != "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee" {
		t.Fatalf("id=%q", gotID)
	}
	if stdout.String() != "sent to session aaaaaaaa\n" {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestSessionSendFlags_MissingSource(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := runSession([]string{"send", "hi"}, &stdout, &stderr, TestRun{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr.String(), "expected --session-id, or --tab / --tab-index") {
		t.Fatal(stderr.String())
	}
}

func TestSessionSendFlags_ConflictTabAndSessionID(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := runSession([]string{"send", "--tab", "next", "--session-id", "abc", "hi"}, &stdout, &stderr, TestRun{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr.String(), "--session-id cannot be combined") {
		t.Fatal(stderr.String())
	}
}

func TestSessionSendFlags_ConflictTabAndTabIndex(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := runSession([]string{"send", "--tab", "next", "--tab-index", "0", "hi"}, &stdout, &stderr, TestRun{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr.String(), "--tab and --tab-index cannot be specified together") {
		t.Fatal(stderr.String())
	}
}

func TestSessionSendFlags_TabNextAtLast(t *testing.T) {
	env := sendTabStatusEnv("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	env.SendText = func(sessionID, text string, opts lib.SendTextOptions, cfg *lib.SendTextConfig) error {
		t.Fatal("SendText should not run")
		return nil
	}
	var stdout, stderr bytes.Buffer
	err := runSession([]string{"send", "--tab", "next", "hi"}, &stdout, &stderr, env)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr.String(), "no tab to the right") {
		t.Fatal(stderr.String())
	}
}

func TestSessionSendPositional_RejectsTabFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := runSession([]string{"11111111", "send", "--tab", "next", "hi"}, &stdout, &stderr, TestRun{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr.String(), "belong on") {
		t.Fatal(stderr.String())
	}
}

func TestSessionSendFlags_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := runSession([]string{"send", "-h"}, &stdout, &stderr, TestRun{})
	if err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{"--tab", "--tab-index", "--session-id", "session send"} {
		if !strings.Contains(out, want) {
			t.Fatalf("help missing %q:\n%s", want, out)
		}
	}
}
