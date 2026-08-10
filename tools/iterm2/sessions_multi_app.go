package iterm2

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// Canonical iTerm2 install paths recorded in SaveWindow.app (D9).
// Home form is always ~/… never expanded /Users/….
const (
	CanonicalITermAppSystem = "/Applications/iTerm.app"
	CanonicalITermAppHome   = "~/Applications/iTerm.app"
)

// MultiAppPreflight is the resolved bare-AS app and running install list (save-only).
type MultiAppPreflight struct {
	// AsApp is the canonical app that bare `tell application "iTerm2"` targets.
	AsApp string
	// RunningApps is the set of canonical installs with a live process
	// (home and/or system). Order is not significant.
	RunningApps []string
}

// MultiAppPreflightFn resolves preflight for multi-app save.
type MultiAppPreflightFn func() (MultiAppPreflight, error)

var (
	multiAppMu          sync.Mutex
	multiAppPreflightFn MultiAppPreflightFn // nil → live discovery
)

// SetMultiAppPreflightForTest installs a preflight resolver for tests.
// Pass nil to restore live discovery. Prefer t.Cleanup.
func SetMultiAppPreflightForTest(fn MultiAppPreflightFn) {
	multiAppMu.Lock()
	defer multiAppMu.Unlock()
	multiAppPreflightFn = fn
}

func currentMultiAppPreflightFn() MultiAppPreflightFn {
	multiAppMu.Lock()
	defer multiAppMu.Unlock()
	return multiAppPreflightFn
}

// canonicalITermAppPath maps a filesystem path (or already-canonical form) to
// one of the two canonical app strings. Non-standard home installs (e.g. .bak)
// still map to ~/Applications/iTerm.app (D10).
func canonicalITermAppPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	// Already canonical.
	if path == CanonicalITermAppSystem || path == CanonicalITermAppHome {
		return path
	}
	// Expand ~ for comparison only.
	expanded := path
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			expanded = filepath.Join(home, path[2:])
		}
	}
	// Normalize trailing slash / .app casing via Clean.
	expanded = filepath.Clean(expanded)
	base := filepath.Base(expanded)
	// Treat any …/iTerm.app (or iTerm2.app variants under Applications) as iTerm.
	if !strings.EqualFold(base, "iTerm.app") && !strings.EqualFold(base, "iTerm2.app") {
		// Still accept paths that end with the MacOS binary parent bundle.
		if !strings.Contains(strings.ToLower(expanded), "iterm.app") {
			return ""
		}
	}
	// System Applications.
	if strings.HasPrefix(expanded, "/Applications/") {
		return CanonicalITermAppSystem
	}
	// Home Applications (any user home prefix, or already ~/…).
	if strings.HasPrefix(path, "~/Applications/") || strings.Contains(expanded, "/Applications/iTerm") {
		// Prefer home form when under a user home or explicit ~/Applications.
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			if strings.HasPrefix(expanded, home+string(os.PathSeparator)+"Applications"+string(os.PathSeparator)) ||
				strings.HasPrefix(expanded, home+"/Applications/") {
				return CanonicalITermAppHome
			}
		}
		if strings.HasPrefix(path, "~/Applications/") {
			return CanonicalITermAppHome
		}
	}
	// Fallback: system if absolute under /Applications, else home canonical for other installs.
	if strings.HasPrefix(expanded, "/Applications/") {
		return CanonicalITermAppSystem
	}
	// Non-standard (e.g. ~/Applications/iTerm.app.bak → still home form per D10).
	if strings.Contains(strings.ToLower(expanded), "/applications/") {
		return CanonicalITermAppHome
	}
	return ""
}

// preferredITermApp returns the preferred restore/save default (system when present).
func preferredITermApp() string {
	return CanonicalITermAppSystem
}

// attachAppField sets SaveWindow.App from SnapshotWindow.App or fallback asApp/preferred.
// empty asApp → use preferred when known (D1: always set when known).
func attachAppField(sw *SaveWindow, win SnapshotWindow, asApp string) {
	if sw == nil {
		return
	}
	app := strings.TrimSpace(win.App)
	if app == "" {
		app = strings.TrimSpace(asApp)
	}
	if app == "" {
		app = preferredITermApp()
	}
	if c := canonicalITermAppPath(app); c != "" {
		sw.App = c
		return
	}
	// Prefer raw if already one of the two canonicals.
	if app == CanonicalITermAppSystem || app == CanonicalITermAppHome {
		sw.App = app
		return
	}
	// Last resort: preferred.
	sw.App = preferredITermApp()
}

// resolveMultiAppPreflight returns asApp + running apps (inject or live).
func resolveMultiAppPreflight() MultiAppPreflight {
	if fn := currentMultiAppPreflightFn(); fn != nil {
		if pf, err := fn(); err == nil {
			return normalizePreflight(pf)
		}
	}
	return normalizePreflight(liveMultiAppPreflight())
}

func normalizePreflight(pf MultiAppPreflight) MultiAppPreflight {
	if c := canonicalITermAppPath(pf.AsApp); c != "" {
		pf.AsApp = c
	} else if pf.AsApp == "" {
		pf.AsApp = preferredITermApp()
	}
	seen := map[string]struct{}{}
	var apps []string
	for _, a := range pf.RunningApps {
		c := canonicalITermAppPath(a)
		if c == "" {
			continue
		}
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		apps = append(apps, c)
	}
	// Ensure asApp is listed when known.
	if pf.AsApp != "" {
		if _, ok := seen[pf.AsApp]; !ok {
			apps = append([]string{pf.AsApp}, apps...)
			seen[pf.AsApp] = struct{}{}
		}
	}
	// If discovery found nothing, still record preferred as running (single-app).
	if len(apps) == 0 {
		apps = []string{preferredITermApp()}
		pf.AsApp = preferredITermApp()
	}
	pf.RunningApps = apps
	return pf
}

// liveMultiAppPreflight discovers asApp + running home/system installs.
// asApp: prefer system when both, else whichever is running; falls back to preferred.
func liveMultiAppPreflight() MultiAppPreflight {
	running := discoverRunningITermApps()
	asApp := preferredITermApp()
	// Prefer system as bare AS target when present; else home if only home running.
	hasSystem, hasHome := false, false
	for _, a := range running {
		if a == CanonicalITermAppSystem {
			hasSystem = true
		}
		if a == CanonicalITermAppHome {
			hasHome = true
		}
	}
	if hasSystem {
		asApp = CanonicalITermAppSystem
	} else if hasHome {
		asApp = CanonicalITermAppHome
	}
	return MultiAppPreflight{AsApp: asApp, RunningApps: running}
}

// discoverRunningITermApps returns canonical apps with a live iTerm2 process.
// Uses `ps` so non-standard home paths (e.g. iTerm.app.bak-*) still map to home.
func discoverRunningITermApps() []string {
	out, err := exec.Command("ps", "-axo", "args=").Output()
	if err != nil {
		// Fallback: pgrep exact candidates.
		return discoverRunningITermAppsPgrep()
	}
	seen := map[string]struct{}{}
	var apps []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, "iTerm") {
			continue
		}
		// Executable path is first token of args.
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		exe := fields[0]
		if !strings.HasSuffix(exe, "/Contents/MacOS/iTerm2") && !strings.HasSuffix(exe, "/MacOS/iTerm2") {
			continue
		}
		// Map …/Something.iTerm.app… or …/iTerm.app… → canonical.
		c := canonicalITermAppPath(filepath.Dir(filepath.Dir(filepath.Dir(exe)))) // up to .app
		// Dir thrice: MacOS → Contents → .app
		if c == "" {
			// Try walking up to find *iTerm*.app
			p := exe
			for i := 0; i < 6 && p != "/" && p != "."; i++ {
				if strings.Contains(strings.ToLower(filepath.Base(p)), "iterm") && strings.HasSuffix(strings.ToLower(p), ".app") {
					c = canonicalITermAppPath(p)
					break
				}
				p = filepath.Dir(p)
			}
		}
		if c == "" {
			continue
		}
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		apps = append(apps, c)
	}
	if len(apps) == 0 {
		return discoverRunningITermAppsPgrep()
	}
	return apps
}

func discoverRunningITermAppsPgrep() []string {
	var out []string
	if processMatchesITerm(CanonicalITermAppSystem) {
		out = append(out, CanonicalITermAppSystem)
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		if processMatchesITerm(filepath.Join(home, "Applications", "iTerm.app")) {
			out = append(out, CanonicalITermAppHome)
		}
	}
	return out
}

func processMatchesITerm(appBundle string) bool {
	// pgrep -f on MacOS binary path inside the bundle.
	bin := filepath.Join(appBundle, "Contents", "MacOS", "iTerm2")
	// Also try expanded home.
	if strings.HasPrefix(appBundle, "~/") {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			bin = filepath.Join(home, appBundle[2:], "Contents", "MacOS", "iTerm2")
		}
	}
	cmd := exec.Command("pgrep", "-f", bin)
	if err := cmd.Run(); err == nil {
		return true
	}
	// Broader match on app path fragment.
	cmd = exec.Command("pgrep", "-f", appBundle)
	return cmd.Run() == nil
}

// multiAppCaptureSource is one AppleScript surface for save multi-app capture.
type multiAppCaptureSource struct {
	App  string // canonical app tag
	Tell string // empty = bare "iTerm2"; else absolute .app path
}

// multiAppCaptureSources builds capture order: bare AS (asApp) first, then
// path-tell for each other running install (D6/D7 — no path re-capture of asApp).
func multiAppCaptureSources(pf MultiAppPreflight) []multiAppCaptureSource {
	asApp := pf.AsApp
	if asApp == "" {
		asApp = preferredITermApp()
	}
	sources := []multiAppCaptureSource{{App: asApp, Tell: ""}}
	for _, a := range pf.RunningApps {
		if a == asApp {
			continue
		}
		sources = append(sources, multiAppCaptureSource{
			App:  a,
			Tell: absoluteITermAppPath(a),
		})
	}
	return sources
}

// absoluteITermAppPath expands a canonical app to a filesystem path for AS tell.
func absoluteITermAppPath(canonical string) string {
	c := canonicalITermAppPath(canonical)
	if c == CanonicalITermAppHome {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			return filepath.Join(home, "Applications", "iTerm.app")
		}
	}
	return CanonicalITermAppSystem
}

// --- Restore app disk presence + prefer-home / --same-app targeting ---

// restoreAppDisk describes which canonical iTerm installs exist on disk.
type restoreAppDisk struct {
	HomeExists   bool
	SystemExists bool
}

var (
	restoreAppDiskMu   sync.Mutex
	restoreAppDiskInj  *restoreAppDisk // nil → live os.Stat
	restoreAppDiskHold sync.Mutex      // exclusive inject ownership (parallel-safe)
)

// SetRestoreAppDiskForTest injects which canonical installs exist for restore
// target resolution (prefer-home / --same-app). Takes an exclusive hold until
// ClearRestoreAppDiskForTest so parallel leaves cannot clobber each other.
// Prefer t.Cleanup(ClearRestoreAppDiskForTest).
func SetRestoreAppDiskForTest(homeExists, systemExists bool) {
	restoreAppDiskHold.Lock()
	restoreAppDiskMu.Lock()
	restoreAppDiskInj = &restoreAppDisk{HomeExists: homeExists, SystemExists: systemExists}
	restoreAppDiskMu.Unlock()
}

// ClearRestoreAppDiskForTest restores live disk checks and releases the
// exclusive inject hold from SetRestoreAppDiskForTest.
func ClearRestoreAppDiskForTest() {
	restoreAppDiskMu.Lock()
	restoreAppDiskInj = nil
	restoreAppDiskMu.Unlock()
	restoreAppDiskHold.Unlock()
}

// resolveRestoreAppDisk returns home/system install existence (inject or live).
func resolveRestoreAppDisk() restoreAppDisk {
	restoreAppDiskMu.Lock()
	inj := restoreAppDiskInj
	restoreAppDiskMu.Unlock()
	if inj != nil {
		return *inj
	}
	return liveRestoreAppDisk()
}

func liveRestoreAppDisk() restoreAppDisk {
	var d restoreAppDisk
	if _, err := os.Stat(CanonicalITermAppSystem); err == nil {
		d.SystemExists = true
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		p := filepath.Join(home, "Applications", "iTerm.app")
		if _, err := os.Stat(p); err == nil {
			d.HomeExists = true
		}
	}
	return d
}

// RestoreAppTarget is a resolved create/tell target for restore.
type RestoreAppTarget struct {
	// Canonical is ~/Applications/iTerm.app or /Applications/iTerm.app, or empty when Bare.
	Canonical string
	// Bare means fall back to tell application "iTerm2" (no path).
	Bare bool
	// Warning is optional user message without the "warning:" prefix.
	Warning string
}

// Display returns the plan-text form: canonical path or bare "iTerm2".
func (t RestoreAppTarget) Display() string {
	if t.Bare || t.Canonical == "" {
		return "iTerm2"
	}
	return t.Canonical
}

// TellPath returns the absolute .app path for AppleScript path-tell, or "" for bare.
func (t RestoreAppTarget) TellPath() string {
	if t.Bare || t.Canonical == "" {
		return ""
	}
	return absoluteITermAppPath(t.Canonical)
}

// resolvePreferHomeTarget picks one global create target from disk presence:
// both → home; only home → home; only system → system; neither → bare + warn.
func resolvePreferHomeTarget(disk restoreAppDisk) RestoreAppTarget {
	if disk.HomeExists {
		// Prefer home when both exist, or only home.
		return RestoreAppTarget{Canonical: CanonicalITermAppHome}
	}
	if disk.SystemExists {
		return RestoreAppTarget{Canonical: CanonicalITermAppSystem}
	}
	return RestoreAppTarget{
		Bare: true,
		Warning: "neither iTerm install found on disk (missing ~/Applications/iTerm.app " +
			"and /Applications/iTerm.app); falling back to bare application \"iTerm2\"",
	}
}

// resolveSameAppWindowTarget picks the per-window create target under --same-app.
// Recorded app on disk → that path; empty/missing app or path not on disk →
// prefer-home fallback + warning.
func resolveSameAppWindowTarget(recordedApp string, disk restoreAppDisk) RestoreAppTarget {
	recorded := strings.TrimSpace(recordedApp)
	if recorded == "" {
		fb := resolvePreferHomeTarget(disk)
		fb.Warning = "window has empty/missing recorded app; falling back to prefer-home target"
		if fb.Bare {
			fb.Warning = "window has empty/missing recorded app and neither iTerm install " +
				"found on disk; falling back to bare application \"iTerm2\""
		}
		return fb
	}
	c := canonicalITermAppPath(recorded)
	if c == "" {
		// Non-canonical recorded path — still try exact home/system match strings.
		if recorded == CanonicalITermAppHome || recorded == CanonicalITermAppSystem {
			c = recorded
		} else {
			fb := resolvePreferHomeTarget(disk)
			fb.Warning = fmt.Sprintf("recorded app %q is not a known iTerm install; falling back to prefer-home target", recorded)
			return fb
		}
	}
	onDisk := (c == CanonicalITermAppHome && disk.HomeExists) ||
		(c == CanonicalITermAppSystem && disk.SystemExists)
	if onDisk {
		return RestoreAppTarget{Canonical: c}
	}
	fb := resolvePreferHomeTarget(disk)
	fb.Warning = fmt.Sprintf("recorded app %s not found on disk; falling back to prefer-home target", c)
	if fb.Bare {
		fb.Warning = fmt.Sprintf("recorded app %s not found on disk and neither install available; falling back to bare application \"iTerm2\"", c)
	}
	return fb
}

// CaptureSnapshotForSave captures all multi-app sources when live (not fixture).
// Fixture collectors already carry dual App tags in one snapshot — single pass.
// opts.SpaceAllow enables space-first deep-capture filter (save --spaces).
func CaptureSnapshotForSave(opts CaptureOpts) (*Snapshot, []string, error) {
	return CaptureSnapshotForSaveStream(opts, nil)
}

// CaptureSnapshotForSaveStream is CaptureSnapshotForSave with an optional
// per-window callback after each deep-captured window is ready (streaming dry-run).
func CaptureSnapshotForSaveStream(opts CaptureOpts, onWindowReady func(win SnapshotWindow) error) (*Snapshot, []string, error) {
	c := activeCollector()
	// Fixture / test inject: single collector already has all windows.
	if c.fixtureEnabled {
		return c.capture(onWindowReady, opts)
	}
	pf := resolveMultiAppPreflight()
	return captureLiveMultiApp(c, opts, pf, onWindowReady)
}

// captureLiveMultiApp runs AppleScript against bare iTerm2 + path-tells for
// other running installs, merges windows (dedupe by WindowID), stamps App.
// Space-first (opts.SpaceAllow) skips deep capture on non-matching Desktops.
// onWindowReady is invoked as each deep-captured window completes (stream).
func captureLiveMultiApp(base *SnapshotCollector, opts CaptureOpts, pf MultiAppPreflight, onWindowReady func(win SnapshotWindow) error) (*Snapshot, []string, error) {
	if base == nil {
		base = defaultCollector()
	}
	if base.ITermRunning != nil && !base.ITermRunning() {
		return nil, nil, fmt.Errorf("Error: iTerm2 is not running")
	}
	sources := multiAppCaptureSources(pf)
	var all []SnapshotWindow
	var allWarn []string
	seenID := map[uint64]struct{}{}
	asWindowIDs := map[uint64]struct{}{}
	totalSpaceSkipped := 0

	for si, src := range sources {
		col := *base
		col.AppTell = src.Tell
		col.AppTag = src.App
		// Fresh fixture flags off for live path (base may be default).
		col.fixtureEnabled = false
		col.fixtureWindows = nil

		srcSkipped := 0
		srcOpts := opts
		srcOpts.SpaceSkipped = &srcSkipped

		// Always merge in the per-window callback so live write (nil onWindowReady)
		// still collects windows; stream when onWindowReady is set.
		snap, warns, err := col.capture(func(win SnapshotWindow) error {
			if win.App == "" {
				win.App = src.App
			}
			// Dedupe across sources before stream emit.
			if win.WindowID != 0 {
				if _, ok := seenID[win.WindowID]; ok {
					return nil
				}
				// Path source identical to bare AS surface: skip duplicate.
				if si > 0 {
					if _, ok := asWindowIDs[win.WindowID]; ok {
						return nil
					}
				}
				seenID[win.WindowID] = struct{}{}
				if si == 0 {
					asWindowIDs[win.WindowID] = struct{}{}
				}
			}
			// Assign provisional global index for streaming W{n}.
			win.Index = len(all) + 1
			all = append(all, win)
			if onWindowReady != nil {
				return onWindowReady(win)
			}
			return nil
		}, srcOpts)
		allWarn = append(allWarn, warns...)
		totalSpaceSkipped += srcSkipped
		if err != nil {
			// First source (bare) is fatal; secondary path failures soft-warn.
			if si == 0 {
				if opts.SpaceSkipped != nil {
					*opts.SpaceSkipped = totalSpaceSkipped
				}
				return nil, allWarn, err
			}
			allWarn = append(allWarn, "warning: failed to capture "+src.App+": "+err.Error())
			continue
		}
		_ = snap
	}

	// Renumber indices globally 1…N (callback may have set provisional indices).
	for i := range all {
		all[i].Index = i + 1
	}
	if opts.SpaceSkipped != nil {
		*opts.SpaceSkipped = totalSpaceSkipped
	}

	nowFn := base.Now
	if nowFn == nil {
		nowFn = defaultCollector().Now
	}
	hostFn := base.Hostname
	if hostFn == nil {
		hostFn = defaultCollector().Hostname
	}
	now := nowFn()
	host, _ := hostFn()

	nTabs, nSess, nIdle, nBusy, nUnknown := 0, 0, 0, 0, 0
	for _, win := range all {
		for _, t := range win.Tabs {
			nTabs++
			for _, s := range t.Sessions {
				nSess++
				if s.Idle == nil {
					nUnknown++
				} else if *s.Idle {
					nIdle++
				} else {
					nBusy++
				}
			}
		}
	}
	snap := &Snapshot{
		CapturedAt: now.Format("2006-01-02T15:04:05") + zoneOffset(now),
		Host:       host,
		Source:     "iterm2",
		Summary: SnapshotSummary{
			Windows:  len(all),
			Tabs:     nTabs,
			Sessions: nSess,
			Idle:     nIdle,
			Busy:     nBusy,
			Unknown:  nUnknown,
		},
		Windows: all,
	}
	return snap, allWarn, nil
}

// multiAppCollapseWarning is the dual-running / no-new-ids stderr message (D2).
func multiAppCollapseWarning(asApp string, other string) string {
	if other == "" {
		other = "other install"
	}
	return fmt.Sprintf(
		"dual iTerm installs running; path capture for %s yielded no new windows (same iterm_window_id as %s); saving partial merge",
		other, asApp,
	)
}

// evaluateMultiAppCollapse reports whether dual-running collapsed (other path
// added no window ids not already present under asApp / primary set).
// appsPresent is the set of App tags on kept critical windows after merge.
func evaluateMultiAppCollapse(pf MultiAppPreflight, appsPresent map[string]bool) (warn string) {
	if len(pf.RunningApps) < 2 {
		return ""
	}
	asApp := pf.AsApp
	if asApp == "" {
		asApp = preferredITermApp()
	}
	for _, a := range pf.RunningApps {
		if a == asApp {
			continue
		}
		// Other running install contributed no windows with its App tag.
		if !appsPresent[a] {
			return multiAppCollapseWarning(asApp, a)
		}
	}
	return ""
}

// renumberSaveSourceIndexes sets source_index globally to 1…N in order (D5).
func renumberSaveSourceIndexes(windows []SaveWindow) {
	for i := range windows {
		windows[i].SourceIndex = i + 1
	}
}

// collectAppsPresent returns the set of non-empty App tags on save windows.
func collectAppsPresent(windows []SaveWindow) map[string]bool {
	m := map[string]bool{}
	for _, w := range windows {
		if w.App != "" {
			m[w.App] = true
		}
	}
	return m
}

// applyMultiAppToSaveWindows attaches App (using asApp fallback), renumbers
// source_index, and returns an optional dual-collapse warning.
func applyMultiAppToSaveWindows(windows []SaveWindow, snap *Snapshot, pf MultiAppPreflight) (out []SaveWindow, collapseWarn string) {
	// Build index → SnapshotWindow for App tags from live snap when available.
	byIndex := map[int]SnapshotWindow{}
	if snap != nil {
		for _, w := range snap.Windows {
			byIndex[w.Index] = w
		}
	}
	out = make([]SaveWindow, len(windows))
	copy(out, windows)
	for i := range out {
		win, ok := byIndex[out[i].SourceIndex]
		if !ok {
			// Fall back to synthetic window with App already on SaveWindow if set.
			win = SnapshotWindow{Index: out[i].SourceIndex, App: out[i].App}
		}
		attachAppField(&out[i], win, pf.AsApp)
	}
	// Hard-dedupe by iterm_window_id (keep first).
	out = dedupeSaveWindowsByITermID(out)
	renumberSaveSourceIndexes(out)
	collapseWarn = evaluateMultiAppCollapse(pf, collectAppsPresent(out))
	return out, collapseWarn
}

// dedupeSaveWindowsByITermID keeps the first window for each non-zero iterm_window_id.
// Windows with id 0 are always kept (cannot dedupe).
func dedupeSaveWindowsByITermID(windows []SaveWindow) []SaveWindow {
	seen := map[int64]struct{}{}
	var out []SaveWindow
	for _, w := range windows {
		if w.ItermWindowID != 0 {
			if _, ok := seen[w.ItermWindowID]; ok {
				continue
			}
			seen[w.ItermWindowID] = struct{}{}
		}
		out = append(out, w)
	}
	return out
}

// tagFixtureAppFromName sets SnapshotWindow.App from multi-app fixture names
// when App is empty (From-System / From-Home / System-Space-0 / Home-Space-2).
// Used only by InstallPhasedFixtureCollectorForTest so sealed harness topology
// fixtures get dual App tags without editing sealed DOCTEST helpers.
func tagFixtureAppFromName(w *SnapshotWindow) {
	if w == nil || w.App != "" {
		return
	}
	n := strings.ToLower(w.Name)
	switch {
	case strings.Contains(n, "home"):
		w.App = CanonicalITermAppHome
	case strings.Contains(n, "system"):
		w.App = CanonicalITermAppSystem
	}
}
