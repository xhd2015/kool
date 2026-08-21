package iterm2

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/computer-use/macos/space"
)

// maxRecordedSpace is the exclusive upper bound for a 0-based Space index
// (macOS hard-caps at 16 Desktops → indices 0..15).
const maxRecordedSpace = 16

// spaceSwitchRetries is how many Switch attempts before soft-falling back to
// the current Desktop (Mission Control AX is racy; matches space.CreateAndActivate).
const spaceSwitchRetries = 3

// spaceSwitchSettle waits between Switch retries. Tests may set to 0 via
// SetSpaceSwitchSettleForTest.
var (
	spaceSwitchSettleMu sync.Mutex
	spaceSwitchSettle   = 500 * time.Millisecond
)

// SpaceIndexResolver maps an iTerm/CG window id to a 0-based Desktop index.
type SpaceIndexResolver func(windowID uint64) (int, error)

// CurrentSpaceIndexFunc returns the frontmost 0-based user Desktop index
// (same dense numbering as save / SpaceIndexForWindow).
type CurrentSpaceIndexFunc func() (int, error)

// SpaceBackend is Create / Switch / Highest for restore placement.
// Matches space.Backend (Create/Switch/List/Highest); tests typically inject
// space.MockBackend.
type SpaceBackend = space.Backend

var (
	spaceHookMu sync.Mutex

	// spaceIndexForWindowFn resolves window id → space index (injectable).
	spaceIndexForWindowFn SpaceIndexResolver = defaultSpaceIndexForWindow

	// currentSpaceIndexFn resolves the frontmost Desktop (injectable).
	currentSpaceIndexFn CurrentSpaceIndexFunc = defaultCurrentSpaceIndex

	// spaceBackend is optional inject for Create/Switch/Highest (tests).
	// When nil, production uses live space package wrappers.
	spaceBackend SpaceBackend
)

func defaultSpaceIndexForWindow(windowID uint64) (int, error) {
	return space.SpaceIndexForWindow(windowID)
}

// defaultCurrentSpaceIndex uses CGS managed displays (no Mission Control UI).
func defaultCurrentSpaceIndex() (int, error) {
	spaces, err := space.ListUserSpaces()
	if err != nil {
		return 0, err
	}
	for _, s := range spaces {
		if s.Current {
			return s.Index, nil
		}
	}
	return 0, fmt.Errorf("space: no current user Desktop")
}

// SetSpaceIndexForWindowForTest installs a Space index resolver. Pass nil to restore.
// Tests should t.Cleanup restore.
func SetSpaceIndexForWindowForTest(fn SpaceIndexResolver) {
	spaceHookMu.Lock()
	defer spaceHookMu.Unlock()
	if fn == nil {
		spaceIndexForWindowFn = defaultSpaceIndexForWindow
		return
	}
	spaceIndexForWindowFn = fn
}

// SetSpaceBackendForTest installs a Space Backend for restore placement.
// Pass nil to restore live path. Tests should t.Cleanup restore.
func SetSpaceBackendForTest(b SpaceBackend) {
	spaceHookMu.Lock()
	defer spaceHookMu.Unlock()
	spaceBackend = b
}

// SetCurrentSpaceIndexForTest installs a current-Desktop resolver.
// Pass nil to restore the live ListUserSpaces path. Tests should t.Cleanup restore.
func SetCurrentSpaceIndexForTest(fn CurrentSpaceIndexFunc) {
	spaceHookMu.Lock()
	defer spaceHookMu.Unlock()
	if fn == nil {
		currentSpaceIndexFn = defaultCurrentSpaceIndex
		return
	}
	currentSpaceIndexFn = fn
}

// SetSpaceSwitchSettleForTest overrides the delay between Switch retries.
// Pass a negative duration to restore the default. Tests should t.Cleanup restore.
func SetSpaceSwitchSettleForTest(d time.Duration) {
	spaceSwitchSettleMu.Lock()
	defer spaceSwitchSettleMu.Unlock()
	if d < 0 {
		spaceSwitchSettle = 500 * time.Millisecond
		return
	}
	spaceSwitchSettle = d
}

func currentSpaceSwitchSettle() time.Duration {
	spaceSwitchSettleMu.Lock()
	defer spaceSwitchSettleMu.Unlock()
	return spaceSwitchSettle
}

func currentSpaceIndexResolver() SpaceIndexResolver {
	spaceHookMu.Lock()
	defer spaceHookMu.Unlock()
	return spaceIndexForWindowFn
}

func currentSpaceBackend() SpaceBackend {
	spaceHookMu.Lock()
	defer spaceHookMu.Unlock()
	if spaceBackend != nil {
		return spaceBackend
	}
	return liveSpaceBackend{}
}

func currentSpaceIndexResolverFn() CurrentSpaceIndexFunc {
	spaceHookMu.Lock()
	defer spaceHookMu.Unlock()
	return currentSpaceIndexFn
}

// alreadyOnSpace reports whether the frontmost Desktop is spaceIdx (0-based).
// On lookup failure returns false so callers fall through to Switch.
func alreadyOnSpace(spaceIdx int) bool {
	fn := currentSpaceIndexResolverFn()
	cur, err := fn()
	if err != nil {
		return false
	}
	return cur == spaceIdx
}

// liveSpaceBackend wraps the production space package.
type liveSpaceBackend struct{}

func (liveSpaceBackend) Create() error { return space.Create(nil) }
func (liveSpaceBackend) Switch(n int) error {
	return space.Switch(n, nil)
}
func (liveSpaceBackend) List() ([]space.Desktop, error) { return space.List(nil) }
func (liveSpaceBackend) Highest() (int, error)          { return space.Highest(nil) }

// resolveSpaceForWindow maps a snapshot window to (space, iterm_window_id, warning).
// On missing id or resolve failure: space=0 + warning; still returns iterm id when known.
// FixedSpace (fixture) wins when set — no resolver call.
func resolveSpaceForWindow(win SnapshotWindow) (spaceIdx int, itermID int64, warn string) {
	if win.WindowID != 0 {
		itermID = int64(win.WindowID)
	}
	if win.FixedSpace != nil {
		idx := *win.FixedSpace
		if idx < 0 {
			return 0, itermID, fmt.Sprintf("invalid macOS Space index %d; using space 0", idx)
		}
		return idx, itermID, ""
	}
	if itermID == 0 {
		return 0, 0, "could not resolve macOS Space (missing iterm window id); using space 0"
	}
	fn := currentSpaceIndexResolver()
	idx, err := fn(uint64(itermID))
	if err != nil {
		return 0, itermID, fmt.Sprintf("could not resolve macOS Space for window id %d: %v; using space 0", itermID, err)
	}
	if idx < 0 {
		return 0, itermID, fmt.Sprintf("invalid macOS Space index %d for window id %d; using space 0", idx, itermID)
	}
	return idx, itermID, ""
}

// attachSpaceFields fills Space / ItermWindowID on a critical SaveWindow.
// When ignore is true, marks the window to omit both fields on marshal (no resolve).
// Returns an optional warning (without "warning:" prefix).
func attachSpaceFields(sw *SaveWindow, win SnapshotWindow, ignore bool) string {
	if sw == nil {
		return ""
	}
	if ignore {
		sw.noSpaceRecord = true
		sw.Space = 0
		sw.ItermWindowID = 0
		return ""
	}
	sw.noSpaceRecord = false
	idx, id, warn := resolveSpaceForWindow(win)
	sw.Space = idx
	sw.ItermWindowID = id
	return warn
}

// clampSpaceIndex applies restore rules: missing already 0; s>=16 → 0 + warn.
func clampSpaceIndex(s int) (clamped int, warn string) {
	if s >= maxRecordedSpace {
		return 0, fmt.Sprintf("invalid space index %d (>= %d); using space 0 (Desktop 1)", s, maxRecordedSpace)
	}
	if s < 0 {
		return 0, fmt.Sprintf("invalid space index %d; using space 0 (Desktop 1)", s)
	}
	return s, ""
}

// formatSpaceDesktopLabel returns "space N (Desktop N+1)".
func formatSpaceDesktopLabel(spaceIdx int) string {
	return fmt.Sprintf("space %d (Desktop %d)", spaceIdx, spaceIdx+1)
}

// ensureSpacePlacement switches (and creates if needed) so Desktop (spaceIdx+1)
// is frontmost. spaceIdx is 0-based after clamp.
//
// Rules:
//   - If already on space s (CGS current), skip Highest/Create/Switch.
//   - s==0: Switch(Desktop 1) when not already there.
//   - s>0: Create until Highest >= s+1; on max-cap fail → warn, Switch(1);
//     else Switch(s+1)
//   - Switch is retried on transient Mission Control AX errors; after retries
//     still fail → warn and continue on the current Desktop (soft placement).
//
// Warnings are returned for soft placement fallbacks (caller prints them).
// Hard errors remain for Highest/Create failures (except max-Desktop Create).
func ensureSpacePlacement(spaceIdx int) (warnings []string, err error) {
	s, clampWarn := clampSpaceIndex(spaceIdx)
	if clampWarn != "" {
		warnings = append(warnings, clampWarn)
	}
	if alreadyOnSpace(s) {
		return warnings, nil
	}
	b := currentSpaceBackend()
	if s == 0 {
		if w := switchToSpaceSoft(b, 0); w != "" {
			warnings = append(warnings, w)
		}
		return warnings, nil
	}
	desktop := s + 1
	for {
		h, herr := b.Highest()
		if herr != nil {
			return warnings, fmt.Errorf("highest Desktop: %w", herr)
		}
		if h >= desktop {
			break
		}
		if cerr := b.Create(); cerr != nil {
			if errors.Is(cerr, space.ErrMaxDesktops) {
				warnings = append(warnings,
					fmt.Sprintf("cannot create Desktop %d (at macOS max); using space 0 (Desktop 1)", desktop))
				if w := switchToSpaceSoft(b, 0); w != "" {
					warnings = append(warnings, w)
				}
				return warnings, nil
			}
			return warnings, fmt.Errorf("create Desktop: %w", cerr)
		}
	}
	if w := switchToSpaceSoft(b, s); w != "" {
		warnings = append(warnings, w)
	}
	return warnings, nil
}

// switchToSpaceSoft switches to 0-based spaceIdx when not already there.
// Skips Mission Control Switch if CGS reports we are already on that Desktop.
func switchToSpaceSoft(b SpaceBackend, spaceIdx int) string {
	if alreadyOnSpace(spaceIdx) {
		return ""
	}
	return switchDesktopSoft(b, spaceIdx+1)
}

// switchDesktopSoft retries Switch on transient AX errors, then soft-falls back
// to the current Desktop with a warning (empty string on success).
func switchDesktopSoft(b SpaceBackend, desktop int) string {
	if err := switchDesktopWithRetry(b, desktop); err != nil {
		return fmt.Sprintf(
			"could not switch to Desktop %d after %d attempts: %v; using current Desktop",
			desktop, spaceSwitchRetries, err)
	}
	return ""
}

// switchDesktopWithRetry calls Backend.Switch up to spaceSwitchRetries times.
// Non-transient errors return immediately; transient errors settle and retry.
func switchDesktopWithRetry(b SpaceBackend, desktop int) error {
	var last error
	for attempt := 0; attempt < spaceSwitchRetries; attempt++ {
		last = b.Switch(desktop)
		if last == nil {
			return nil
		}
		if !isTransientSpacePlacementError(last) || attempt+1 == spaceSwitchRetries {
			return last
		}
		if d := currentSpaceSwitchSettle(); d > 0 {
			time.Sleep(d)
		}
	}
	return last
}

// isTransientSpacePlacementError mirrors space.isTransientSpaceError for retry
// decisions without importing unexported helpers from dot-pkgs.
func isTransientSpacePlacementError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	switch {
	case strings.Contains(s, "desktop not found"):
		return true
	case strings.Contains(s, "Invalid index"):
		return true
	case strings.Contains(s, "-1719"):
		return true
	case strings.Contains(s, "no Desktop buttons found"):
		return true
	case strings.Contains(s, "can't get group") || strings.Contains(s, "Can't get group"):
		return true
	case strings.Contains(s, "Mission Control"):
		return strings.Contains(s, "FAIL:")
	default:
		return false
	}
}
