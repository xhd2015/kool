package iterm2

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestEnrichIdleShell(t *testing.T) {
	t.Parallel()
	procs := []rawProc{
		{PID: 1, PPID: 0, Stat: "Ss", Etime: "1:00", Command: "login -fp u"},
		{PID: 2, PPID: 1, Stat: "S+", Etime: "0:59", Command: "-bash"},
	}
	idle, shell, chosen, _, _ := enrichFromProcs(procs, nil, time.Now())
	if idle == nil || !*idle {
		t.Fatalf("want idle, got %#v", idle)
	}
	if shell == nil || *shell != 2 {
		t.Fatalf("shell pid %#v", shell)
	}
	if chosen == nil || chosen.PID != 2 {
		t.Fatalf("chosen %#v", chosen)
	}
}

func TestEnrichBusyChild(t *testing.T) {
	t.Parallel()
	procs := []rawProc{
		{PID: 1, PPID: 0, Stat: "Ss", Etime: "1:00", Command: "login -fp u"},
		{PID: 2, PPID: 1, Stat: "S", Etime: "0:59", Command: "-bash"},
		{PID: 3, PPID: 2, Stat: "S+", Etime: "0:50", Command: "grok --always-approve"},
	}
	idle, shell, chosen, _, _ := enrichFromProcs(procs, map[int]string{2: "/tmp/x", 3: "/tmp/x"}, time.Now())
	if idle == nil || *idle {
		t.Fatalf("want busy, got %#v", idle)
	}
	if shell == nil || *shell != 2 {
		t.Fatalf("shell %#v", shell)
	}
	if chosen == nil || chosen.PID != 3 {
		t.Fatalf("chosen %#v", chosen)
	}
}

func TestFindSessionsByRef(t *testing.T) {
	t.Parallel()
	pid := 94616
	shell := 100
	snap := &Snapshot{
		Windows: []SnapshotWindow{{
			Index: 1, Name: "Win A",
			Tabs: []SnapshotTab{
				{
					Index: 1, Name: "Tab A",
					Sessions: []SnapshotSession{{
						Index: 1, ID: "D922B298-25FB-41FA-BAF8-7AC7A1D56758",
						Name: "grok", TTY: "/dev/ttys003", PID: &pid,
					}},
				},
				{
					Index: 2, Name: "idle",
					Sessions: []SnapshotSession{{
						Index: 1, ID: "AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE",
						Name: "-bash", TTY: "/dev/ttys004", ShellPID: &shell,
					}},
				},
			},
		}},
	}
	if m := FindSessionsByRef(snap, "D922B298"); len(m) != 1 {
		t.Fatalf("prefix match %d", len(m))
	}
	if m := FindSessionsByRef(snap, "ttys003"); len(m) != 1 {
		t.Fatalf("tty match %d", len(m))
	}
	if m := FindSessionsByRef(snap, "94616"); len(m) != 1 {
		t.Fatalf("pid match %d", len(m))
	}
	if m := FindSessionsByRef(snap, "nope"); len(m) != 0 {
		t.Fatal("expected none")
	}
}

func TestCaptureWithFakeCollector(t *testing.T) {
	InstallPhasedFixtureCollectorForTest(t, PhasedFixtureOpts{
		ITermRunning: true,
		Windows: []SnapshotWindow{
			{
				Index: 1,
				Name:  "W",
				Tabs: []SnapshotTab{
					{
						Index: 1,
						Name:  "T",
						Sessions: []SnapshotSession{
							{
								Index: 1, ID: "ABCD1234-0000-0000-0000-000000000001",
								Name: "busy-sess", TTY: "/dev/ttys001", Profile: "Default",
							},
							{
								Index: 2, ID: "ABCD1234-0000-0000-0000-000000000002",
								Name: "idle-sess", TTY: "/dev/ttys002", Profile: "Default",
							},
						},
					},
				},
			},
		},
		BusyTTYs: []string{"ttys001"},
		IdleTTYs: []string{"ttys002"},
		BusyLeafByTTY: map[string]string{
			"ttys001": "spl seatalk --app-secret=TOPSECRET",
		},
		CwdByTTY: map[string]string{
			"ttys001": "/tmp/a",
			"ttys002": "/tmp/b",
		},
		Now:      time.Date(2026, 7, 25, 12, 0, 0, 0, time.Local),
		Hostname: "testhost",
	})

	snap, _, err := CaptureSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	if snap.Summary.Sessions != 2 || snap.Summary.Busy != 1 || snap.Summary.Idle != 1 {
		t.Fatalf("summary %+v", snap.Summary)
	}
	busy := snap.Windows[0].Tabs[0].Sessions[0]
	if busy.Idle == nil || *busy.Idle {
		t.Fatal("first should be busy")
	}
	if busy.CommandLine == nil || strings.Contains(*busy.CommandLine, "TOPSECRET") {
		t.Fatalf("secret not redacted: %#v", busy.CommandLine)
	}
	if busy.CommandLine == nil || !strings.Contains(*busy.CommandLine, "***") {
		t.Fatalf("expected redaction: %#v", busy.CommandLine)
	}

	var b bytes.Buffer
	if err := RenderSnapshot(&b, snap, RenderOptions{Format: FormatJSON, NoColor: true}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), `"id": "ABCD1234-0000-0000-0000-000000000001"`) {
		t.Fatalf("json missing id: %s", b.String()[:200])
	}
	b.Reset()
	if err := RenderSnapshot(&b, snap, RenderOptions{Format: FormatCLI, NoColor: true}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "busy") || !strings.Contains(b.String(), "idle") {
		t.Fatal(b.String())
	}
	b.Reset()
	if err := RenderSnapshot(&b, snap, RenderOptions{Format: FormatMarkdown, NoColor: true}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "# iTerm2 snapshot") {
		t.Fatal(b.String())
	}
	b.Reset()
	if err := RenderSnapshot(&b, snap, RenderOptions{Format: FormatHTML, NoColor: true}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "<html") {
		t.Fatal(b.String())
	}
}
