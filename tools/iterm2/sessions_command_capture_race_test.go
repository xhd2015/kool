package iterm2

import (
	"fmt"
	"strings"
	"testing"
)

func TestForegroundEvidenceRaceGuard(t *testing.T) {
	for _, mode := range []string{"stable", "pid reused", "exec changed", "background now", "start changed before read", "identity unavailable"} {
		t.Run(mode, func(t *testing.T) {
			win := SnapshotWindow{Tabs: []SnapshotTab{{Sessions: []SnapshotSession{{ShellPID: intPtr(10), Processes: []SnapshotProc{{PID: 10, Stat: "Ss"}, {PID: 20, PPID: 10, Stat: "S+", StartTimeUnix: int64Ptr(100), Command: "vite --host"}}}}}}}
			identities, args := 0, 0
			reads := foregroundCommandReads{
				identity: func(pid int) (foregroundProcessIdentity, error) {
					identities++
					id := foregroundProcessIdentity{PID: pid, PPID: 10, PGID: 20, TPGID: 20, StartSec: 100, StartUsec: 1}
					switch mode {
					case "pid reused":
						if identities > 1 {
							id.StartUsec++
						}
					case "background now":
						id.TPGID = 10
					case "start changed before read":
						id.StartSec = 200
					case "identity unavailable":
						return id, fmt.Errorf("gone")
					}
					return id, nil
				},
				argv: func(int) ([]string, error) {
					args++
					if mode == "exec changed" && args > 1 {
						return []string{"rm", "file"}, nil
					}
					return []string{"vite", "--host"}, nil
				},
				cwd: func(int) (string, error) { return "/work", nil },
			}
			enrichForegroundCommandsWith(&win, reads)
			f := win.Tabs[0].Sessions[0].Foreground
			if mode == "stable" {
				if f.Reason != "" || len(f.Argv) != 2 || f.Cwd != "/work" {
					t.Fatalf("got %+v", f)
				}
				return
			}
			if f == nil || !f.Complex || f.Reason == "" || len(f.Argv) != 0 || f.Cwd != "" {
				t.Fatalf("unsafe evidence %+v", f)
			}
		})
	}
}

func TestForegroundEvidenceSkipsExistingKinds(t *testing.T) {
	for _, kind := range []string{"grok", "codex", "mark"} {
		t.Run(kind, func(t *testing.T) {
			s := SnapshotSession{ShellPID: intPtr(10), Processes: []SnapshotProc{{PID: 10, Stat: "Ss"}, {PID: 20, PPID: 10, Stat: "S+", Command: kind + " hello"}}}
			if kind != "mark" {
				s.Agent = &SessionAgent{Kind: kind, SessionID: "existing"}
			}
			win := SnapshotWindow{Tabs: []SnapshotTab{{Sessions: []SnapshotSession{s}}}}
			// Nil platform readers panic if called: existing critical identities need no
			// generic command collection, even when their processes have foreground '+'.
			enrichForegroundCommandsWith(&win, foregroundCommandReads{})
			if win.Tabs[0].Sessions[0].Foreground != nil {
				t.Fatal("unexpected generic evidence")
			}
		})
	}
}

func TestForegroundEvidenceSkipsNestedIdleShell(t *testing.T) {
	for _, argv := range [][]string{{"-zsh"}, {"/bin/bash", "-il"}} {
		win := SnapshotWindow{Tabs: []SnapshotTab{{Sessions: []SnapshotSession{{ShellPID: intPtr(10), Processes: []SnapshotProc{{PID: 10, Stat: "Ss"}, {PID: 20, PPID: 10, Stat: "S+", Command: "shell"}}}}}}}
		enrichForegroundCommandsWith(&win, foregroundCommandReads{argv: func(int) ([]string, error) { return argv, nil }})
		if win.Tabs[0].Sessions[0].Foreground != nil {
			t.Fatal("idle nested shell should be excluded")
		}
	}
	if foregroundBareShell([]string{"bash", "-c", "serve"}) || foregroundBareShell([]string{"bash", "script.sh"}) {
		t.Fatal("shell commands must retain review evidence")
	}
}

func TestForegroundCommandStrictLiveScanFixture(t *testing.T) {
	fg := &ForegroundCommand{PID: 42, Cwd: "/fixture cwd", Argv: []string{"vite", "--host", "127.0.0.1"}}
	InstallPhasedFixtureCollectorForTest(t, PhasedFixtureOpts{ITermRunning: true, Windows: []SnapshotWindow{{Index: 1, Tabs: []SnapshotTab{{Index: 1, Sessions: []SnapshotSession{{Index: 1, ID: "LIVE-COMMAND", TTY: "/dev/ttys987", Foreground: fg}}}}}}})
	snap, _, err := CaptureSnapshotAcrossAppsStrict(CaptureOpts{})
	if err != nil {
		t.Fatal(err)
	}
	pane := snap.Windows[0].Tabs[0].Sessions[0]
	if pane.Foreground == nil || pane.Foreground.Cwd != fg.Cwd {
		t.Fatalf("foreground lost through strict capture: %+v", pane.Foreground)
	}
	tab, warning := classifyCriticalTab(snap.Windows[0], snap.Windows[0].Tabs[0], pane)
	if tab == nil || tab.Kind != "command" {
		t.Fatalf("classify got %+v warning %s", tab, warning)
	}
	doc := &SaveDocument{Windows: []SaveWindow{{Tabs: []SaveTab{*tab}}}}
	idx, warnings, err := scanLiveCriticalAcrossApps(doc, true)
	if err != nil {
		t.Fatalf("scan: %v warnings=%v", err, warnings)
	}
	if _, ok := idx.take(*tab); !ok {
		t.Fatal("live scan did not find exact command evidence")
	}
	if _, ok := idx.take(*tab); ok {
		t.Fatal("one live pane consumed twice")
	}
	if !strings.HasPrefix(criticalMatchKey(*tab), "command:") {
		t.Fatal("fixture did not exercise generic-command scan")
	}
}
