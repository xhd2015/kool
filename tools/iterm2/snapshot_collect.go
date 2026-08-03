package iterm2

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
)

// SnapshotCollector gathers hierarchy + process enrichment. Fields may be
// overridden in tests.
type SnapshotCollector struct {
	// RunAppleScript runs an AppleScript body and returns stdout.
	RunAppleScript func(script string) (string, error)
	// ListProcs returns processes on a short tty name (e.g. "ttys003").
	ListProcs func(ttyShort string) ([]rawProc, error)
	// ListCwds returns cwd paths keyed by pid for the given pids.
	ListCwds func(pids []int) (map[int]string, error)
	// ITermRunning reports whether iTerm2 appears to be running.
	ITermRunning func() bool
	// Now is the clock (defaults to time.Now).
	Now func() time.Time
	// Hostname defaults to os.Hostname.
	Hostname func() (string, error)

	// OnListWindows is an optional test hook invoked at the start of ListWindows.
	OnListWindows func()
	// OnListTabs is an optional test hook invoked at the start of ListTabsAndSessions.
	OnListTabs func(windowIndex int)

	// ResolveFromPID optionally overrides live procresolve (production default).
	ResolveFromPID func(pid int) (*procresolve.Result, error)

	// fixtureEnabled + fixtureWindows drive phased APIs without AppleScript.
	fixtureEnabled bool
	fixtureWindows []SnapshotWindow
	// agentResolveByTTY injects procresolve results by short tty (tests).
	agentResolveByTTY map[string]AgentResolveFixture
}

func defaultCollector() *SnapshotCollector {
	return &SnapshotCollector{
		RunAppleScript: defaultRunAppleScript,
		ListProcs:      defaultListProcs,
		ListCwds:       defaultListCwds,
		ITermRunning:   defaultITermRunning,
		Now:            time.Now,
		Hostname: func() (string, error) {
			h, err := os.Hostname()
			if err != nil {
				return "", err
			}
			if i := strings.Index(h, "."); i > 0 {
				return h[:i], nil
			}
			return h, nil
		},
	}
}

var (
	testCollector   *SnapshotCollector
	testCollectorMu sync.Mutex
	// testCollectorHold serializes tests that inject a collector so parallel
	// doctest leaves do not stomp each other.
	testCollectorHold sync.Mutex
)

// SetSnapshotCollectorForTest overrides the collector used by CaptureSnapshot.
// Pass nil to restore production defaults. Tests should t.Cleanup restore.
// Prefer InstallPhasedFixtureCollectorForTest for parallel-safe inject.
func SetSnapshotCollectorForTest(c *SnapshotCollector) {
	testCollectorMu.Lock()
	defer testCollectorMu.Unlock()
	testCollector = c
}

// ActiveSnapshotCollectorForTest returns the collector CaptureSnapshot will use.
func ActiveSnapshotCollectorForTest() *SnapshotCollector {
	return activeCollector()
}

func activeCollector() *SnapshotCollector {
	testCollectorMu.Lock()
	defer testCollectorMu.Unlock()
	if testCollector != nil {
		return testCollector
	}
	return defaultCollector()
}

// holdTestCollector acquires exclusive inject ownership until release is called.
// Used so parallel tests share one global inject slot safely.
func holdTestCollector() (release func()) {
	testCollectorHold.Lock()
	var once sync.Once
	return func() {
		once.Do(func() {
			SetSnapshotCollectorForTest(nil)
			testCollectorHold.Unlock()
		})
	}
}

// PhasedFixtureOpts configures InstallPhasedFixtureCollectorForTest.
type PhasedFixtureOpts struct {
	Windows       []SnapshotWindow
	ITermRunning  bool
	OnListWindows func()
	OnListTabs    func(windowIndex int)
	// IdleTTYs are short tty names (e.g. "ttys001") classified idle (shell only).
	IdleTTYs []string
	// BusyTTYs are short tty names classified busy (non-shell foreground work).
	BusyTTYs []string
	// BusyLeafByTTY overrides the default busy leaf command (python train.py)
	// for that short tty. Use e.g. "mark still waiting" for mark fixtures.
	BusyLeafByTTY map[string]string
	// CwdByTTY sets the cwd returned for processes on that short tty (default /tmp).
	CwdByTTY map[string]string
	Now      time.Time
	Hostname string
	// AgentResolveByTTY injects procresolve results keyed by short tty (e.g. "ttys002").
	// Applied after process enrich for busy sessions when enrich is on.
	AgentResolveByTTY map[string]AgentResolveFixture
}

// InstallPhasedFixtureCollectorForTest installs an injectable SnapshotCollector
// that implements ListWindows / ListTabsAndSessions from opts.Windows and
// process enrich from IdleTTYs / BusyTTYs. Restored via t.Cleanup.
// Holds an exclusive inject lock so parallel doctest leaves cannot race.
func InstallPhasedFixtureCollectorForTest(t testing.TB, opts PhasedFixtureOpts) {
	t.Helper()
	release := holdTestCollector()
	t.Cleanup(release)

	idleSet := map[string]bool{}
	for _, tty := range opts.IdleTTYs {
		idleSet[strings.TrimPrefix(tty, "/dev/")] = true
	}
	busySet := map[string]bool{}
	for _, tty := range opts.BusyTTYs {
		busySet[strings.TrimPrefix(tty, "/dev/")] = true
	}
	leafByTTY := map[string]string{}
	for k, v := range opts.BusyLeafByTTY {
		leafByTTY[strings.TrimPrefix(k, "/dev/")] = v
	}
	cwdByTTY := map[string]string{}
	for k, v := range opts.CwdByTTY {
		cwdByTTY[strings.TrimPrefix(k, "/dev/")] = v
	}
	// Map pid → tty short for ListCwds (fixture pids are unique per tty family).
	pidTTY := map[int]string{}
	now := opts.Now
	if now.IsZero() {
		now = time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	}
	host := opts.Hostname
	if host == "" {
		host = "testhost"
	}
	// Clone windows so enrich mutations do not alter the caller's slice.
	fxWindows := cloneWindows(opts.Windows)

	agentByTTY := opts.AgentResolveByTTY
	if agentByTTY != nil {
		// shallow copy map so tests cannot mutate after install
		cp := make(map[string]AgentResolveFixture, len(agentByTTY))
		for k, v := range agentByTTY {
			cp[k] = v
		}
		agentByTTY = cp
	}

	c := &SnapshotCollector{
		fixtureEnabled:    true,
		fixtureWindows:    fxWindows,
		agentResolveByTTY: agentByTTY,
		OnListWindows:     opts.OnListWindows,
		OnListTabs:        opts.OnListTabs,
		ITermRunning:      func() bool { return opts.ITermRunning },
		Now:               func() time.Time { return now },
		Hostname:          func() (string, error) { return host, nil },
		ListProcs: func(ttyShort string) ([]rawProc, error) {
			short := strings.TrimPrefix(ttyShort, "/dev/")
			if idleSet[short] {
				pidTTY[1] = short
				pidTTY[2] = short
				return []rawProc{
					{PID: 1, PPID: 0, Stat: "Ss", Etime: "1:00", RSSKB: 1000, Command: "login -fp u"},
					{PID: 2, PPID: 1, Stat: "S+", Etime: "0:59", RSSKB: 2000, Command: "-zsh"},
				}, nil
			}
			if busySet[short] {
				leaf := leafByTTY[short]
				if leaf == "" {
					leaf = "python train.py"
				}
				// Distinct pids per tty so cwd map can key by pid.
				base := 100 + len(pidTTY)*10
				pidTTY[base] = short
				pidTTY[base+1] = short
				pidTTY[base+2] = short
				return []rawProc{
					{PID: base, PPID: 0, Stat: "Ss", Etime: "1:00", RSSKB: 1000, Command: "login -fp u"},
					{PID: base + 1, PPID: base, Stat: "S", Etime: "0:59", RSSKB: 2000, Command: "-zsh"},
					{PID: base + 2, PPID: base + 1, Stat: "R+", Etime: "0:30", RSSKB: 8000, Command: leaf},
				}, nil
			}
			return nil, nil
		},
		ListCwds: func(pids []int) (map[int]string, error) {
			m := map[int]string{}
			for _, p := range pids {
				cwd := "/tmp"
				if short, ok := pidTTY[p]; ok {
					if c, ok := cwdByTTY[short]; ok && c != "" {
						cwd = c
					}
				}
				m[p] = cwd
			}
			return m, nil
		},
		RunAppleScript: func(string) (string, error) {
			return "", fmt.Errorf("fixture collector: AppleScript not used")
		},
	}
	SetSnapshotCollectorForTest(c)
}

// CaptureSnapshot builds a full live snapshot of iTerm2 sessions.
func CaptureSnapshot() (*Snapshot, []string, error) {
	return activeCollector().CaptureWith(CaptureOpts{})
}

// CaptureSnapshotWith builds a snapshot with capture options (e.g. NoEnrich).
func CaptureSnapshotWith(opts CaptureOpts) (*Snapshot, []string, error) {
	return activeCollector().CaptureWith(opts)
}

// Capture runs phased hierarchy collection + process enrichment.
func (c *SnapshotCollector) Capture() (*Snapshot, []string, error) {
	return c.capture(nil, CaptureOpts{})
}

// CaptureWith runs Capture with options (e.g. skip agent enrich).
func (c *SnapshotCollector) CaptureWith(opts CaptureOpts) (*Snapshot, []string, error) {
	return c.capture(nil, opts)
}

// capture is Capture with an optional per-window callback after enrich
// (used by streaming CLI to emit each window block as soon as it is ready).
func (c *SnapshotCollector) capture(onWindowReady func(win SnapshotWindow) error, opts CaptureOpts) (*Snapshot, []string, error) {
	if c.ITermRunning != nil && !c.ITermRunning() {
		return nil, nil, fmt.Errorf("Error: iTerm2 is not running")
	}
	listProcs := c.ListProcs
	if listProcs == nil {
		listProcs = defaultListProcs
	}
	listCwds := c.ListCwds
	if listCwds == nil {
		listCwds = defaultListCwds
	}
	nowFn := c.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	hostFn := c.Hostname
	if hostFn == nil {
		hostFn = os.Hostname
	}

	now := nowFn()
	host, _ := hostFn()

	headers, warnings, err := c.ListWindows()
	if err != nil {
		return nil, warnings, err
	}

	// Cache process data by tty short name.
	type ttyCache struct {
		procs []rawProc
		cwds  map[int]string
	}
	cache := map[string]*ttyCache{}
	getTTY := func(tty string) *ttyCache {
		short := strings.TrimPrefix(tty, "/dev/")
		if short == "" {
			return &ttyCache{}
		}
		if cached, ok := cache[short]; ok {
			return cached
		}
		procs, err := listProcs(short)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("warning: ps failed for %s: %v", tty, err))
			procs = nil
		}
		var pids []int
		for _, p := range procs {
			pids = append(pids, p.PID)
		}
		cwds := map[int]string{}
		if len(pids) > 0 {
			if m, err := listCwds(pids); err != nil {
				warnings = append(warnings, fmt.Sprintf("warning: cwd probe failed for %s: %v", tty, err))
			} else {
				cwds = m
			}
		}
		if len(procs) == 0 {
			warnings = append(warnings, fmt.Sprintf("warning: no processes on %s", tty))
		}
		tc := &ttyCache{procs: procs, cwds: cwds}
		cache[short] = tc
		return tc
	}

	var nTabs, nSess, nIdle, nBusy, nUnknown int
	windows := make([]SnapshotWindow, 0, len(headers))
	for _, hdr := range headers {
		tabs, w2, err := c.ListTabsAndSessions(hdr.Index)
		if err != nil {
			return nil, append(warnings, w2...), err
		}
		warnings = append(warnings, w2...)
		win := SnapshotWindow{Index: hdr.Index, Name: hdr.Name, WindowID: hdr.WindowID, Tabs: tabs}
		for ti := range win.Tabs {
			t := &win.Tabs[ti]
			nTabs++
			for si := range t.Sessions {
				s := &t.Sessions[si]
				nSess++
				s.WindowIndex = win.Index
				s.TabIndex = t.Index
				tc := getTTY(s.TTY)
				idle, shellPID, chosen, cwd, snapProcs := enrichFromProcs(tc.procs, tc.cwds, now)
				applyChosenToSession(s, idle, shellPID, chosen, cwd, snapProcs, now)
				c.attachAgent(s, opts.NoEnrich)
				if s.Idle == nil {
					nUnknown++
				} else if *s.Idle {
					nIdle++
				} else {
					nBusy++
				}
			}
		}
		if onWindowReady != nil {
			if err := onWindowReady(win); err != nil {
				return nil, warnings, err
			}
		}
		windows = append(windows, win)
	}

	snap := &Snapshot{
		CapturedAt: now.Format("2006-01-02T15:04:05") + zoneOffset(now),
		Host:       host,
		Source:     "iterm2",
		Summary: SnapshotSummary{
			Windows:  len(windows),
			Tabs:     nTabs,
			Sessions: nSess,
			Idle:     nIdle,
			Busy:     nBusy,
			Unknown:  nUnknown,
		},
		Windows: windows,
	}
	return snap, warnings, nil
}

// ListWindows returns window index + name headers (tabs may be empty).
func (c *SnapshotCollector) ListWindows() (windows []SnapshotWindow, warnings []string, err error) {
	if c.OnListWindows != nil {
		c.OnListWindows()
	}
	if c.fixtureEnabled {
		out := make([]SnapshotWindow, len(c.fixtureWindows))
		for i, w := range c.fixtureWindows {
			out[i] = SnapshotWindow{Index: w.Index, Name: w.Name, WindowID: w.WindowID}
		}
		return out, nil, nil
	}
	runAS := c.RunAppleScript
	if runAS == nil {
		runAS = defaultRunAppleScript
	}
	raw, err := runAS(listWindowsAppleScript)
	if err != nil {
		return nil, nil, fmt.Errorf("Error: failed to query iTerm2: %w", err)
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		// No windows is valid.
		return []SnapshotWindow{}, nil, nil
	}
	wins, warnings := parseHierarchy(raw)
	// Ensure headers-only (strip any accidental tabs).
	for i := range wins {
		wins[i].Tabs = nil
	}
	return wins, warnings, nil
}

// ListTabsAndSessions returns tabs and sessions for one window (by 1-based index).
func (c *SnapshotCollector) ListTabsAndSessions(windowIndex int) (tabs []SnapshotTab, warnings []string, err error) {
	if c.OnListTabs != nil {
		c.OnListTabs(windowIndex)
	}
	if c.fixtureEnabled {
		for _, w := range c.fixtureWindows {
			if w.Index == windowIndex {
				return cloneTabs(w.Tabs), nil, nil
			}
		}
		return nil, nil, fmt.Errorf("Error: window %d not found", windowIndex)
	}
	runAS := c.RunAppleScript
	if runAS == nil {
		runAS = defaultRunAppleScript
	}
	script := listTabsAndSessionsAppleScript(windowIndex)
	raw, err := runAS(script)
	if err != nil {
		return nil, nil, fmt.Errorf("Error: failed to query iTerm2 window %d: %w", windowIndex, err)
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []SnapshotTab{}, nil, nil
	}
	// parseHierarchy expects optional ###W###; tabs may appear alone.
	// Prefix a synthetic window so session/tab parsing works.
	wrapped := fmt.Sprintf("###W###%d###\n%s", windowIndex, raw)
	wins, warnings := parseHierarchy(wrapped)
	if len(wins) == 0 {
		return []SnapshotTab{}, warnings, nil
	}
	return wins[0].Tabs, warnings, nil
}

func cloneWindows(in []SnapshotWindow) []SnapshotWindow {
	out := make([]SnapshotWindow, len(in))
	for i, w := range in {
		out[i] = SnapshotWindow{Index: w.Index, Name: w.Name, WindowID: w.WindowID, Tabs: cloneTabs(w.Tabs)}
	}
	return out
}

func cloneTabs(tabs []SnapshotTab) []SnapshotTab {
	if tabs == nil {
		return nil
	}
	out := make([]SnapshotTab, len(tabs))
	for i, t := range tabs {
		out[i] = SnapshotTab{Index: t.Index, Name: t.Name}
		if len(t.Sessions) > 0 {
			out[i].Sessions = make([]SnapshotSession, len(t.Sessions))
			copy(out[i].Sessions, t.Sessions)
		}
	}
	return out
}

// FindSessionInSnapshot returns sessions matching a user-supplied id token.
func FindSessionsByRef(snap *Snapshot, ref string) []*SnapshotSession {
	ref = strings.TrimSpace(ref)
	if ref == "" || snap == nil {
		return nil
	}
	var out []*SnapshotSession
	refLower := strings.ToLower(ref)
	refTTY := ref
	if !strings.HasPrefix(refTTY, "/dev/") && strings.HasPrefix(refLower, "ttys") {
		refTTY = "/dev/" + ref
	}

	// PID exact?
	pidVal, pidErr := strconv.Atoi(ref)
	pidOK := pidErr == nil

	for wi := range snap.Windows {
		for ti := range snap.Windows[wi].Tabs {
			for si := range snap.Windows[wi].Tabs[ti].Sessions {
				s := &snap.Windows[wi].Tabs[ti].Sessions[si]
				// Full UUID or prefix (case-insensitive)
				idLower := strings.ToLower(s.ID)
				if idLower == refLower || (len(refLower) >= 8 && strings.HasPrefix(idLower, refLower)) {
					out = append(out, s)
					continue
				}
				// tty
				if s.TTY == ref || s.TTY == refTTY || strings.TrimPrefix(s.TTY, "/dev/") == strings.TrimPrefix(ref, "/dev/") {
					// avoid double-add if also matched id
					if !containsSession(out, s) {
						out = append(out, s)
					}
					continue
				}
				// pid
				if pidOK {
					if s.PID != nil && *s.PID == pidVal {
						if !containsSession(out, s) {
							out = append(out, s)
						}
					} else if s.ShellPID != nil && *s.ShellPID == pidVal {
						if !containsSession(out, s) {
							out = append(out, s)
						}
					}
				}
			}
		}
	}
	return out
}

func containsSession(list []*SnapshotSession, s *SnapshotSession) bool {
	for _, x := range list {
		if x == s || (x.ID != "" && x.ID == s.ID) {
			return true
		}
	}
	return false
}

const listWindowsAppleScript = `
tell application "iTerm2"
  set out to ""
  set wi to 0
  repeat with w in windows
    set wi to wi + 1
    try
      set wname to name of w
    on error
      set wname to ""
    end try
    try
      set wid to id of w
    on error
      set wid to 0
    end try
    set out to out & "###W###" & wi & "###" & wname & "###" & wid & linefeed
  end repeat
  return out
end tell
`

func listTabsAndSessionsAppleScript(windowIndex int) string {
	// Iterate windows by 1-based ordinal matching ListWindows numbering.
	return fmt.Sprintf(`
tell application "iTerm2"
  set out to ""
  set wi to 0
  set target to %d
  repeat with w in windows
    set wi to wi + 1
    if wi is target then
      set ti to 0
      repeat with t in tabs of w
        set ti to ti + 1
        try
          set tname to name of current session of t
        on error
          set tname to ""
        end try
        set out to out & "###T###" & ti & "###" & tname & linefeed
        set si to 0
        repeat with s in sessions of t
          set si to si + 1
          try
            set nm to name of s
          on error
            set nm to "?"
          end try
          try
            set ttyn to tty of s
          on error
            set ttyn to ""
          end try
          try
            set prof to profile name of s
          on error
            set prof to ""
          end try
          try
            set proc to is processing of s
          on error
            set proc to false
          end try
          try
            set uid to unique ID of s
          on error
            set uid to ""
          end try
          set out to out & "###S###" & si & "###" & ttyn & "###" & proc & "###" & prof & "###" & uid & "###" & nm & linefeed
        end repeat
      end repeat
      exit repeat
    end if
  end repeat
  return out
end tell
`, windowIndex)
}

func parseHierarchy(raw string) ([]SnapshotWindow, []string) {
	var warnings []string
	var windows []SnapshotWindow
	var curW *SnapshotWindow
	var curT *SnapshotTab

	for _, row := range strings.Split(raw, "\n") {
		row = strings.TrimRight(row, "\r")
		if row == "" {
			continue
		}
		switch {
		case strings.HasPrefix(row, "###W###"):
			rest := strings.TrimPrefix(row, "###W###")
			// Formats: "idx###name" or "idx###name###windowID"
			parts := strings.SplitN(rest, "###", 3)
			idx, _ := strconv.Atoi(parts[0])
			name := ""
			if len(parts) > 1 {
				name = parts[1]
			}
			var wid uint64
			if len(parts) > 2 {
				if v, err := strconv.ParseUint(strings.TrimSpace(parts[2]), 10, 64); err == nil {
					wid = v
				}
			}
			windows = append(windows, SnapshotWindow{Index: idx, Name: name, WindowID: wid})
			curW = &windows[len(windows)-1]
			curT = nil
		case strings.HasPrefix(row, "###T###"):
			if curW == nil {
				warnings = append(warnings, "warning: tab before window in hierarchy")
				continue
			}
			rest := strings.TrimPrefix(row, "###T###")
			idxStr, name, _ := strings.Cut(rest, "###")
			idx, _ := strconv.Atoi(idxStr)
			curW.Tabs = append(curW.Tabs, SnapshotTab{Index: idx, Name: name})
			curT = &curW.Tabs[len(curW.Tabs)-1]
		case strings.HasPrefix(row, "###S###"):
			if curT == nil {
				warnings = append(warnings, "warning: session before tab in hierarchy")
				continue
			}
			rest := strings.TrimPrefix(row, "###S###")
			parts := strings.SplitN(rest, "###", 6)
			if len(parts) < 6 {
				warnings = append(warnings, "warning: malformed session line")
				continue
			}
			si, _ := strconv.Atoi(parts[0])
			proc := parts[2] == "true"
			curT.Sessions = append(curT.Sessions, SnapshotSession{
				Index:             si,
				TTY:               parts[1],
				ItermIsProcessing: proc,
				Profile:           parts[3],
				ID:                parts[4],
				Name:              parts[5],
			})
		}
	}
	return windows, warnings
}

func defaultITermRunning() bool {
	// System Events name check
	cmd := exec.Command("osascript", "-e", `tell application "System Events" to (name of processes) contains "iTerm2"`)
	out, err := cmd.Output()
	if err == nil && strings.Contains(strings.ToLower(string(out)), "true") {
		return true
	}
	// process path fallback
	cmd = exec.Command("pgrep", "-f", "/Applications/iTerm.app/Contents/MacOS/iTerm2")
	if err := cmd.Run(); err == nil {
		return true
	}
	return false
}

func defaultRunAppleScript(script string) (string, error) {
	cmd := exec.Command("osascript", "-e", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%s", msg)
	}
	return stdout.String(), nil
}

var psLineRe = regexp.MustCompile(`^\s*(\d+)\s+(\d+)\s+(\S+)\s+(\S+)\s+(\d+)\s+([A-Z][a-z]{2}\s+[A-Z][a-z]{2}\s+\d+\s+\d+:\d+:\d+\s+\d+)\s+(.*)$`)

func defaultListProcs(ttyShort string) ([]rawProc, error) {
	cmd := exec.Command("ps", "-t", ttyShort, "-o", "pid=,ppid=,stat=,etime=,rss=,lstart=,command=")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &bytes.Buffer{}
	if err := cmd.Run(); err != nil {
		// no processes is ok
		if stdout.Len() == 0 {
			return nil, nil
		}
	}
	var out []rawProc
	for _, line := range strings.Split(stdout.String(), "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		m := psLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		pid, _ := strconv.Atoi(m[1])
		ppid, _ := strconv.Atoi(m[2])
		rss, _ := strconv.ParseInt(m[5], 10, 64)
		out = append(out, rawProc{
			PID:     pid,
			PPID:    ppid,
			Stat:    m[3],
			Etime:   m[4],
			RSSKB:   rss,
			Lstart:  m[6],
			Command: m[7],
		})
	}
	return out, nil
}

func defaultListCwds(pids []int) (map[int]string, error) {
	if len(pids) == 0 {
		return map[int]string{}, nil
	}
	args := []string{"-a", "-d", "cwd", "-F", "pn"}
	// lsof -p a,b,c
	pidList := make([]string, len(pids))
	for i, p := range pids {
		pidList[i] = strconv.Itoa(p)
	}
	args = append(args, "-p", strings.Join(pidList, ","))
	cmd := exec.Command("lsof", args...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &bytes.Buffer{}
	_ = cmd.Run() // partial results ok
	result := map[int]string{}
	var cur int
	for _, line := range strings.Split(stdout.String(), "\n") {
		if strings.HasPrefix(line, "p") {
			cur, _ = strconv.Atoi(strings.TrimPrefix(line, "p"))
		} else if strings.HasPrefix(line, "n") && cur != 0 {
			result[cur] = strings.TrimPrefix(line, "n")
		}
	}
	return result, nil
}
