package iterm2

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func installForkGrokFixture(t *testing.T, cwd string) {
	t.Helper()
	InstallPhasedFixtureCollectorForTest(t, PhasedFixtureOpts{
		ITermRunning: true,
		Windows: []SnapshotWindow{{
			Index: 1,
			Tabs: []SnapshotTab{{
				Index: 1,
				Sessions: []SnapshotSession{{
					Index: 1, ID: "D922B298-25FB-41FA-BAF8-7AC7A1D56758",
					Name: "g", TTY: "/dev/ttys020", Profile: "Default",
				}},
			}},
		}},
		BusyTTYs: []string{"ttys020"},
		CwdByTTY: map[string]string{"ttys020": cwd},
		AgentResolveByTTY: map[string]AgentResolveFixture{
			"ttys020": {Kind: "grok", SessionID: "019fabcdef-1234-5678-9abc-def012345678"},
		},
		Now:      time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
		Hostname: "testhost",
	})
}

func TestBuildForkFollowUpCommand(t *testing.T) {
	got := BuildForkFollowUpCommand("agent-run", "/tmp/proj", "sess-1")
	for _, want := range []string{
		"agent-run",
		"run",
		"--agent-runner=grok-tty",
		"--dir=/tmp/proj",
		"--resume-from-grok-session=sess-1",
		"--fork",
		"--open",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
	if strings.Contains(got, "--new-terminal") {
		t.Fatal("must not include --new-terminal")
	}
	if strings.Contains(got, "--auto-send-or-resume") {
		t.Fatal("must use resume-from-grok-session, not auto-send")
	}
	if strings.Contains(got, "/fork") {
		t.Fatal("must use agent-run --fork, not inject /fork")
	}
}

func TestBuildForkPlan_RequiresGrokAndCwd(t *testing.T) {
	_, err := BuildForkPlan(&SnapshotSession{
		ID: "x", Agent: &SessionAgent{Kind: "codex", SessionID: "c"},
		Cwd: strPtr("/tmp"),
	}, "agent-run")
	if err == nil || !strings.Contains(err.Error(), "not a grok") {
		t.Fatalf("got %v", err)
	}

	_, err = BuildForkPlan(&SnapshotSession{
		ID: "y", Agent: &SessionAgent{Kind: "grok", SessionID: "g"},
	}, "agent-run")
	if err == nil || !strings.Contains(err.Error(), "empty cwd") {
		t.Fatalf("got %v", err)
	}

	plan, err := BuildForkPlan(&SnapshotSession{
		ID: "D922B298-25FB-41FA-BAF8-7AC7A1D56758",
		Agent: &SessionAgent{Kind: "grok", SessionID: "g-uuid"},
		Cwd:   strPtr("/tmp"),
	}, "/usr/local/bin/agent-run")
	if err != nil {
		t.Fatal(err)
	}
	if plan.GrokSessionID != "g-uuid" {
		t.Fatal(plan.GrokSessionID)
	}
	if !strings.Contains(plan.FollowUp, "--resume-from-grok-session=g-uuid") {
		t.Fatal(plan.FollowUp)
	}
	if !strings.Contains(plan.FollowUp, "/usr/local/bin/agent-run") {
		t.Fatal(plan.FollowUp)
	}
}

func TestSessionForkHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := runSession([]string{"D922B298", "fork", "-h"}, &stdout, &stderr, TestRun{}); err != nil {
		t.Fatal(err)
	}
	out := stdout.String()
	for _, want := range []string{"fork", "--dry-run", "agent-run", "--fork-session"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q:\n%s", want, out)
		}
	}
}

func TestSessionForkDryRun(t *testing.T) {
	cwd := t.TempDir()
	installForkGrokFixture(t, cwd)

	// Dry-run tolerates missing agent-run on PATH.
	prevLook := sessionLookPath
	sessionLookPath = func(file string) (string, error) {
		return "", os.ErrNotExist
	}
	t.Cleanup(func() { sessionLookPath = prevLook })

	var stdout, stderr bytes.Buffer
	if err := runSession([]string{"D922B298", "fork", "--dry-run"}, &stdout, &stderr, TestRun{}); err != nil {
		t.Fatalf("%v stderr=%s", err, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "Would open new window") {
		t.Fatal(out)
	}
	if !strings.Contains(out, "019fabcdef-1234-5678-9abc-def012345678") {
		t.Fatal(out)
	}
	if !strings.Contains(out, "--resume-from-grok-session=") {
		t.Fatal(out)
	}
	if !strings.Contains(out, "--fork") {
		t.Fatal(out)
	}
	if strings.Contains(out, "/fork") {
		t.Fatal("should not inject /fork:", out)
	}
	if !strings.Contains(out, "agent-run") {
		t.Fatal(out)
	}
	// Realpath of temp cwd should appear.
	if real, err := filepath.EvalSymlinks(cwd); err == nil {
		if !strings.Contains(out, real) && !strings.Contains(out, cwd) {
			t.Fatalf("cwd missing:\n%s", out)
		}
	}
}

func TestSessionForkLiveInject(t *testing.T) {
	cwd := t.TempDir()
	installForkGrokFixture(t, cwd)

	prevLook := sessionLookPath
	sessionLookPath = func(file string) (string, error) {
		return "/fake/bin/agent-run", nil
	}
	t.Cleanup(func() { sessionLookPath = prevLook })

	var gotDir, gotFollow string
	prevOpen := sessionOpenInNewTerminal
	sessionOpenInNewTerminal = func(dir, followUp string) error {
		gotDir, gotFollow = dir, followUp
		return nil
	}
	t.Cleanup(func() { sessionOpenInNewTerminal = prevOpen })

	var stdout, stderr bytes.Buffer
	if err := runSession([]string{"D922B298", "fork"}, &stdout, &stderr, TestRun{}); err != nil {
		t.Fatalf("%v stderr=%s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Opened new window") {
		t.Fatal(stdout.String())
	}
	if !strings.Contains(gotFollow, "--resume-from-grok-session=019fabcdef-1234-5678-9abc-def012345678") {
		t.Fatalf("follow-up: %s", gotFollow)
	}
	if !strings.Contains(gotFollow, "--fork") || !strings.Contains(gotFollow, "--open") {
		t.Fatalf("follow-up: %s", gotFollow)
	}
	if strings.Contains(gotFollow, "/fork") {
		t.Fatalf("must not inject /fork: %s", gotFollow)
	}
	if !strings.Contains(gotFollow, "/fake/bin/agent-run") {
		t.Fatalf("follow-up: %s", gotFollow)
	}
	// dir should be normalized form of cwd
	if gotDir == "" {
		t.Fatal("empty dir")
	}
}

func TestSessionForkNotGrok(t *testing.T) {
	cwd := t.TempDir()
	InstallPhasedFixtureCollectorForTest(t, PhasedFixtureOpts{
		ITermRunning: true,
		Windows: []SnapshotWindow{{
			Index: 1,
			Tabs: []SnapshotTab{{
				Index: 1,
				Sessions: []SnapshotSession{{
					Index: 1, ID: "AAAAAAAA-0000-0000-0000-000000000001",
					TTY: "/dev/ttys030", Profile: "Default",
				}},
			}},
		}},
		BusyTTYs: []string{"ttys030"},
		CwdByTTY: map[string]string{"ttys030": cwd},
		AgentResolveByTTY: map[string]AgentResolveFixture{
			"ttys030": {Kind: "codex", SessionID: "c1"},
		},
		Hostname: "h",
	})
	prevLook := sessionLookPath
	sessionLookPath = func(string) (string, error) { return "/bin/agent-run", nil }
	t.Cleanup(func() { sessionLookPath = prevLook })

	var stdout, stderr bytes.Buffer
	err := runSession([]string{"AAAAAAAA", "fork"}, &stdout, &stderr, TestRun{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr.String(), "not a grok") {
		t.Fatal(stderr.String())
	}
}

func TestSessionForkMissingAgentRun(t *testing.T) {
	cwd := t.TempDir()
	installForkGrokFixture(t, cwd)
	prevLook := sessionLookPath
	sessionLookPath = func(string) (string, error) { return "", os.ErrNotExist }
	t.Cleanup(func() { sessionLookPath = prevLook })

	var stdout, stderr bytes.Buffer
	err := runSession([]string{"D922B298", "fork"}, &stdout, &stderr, TestRun{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr.String(), "agent-run not found") {
		t.Fatal(stderr.String())
	}
}

func TestSessionHelpMentionsListAndFork(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := runSession([]string{"-h"}, &stdout, &stderr, TestRun{}); err != nil {
		t.Fatal(err)
	}
	out := stdout.String()
	for _, want := range []string{"list", "fork", "status", "--grok", "--only-cwd"} {
		if !strings.Contains(out, want) {
			t.Fatalf("session help missing %q:\n%s", want, out)
		}
	}
}
