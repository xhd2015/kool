package iterm2

import (
	"encoding/binary"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestResolveForegroundCommand(t *testing.T) {
	shell := SnapshotProc{PID: 10, PPID: 1, Stat: "Ss", Command: "-zsh"}
	launch := SnapshotProc{PID: 20, PPID: 10, Stat: "S+", Command: "make test"}
	leaf := SnapshotProc{PID: 30, PPID: 20, Stat: "S+", Command: "go test ./..."}
	for _, tt := range []struct {
		name    string
		procs   []SnapshotProc
		shell   *int
		pid     int
		complex bool
		absent  bool
	}{
		{name: "idle shell", procs: []SnapshotProc{{PID: 10, Stat: "Ss+", Command: "-zsh"}}, shell: intPtr(10), absent: true},
		{name: "background job", procs: []SnapshotProc{{PID: 10, Stat: "Ss+", Command: "-zsh"}, {PID: 20, PPID: 10, Stat: "S", Command: "server"}}, shell: intPtr(10), absent: true},
		{name: "launcher not leaf", procs: []SnapshotProc{leaf, shell, launch}, shell: intPtr(10), pid: 20},
		{name: "blocking foreground", procs: []SnapshotProc{shell, launch}, shell: intPtr(10), pid: 20},
		{name: "pipeline", procs: []SnapshotProc{shell, launch, {PID: 21, PPID: 10, Stat: "S+", Command: "tee output"}}, shell: intPtr(10), complex: true},
		{name: "shared foreground shell group", procs: []SnapshotProc{{PID: 10, Stat: "Ss+"}, launch}, shell: intPtr(10), complex: true},
		{name: "missing shell", procs: []SnapshotProc{launch}, complex: true, pid: 20},
		{name: "broken ancestry", procs: []SnapshotProc{shell, leaf}, shell: intPtr(10), complex: true, pid: 30},
		{name: "cycle", procs: []SnapshotProc{shell, {PID: 20, PPID: 30, Stat: "S+"}, {PID: 30, PPID: 20, Stat: "S+"}}, shell: intPtr(10), complex: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveForegroundCommand(SnapshotSession{ShellPID: tt.shell, Processes: tt.procs})
			if tt.absent {
				if got != nil {
					t.Fatalf("got %+v", got)
				}
				return
			}
			if got == nil || got.Complex != tt.complex || (!tt.complex && got.PID != tt.pid) {
				t.Fatalf("got %+v", got)
			}
			if got.Complex && got.Reason == "" {
				t.Fatal("missing review reason")
			}
		})
	}
}

func TestEnrichForegroundCommandReadsLauncher(t *testing.T) {
	win := SnapshotWindow{Tabs: []SnapshotTab{{Sessions: []SnapshotSession{{ShellPID: intPtr(10), Cwd: strPtr("/wrong/leaf"), Processes: []SnapshotProc{{PID: 10, Stat: "Ss"}, {PID: 20, PPID: 10, Stat: "S+", Command: "launcher display"}, {PID: 30, PPID: 20, Stat: "S+", Command: "leaf"}}}}}}}
	argv := []string{"launcher", "arg with spaces", "", "quote'\"", "line\nbreak"}
	enrichForegroundCommandsWith(&win, foregroundCommandReads{
		argv: func(pid int) ([]string, error) {
			if pid != 20 {
				t.Fatalf("argv pid %d", pid)
			}
			return argv, nil
		},
		cwd: func(pid int) (string, error) {
			if pid != 20 {
				t.Fatalf("cwd pid %d", pid)
			}
			return "/launcher cwd", nil
		},
	})
	f := win.Tabs[0].Sessions[0].Foreground
	if f.Cwd != "/launcher cwd" || !reflect.DeepEqual(f.Argv, argv) || f.Reason != "" {
		t.Fatalf("got %+v", f)
	}
	enrichForegroundCommandsWith(&win, foregroundCommandReads{argv: func(int) ([]string, error) { return nil, errors.New("gone") }, cwd: func(int) (string, error) { return "", errors.New("gone") }})
	f = win.Tabs[0].Sessions[0].Foreground
	if len(f.Argv) != 0 || f.Cwd != "" || !strings.Contains(f.Reason, "argv unavailable") || !strings.Contains(f.Reason, "cwd unavailable") {
		t.Fatalf("got %+v", f)
	}
}

func TestParseForegroundProcArgs(t *testing.T) {
	argv := []string{"/bin/tool", "a b", "", "single' double\"", "line\nbreak"}
	data := make([]byte, 4)
	binary.NativeEndian.PutUint32(data, uint32(len(argv)))
	data = append(data, []byte("/bin/tool\x00\x00\x00"+strings.Join(argv, "\x00")+"\x00SECRET=never expose\x00")...)
	got, err := parseForegroundProcArgs(data)
	if err != nil || !reflect.DeepEqual(got, argv) {
		t.Fatalf("got %q, %v", got, err)
	}
	for _, bad := range [][]byte{nil, {1, 0, 0}, {0, 0, 0, 0}, data[:8], data[:len(data)-len("SECRET=never expose\x00")-1]} {
		if got, err := parseForegroundProcArgs(bad); err == nil {
			t.Fatalf("accepted truncated buffer %q", got)
		}
	}
}

func TestForegroundCommandFixtureEvidence(t *testing.T) {
	evidence := &ForegroundCommand{PID: 42, Cwd: "/fixture", Argv: []string{"tool", "a b"}}
	InstallPhasedFixtureCollectorForTest(t, PhasedFixtureOpts{ITermRunning: true, Windows: []SnapshotWindow{{Index: 1, Tabs: []SnapshotTab{{Index: 1, Sessions: []SnapshotSession{{Index: 1, ID: "fixture-session", TTY: "/dev/ttys987", Foreground: evidence}}}}}}})
	snap, _, err := activeCollector().CaptureWith(CaptureOpts{NoEnrich: true})
	if err != nil {
		t.Fatal(err)
	}
	got := snap.Windows[0].Tabs[0].Sessions[0].Foreground
	if !reflect.DeepEqual(got, evidence) {
		t.Fatalf("got %+v", got)
	}
	got.Argv[0] = "changed"
	if evidence.Argv[0] != "tool" {
		t.Fatal("fixture evidence was aliased")
	}
}
