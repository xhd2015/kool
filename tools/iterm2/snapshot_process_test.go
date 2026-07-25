package iterm2

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestEtimeToSeconds(t *testing.T) {
	t.Parallel()
	if etimeToSeconds("05:06") != 5*60+6 {
		t.Fatal(etimeToSeconds("05:06"))
	}
	if etimeToSeconds("01:02:03") != 3600+120+3 {
		t.Fatal(etimeToSeconds("01:02:03"))
	}
	if etimeToSeconds("1-02:03:04") != 86400+2*3600+3*60+4 {
		t.Fatal(etimeToSeconds("1-02:03:04"))
	}
}

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

func TestParseHierarchyAndFind(t *testing.T) {
	t.Parallel()
	raw := `###W###1###Win A
###T###1###Tab A
###S###1###/dev/ttys003###false###Default###D922B298-25FB-41FA-BAF8-7AC7A1D56758###grok (grok)
###T###2###idle
###S###1###/dev/ttys004###false###Default###AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE###-bash
`
	wins, warns := parseHierarchy(raw)
	if len(warns) != 0 {
		t.Fatalf("warnings: %v", warns)
	}
	if len(wins) != 1 || len(wins[0].Tabs) != 2 {
		t.Fatalf("%+v", wins)
	}
	s0 := wins[0].Tabs[0].Sessions[0]
	if s0.ID != "D922B298-25FB-41FA-BAF8-7AC7A1D56758" {
		t.Fatal(s0.ID)
	}

	snap := &Snapshot{Windows: wins}
	// enrich ids only for find tests — set pid
	snap.Windows[0].Tabs[0].Sessions[0].PID = intPtr(94616)
	snap.Windows[0].Tabs[1].Sessions[0].ShellPID = intPtr(100)

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
	c := &SnapshotCollector{
		ITermRunning:   func() bool { return true },
		fixtureEnabled: true,
		fixtureWindows: []SnapshotWindow{
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
		ListProcs: func(ttyShort string) ([]rawProc, error) {
			if ttyShort == "ttys001" {
				return []rawProc{
					{PID: 10, PPID: 1, Stat: "Ss", Etime: "1:00", Lstart: "Sat Jul 25 10:00:00 2026", Command: "login -fp u"},
					{PID: 11, PPID: 10, Stat: "S", Etime: "0:59", Lstart: "Sat Jul 25 10:00:01 2026", Command: "-bash"},
					{PID: 12, PPID: 11, Stat: "S+", Etime: "0:30", Lstart: "Sat Jul 25 10:00:30 2026", Command: "spl seatalk --app-secret=TOPSECRET"},
				}, nil
			}
			return []rawProc{
				{PID: 20, PPID: 1, Stat: "Ss", Etime: "2:00", Command: "login -fp u"},
				{PID: 21, PPID: 20, Stat: "S+", Etime: "1:59", Command: "-bash"},
			}, nil
		},
		ListCwds: func(pids []int) (map[int]string, error) {
			return map[int]string{11: "/tmp/a", 12: "/tmp/a", 21: "/tmp/b"}, nil
		},
		Now:      func() time.Time { return time.Date(2026, 7, 25, 12, 0, 0, 0, time.Local) },
		Hostname: func() (string, error) { return "testhost", nil },
	}
	release := holdTestCollector()
	t.Cleanup(release)
	SetSnapshotCollectorForTest(c)

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

	// render smoke
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
