package iterm2

import (
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/itermsnapshot"
	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot"
)

// SnapshotCollector is a thin adapter over snapshot.Collector plus agent-pro
// itermsnapshot attach. Hierarchy + process enrich + SpaceAllow live in the lib;
// Agent attach and kool Snapshot (with Agent) are composed here.
type SnapshotCollector struct {
	lib *snapshot.Collector

	// OnListWindows is an optional test hook invoked at the start of ListWindows.
	OnListWindows func()
	// OnListTabs is an optional test hook invoked at the start of ListTabsAndSessions.
	OnListTabs func(windowIndex int)

	// ResolveFromPID optionally overrides live procresolve (production default).
	ResolveFromPID func(pid int) (*procresolve.Result, error)

	// AppTell is the AppleScript application target (forwarded to lib).
	AppTell string
	// AppTag is stamped on windows when App is empty (forwarded to lib).
	AppTag string

	// RunAppleScript / ListProcs / ListTTYProcs / ListCwds / ITermRunning are
	// injectable for live_scan and tests; synced onto lib before capture.
	RunAppleScript func(script string) (string, error)
	ListProcs      func(ttyShort string) ([]snapshot.ProcRow, error)
	ListTTYProcs   func(ttyShorts []string) ([]liveTTYProc, error)
	ListCwds       func(pids []int) (map[int]string, error)
	ITermRunning   func() bool
	Now            func() time.Time
	Hostname       func() (string, error)

	// fixtureEnabled is set by InstallPhasedFixtureCollectorForTest (multi-app
	// short-circuits to single-pass capture when true).
	fixtureEnabled bool
	// fixtureWindows retained for multi-app live reset (lib owns fixture data).
	fixtureWindows []SnapshotWindow

	// agentResolveByTTY injects procresolve results by short tty (tests).
	agentResolveByTTY map[string]AgentResolveFixture
}

func defaultCollector() *SnapshotCollector {
	lib := snapshot.NewCollector()
	return &SnapshotCollector{
		lib:            lib,
		RunAppleScript: lib.RunAppleScript,
		ListProcs:      lib.ListProcs,
		ListCwds:       lib.ListCwds,
		ITermRunning:   lib.ITermRunning,
		Now:            lib.Now,
		Hostname:       lib.Hostname,
		ListTTYProcs: func(ttys []string) ([]liveTTYProc, error) {
			rows, err := lib.ListTTYProcs(ttys)
			if err != nil {
				return nil, err
			}
			out := make([]liveTTYProc, len(rows))
			for i, r := range rows {
				out[i] = liveTTYProc{rawProc: r.ProcRow, TTY: r.TTY}
			}
			return out, nil
		},
	}
}

var (
	testCollector     *SnapshotCollector
	testCollectorMu   sync.Mutex
	testCollectorHold sync.Mutex
)

// SetSnapshotCollectorForTest overrides the collector used by CaptureSnapshot.
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
	Windows           []SnapshotWindow
	ITermRunning      bool
	OnListWindows     func()
	OnListTabs        func(windowIndex int)
	IdleTTYs          []string
	BusyTTYs          []string
	BusyLeafByTTY     map[string]string
	CwdByTTY          map[string]string
	Now               time.Time
	Hostname          string
	AgentResolveByTTY map[string]AgentResolveFixture
}

// InstallPhasedFixtureCollectorForTest installs an injectable SnapshotCollector.
func InstallPhasedFixtureCollectorForTest(t testing.TB, opts PhasedFixtureOpts) {
	t.Helper()
	release := holdTestCollector()
	t.Cleanup(release)

	agentByTTY := opts.AgentResolveByTTY
	if agentByTTY != nil {
		cp := make(map[string]AgentResolveFixture, len(agentByTTY))
		for k, v := range agentByTTY {
			cp[k] = v
		}
		agentByTTY = cp
	}

	fxWindows := make([]SnapshotWindow, len(opts.Windows))
	copy(fxWindows, opts.Windows)
	for i := range fxWindows {
		tagFixtureAppFromName(&fxWindows[i])
	}

	lib := snapshot.NewCollector()
	lib.ApplyPhasedFixture(snapshot.PhasedFixtureOpts{
		Windows:       toLibWindows(fxWindows),
		ITermRunning:  opts.ITermRunning,
		IdleTTYs:      opts.IdleTTYs,
		BusyTTYs:      opts.BusyTTYs,
		BusyLeafByTTY: opts.BusyLeafByTTY,
		CwdByTTY:      opts.CwdByTTY,
		Now:           opts.Now,
		Hostname:      opts.Hostname,
	})
	lib.OnListWindows = opts.OnListWindows
	lib.OnListTabs = opts.OnListTabs

	c := &SnapshotCollector{
		lib:               lib,
		OnListWindows:     opts.OnListWindows,
		OnListTabs:        opts.OnListTabs,
		ITermRunning:      lib.ITermRunning,
		Now:               lib.Now,
		Hostname:          lib.Hostname,
		ListProcs:         lib.ListProcs,
		ListCwds:          lib.ListCwds,
		fixtureEnabled:    true,
		agentResolveByTTY: agentByTTY,
	}
	SetSnapshotCollectorForTest(c)
}

func (c *SnapshotCollector) ensureLib() *snapshot.Collector {
	if c.lib == nil {
		c.lib = snapshot.NewCollector()
	}
	return c.lib
}

func (c *SnapshotCollector) syncLib() *snapshot.Collector {
	lib := c.ensureLib()
	lib.OnListWindows = c.OnListWindows
	lib.OnListTabs = c.OnListTabs
	lib.AppTell = c.AppTell
	lib.AppTag = c.AppTag
	if c.RunAppleScript != nil {
		lib.RunAppleScript = c.RunAppleScript
	}
	if c.ListProcs != nil {
		lib.ListProcs = c.ListProcs
		// Injected per-TTY ListProcs must win over production ListAllProcs.
		lib.ListAllProcs = nil
	}
	if c.ListTTYProcs != nil {
		listTTY := c.ListTTYProcs
		lib.ListTTYProcs = func(ttys []string) ([]snapshot.TTYProc, error) {
			rows, err := listTTY(ttys)
			if err != nil {
				return nil, err
			}
			out := make([]snapshot.TTYProc, len(rows))
			for i, r := range rows {
				out[i] = snapshot.TTYProc{ProcRow: r.rawProc, TTY: r.TTY}
			}
			return out, nil
		}
	}
	if c.ListCwds != nil {
		lib.ListCwds = c.ListCwds
	}
	if c.ITermRunning != nil {
		lib.ITermRunning = c.ITermRunning
	}
	if c.Now != nil {
		lib.Now = c.Now
	}
	if c.Hostname != nil {
		lib.Hostname = c.Hostname
	}
	// Prefer kool injectable Space resolver (SetSpaceIndexForWindowForTest).
	lib.ResolveSpace = func(windowID uint64) (int, error) {
		return currentSpaceIndexResolver()(windowID)
	}
	return lib
}

// ListWindows returns window index + name headers (tabs may be empty).
func (c *SnapshotCollector) ListWindows() (windows []SnapshotWindow, warnings []string, err error) {
	if c == nil {
		c = defaultCollector()
	}
	libWins, warnings, err := c.syncLib().ListWindows()
	if err != nil {
		return nil, warnings, err
	}
	out := make([]SnapshotWindow, len(libWins))
	for i, w := range libWins {
		out[i] = toKoolWindow(w)
	}
	return out, warnings, nil
}

// ListTabsAndSessions returns tabs and sessions for one window (by 1-based index).
func (c *SnapshotCollector) ListTabsAndSessions(windowIndex int) (tabs []SnapshotTab, warnings []string, err error) {
	if c == nil {
		c = defaultCollector()
	}
	libTabs, warnings, err := c.syncLib().ListTabsAndSessions(windowIndex)
	if err != nil {
		return nil, warnings, err
	}
	out := make([]SnapshotTab, len(libTabs))
	for i, t := range libTabs {
		out[i] = toKoolTab(t)
	}
	return out, warnings, nil
}

// CaptureSnapshot builds a full live snapshot of iTerm2 sessions (multi-app).
func CaptureSnapshot() (*Snapshot, []string, error) {
	return CaptureSnapshotStream(CaptureOpts{}, nil)
}

// CaptureSnapshotWith builds a snapshot with capture options (multi-app stream).
func CaptureSnapshotWith(opts CaptureOpts) (*Snapshot, []string, error) {
	return CaptureSnapshotStream(opts, nil)
}

// Capture runs phased hierarchy collection + process enrichment + agent attach.
func (c *SnapshotCollector) Capture() (*Snapshot, []string, error) {
	return c.capture(nil, CaptureOpts{})
}

// CaptureWith runs Capture with options (e.g. skip agent enrich, SpaceAllow).
func (c *SnapshotCollector) CaptureWith(opts CaptureOpts) (*Snapshot, []string, error) {
	return c.capture(nil, opts)
}

// capture is Capture with an optional per-window callback after enrich.
func (c *SnapshotCollector) capture(onWindowReady func(win SnapshotWindow) error, opts CaptureOpts) (*Snapshot, []string, error) {
	if c == nil {
		c = defaultCollector()
	}
	if !c.fixtureEnabled {
		// Multi-app live path shallow-copies SnapshotCollector with a shared lib
		// pointer; give each source its own Collector so AppTell/AppTag cannot leak.
		c.lib = snapshot.NewCollector()
	} else if c.lib == nil {
		c.lib = snapshot.NewCollector()
	}
	lib := c.syncLib()

	libOpts := snapshot.CaptureOpts{
		// Kool snapshot historically always attached cwd (save / critical index).
		IncludeCwd:   true,
		SpaceAllow:   opts.SpaceAllow,
		SpaceSkipped: opts.SpaceSkipped,
	}

	var progressiveWindows []SnapshotWindow
	libSnap, warnings, err := lib.CaptureProgressiveWith(libOpts, func(libWin snapshot.SnapshotWindow) error {
		kWin := c.enrichWindowAgents(libWin, opts.NoEnrich)
		progressiveWindows = append(progressiveWindows, kWin)
		if onWindowReady != nil {
			return onWindowReady(kWin)
		}
		return nil
	})
	if err != nil {
		return nil, warnings, err
	}

	var kool *Snapshot
	if len(progressiveWindows) > 0 {
		kool = &Snapshot{
			CapturedAt: libSnap.CapturedAt,
			Host:       libSnap.Host,
			Source:     libSnap.Source,
			Summary: SnapshotSummary{
				Windows:  libSnap.Summary.Windows,
				Tabs:     libSnap.Summary.Tabs,
				Sessions: libSnap.Summary.Sessions,
				Idle:     libSnap.Summary.Idle,
				Busy:     libSnap.Summary.Busy,
				Unknown:  libSnap.Summary.Unknown,
			},
			Windows: progressiveWindows,
		}
	} else {
		kool = toKoolSnapshot(libSnap)
		if !opts.NoEnrich && libSnap != nil {
			res, w2, aerr := itermsnapshot.Capture(itermsnapshot.CaptureOpts{
				Snapshot:       libSnap,
				NoEnrich:       false,
				ResolveFromPID: c.makeResolve(libSnap),
			})
			if aerr == nil && res != nil {
				attachAgents(kool, res.Agents)
			}
			warnings = append(warnings, w2...)
		}
	}
	return kool, warnings, nil
}

// FindSessionsByRef returns sessions matching a user-supplied id token.
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

	pidVal, pidErr := strconv.Atoi(ref)
	pidOK := pidErr == nil

	for wi := range snap.Windows {
		for ti := range snap.Windows[wi].Tabs {
			for si := range snap.Windows[wi].Tabs[ti].Sessions {
				s := &snap.Windows[wi].Tabs[ti].Sessions[si]
				idLower := strings.ToLower(s.ID)
				if idLower == refLower || (len(refLower) >= 8 && strings.HasPrefix(idLower, refLower)) {
					out = append(out, s)
					continue
				}
				if s.TTY == ref || s.TTY == refTTY || strings.TrimPrefix(s.TTY, "/dev/") == strings.TrimPrefix(ref, "/dev/") {
					if !containsSession(out, s) {
						out = append(out, s)
					}
					continue
				}
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

func toLibWindows(in []SnapshotWindow) []snapshot.SnapshotWindow {
	if in == nil {
		return nil
	}
	out := make([]snapshot.SnapshotWindow, len(in))
	for i, w := range in {
		out[i] = toLibWindow(w)
	}
	return out
}

func toLibWindow(w SnapshotWindow) snapshot.SnapshotWindow {
	out := snapshot.SnapshotWindow{
		Index: w.Index, Name: w.Name, WindowID: w.WindowID,
		FixedSpace: cloneIntPtr(w.FixedSpace), App: w.App,
	}
	if len(w.Tabs) == 0 {
		return out
	}
	out.Tabs = make([]snapshot.SnapshotTab, len(w.Tabs))
	for i, t := range w.Tabs {
		out.Tabs[i] = toLibTab(t)
	}
	return out
}

func toLibTab(t SnapshotTab) snapshot.SnapshotTab {
	out := snapshot.SnapshotTab{Index: t.Index, Name: t.Name}
	if len(t.Sessions) == 0 {
		return out
	}
	out.Sessions = make([]snapshot.SnapshotSession, len(t.Sessions))
	for i, s := range t.Sessions {
		out.Sessions[i] = toLibSession(s)
	}
	return out
}

func toLibSession(s SnapshotSession) snapshot.SnapshotSession {
	return snapshot.SnapshotSession{
		Index: s.Index, ID: s.ID, Name: s.Name, TTY: s.TTY, Profile: s.Profile,
		ItermIsProcessing: s.ItermIsProcessing, Idle: s.Idle, Cwd: s.Cwd,
		ShellPID: s.ShellPID, PID: s.PID, PPID: s.PPID, Stat: s.Stat,
		Command: s.Command, CommandLine: s.CommandLine, StartTime: s.StartTime,
		StartTimeUnix: s.StartTimeUnix, DurationSeconds: s.DurationSeconds,
		Duration: s.Duration, Etime: s.Etime, RSSKB: s.RSSKB,
		Processes: toLibProcs(s.Processes), WindowIndex: s.WindowIndex, TabIndex: s.TabIndex,
	}
}

func toLibProcs(in []SnapshotProc) []snapshot.SnapshotProc {
	if in == nil {
		return nil
	}
	out := make([]snapshot.SnapshotProc, len(in))
	for i, p := range in {
		out[i] = snapshot.SnapshotProc{
			PID: p.PID, PPID: p.PPID, Stat: p.Stat, Etime: p.Etime,
			DurationSeconds: p.DurationSeconds, Duration: p.Duration,
			StartTime: p.StartTime, StartTimeUnix: p.StartTimeUnix,
			RSSKB: p.RSSKB, Command: p.Command,
		}
	}
	return out
}

func toKoolSnapshot(s *snapshot.Snapshot) *Snapshot {
	if s == nil {
		return nil
	}
	out := &Snapshot{
		CapturedAt: s.CapturedAt, Host: s.Host, Source: s.Source,
		Summary: SnapshotSummary{
			Windows: s.Summary.Windows, Tabs: s.Summary.Tabs, Sessions: s.Summary.Sessions,
			Idle: s.Summary.Idle, Busy: s.Summary.Busy, Unknown: s.Summary.Unknown,
		},
	}
	if len(s.Windows) > 0 {
		out.Windows = make([]SnapshotWindow, len(s.Windows))
		for i, w := range s.Windows {
			out.Windows[i] = toKoolWindow(w)
		}
	}
	return out
}

func toKoolWindow(w snapshot.SnapshotWindow) SnapshotWindow {
	out := SnapshotWindow{
		Index: w.Index, Name: w.Name, WindowID: w.WindowID,
		FixedSpace: cloneIntPtr(w.FixedSpace), App: w.App,
	}
	if len(w.Tabs) == 0 {
		return out
	}
	out.Tabs = make([]SnapshotTab, len(w.Tabs))
	for i, t := range w.Tabs {
		out.Tabs[i] = toKoolTab(t)
	}
	return out
}

func toKoolTab(t snapshot.SnapshotTab) SnapshotTab {
	out := SnapshotTab{Index: t.Index, Name: t.Name}
	if len(t.Sessions) == 0 {
		return out
	}
	out.Sessions = make([]SnapshotSession, len(t.Sessions))
	for i, s := range t.Sessions {
		out.Sessions[i] = toKoolSession(s)
	}
	return out
}

func toKoolSession(s snapshot.SnapshotSession) SnapshotSession {
	return SnapshotSession{
		Index: s.Index, ID: s.ID, Name: s.Name, TTY: s.TTY, Profile: s.Profile,
		ItermIsProcessing: s.ItermIsProcessing, Idle: s.Idle, Cwd: s.Cwd,
		ShellPID: s.ShellPID, PID: s.PID, PPID: s.PPID, Stat: s.Stat,
		Command: s.Command, CommandLine: s.CommandLine, StartTime: s.StartTime,
		StartTimeUnix: s.StartTimeUnix, DurationSeconds: s.DurationSeconds,
		Duration: s.Duration, Etime: s.Etime, RSSKB: s.RSSKB,
		Processes: toKoolProcs(s.Processes), WindowIndex: s.WindowIndex, TabIndex: s.TabIndex,
	}
}

func toKoolProcs(in []snapshot.SnapshotProc) []SnapshotProc {
	if in == nil {
		return nil
	}
	out := make([]SnapshotProc, len(in))
	for i, p := range in {
		out[i] = SnapshotProc{
			PID: p.PID, PPID: p.PPID, Stat: p.Stat, Etime: p.Etime,
			DurationSeconds: p.DurationSeconds, Duration: p.Duration,
			StartTime: p.StartTime, StartTimeUnix: p.StartTimeUnix,
			RSSKB: p.RSSKB, Command: p.Command,
		}
	}
	return out
}

func cloneIntPtr(p *int) *int {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}
