package iterm2

import (
	"errors"
	"fmt"
	"sync"

	"github.com/xhd2015/dot-pkgs/go-pkgs/computer-use/macos/space"
)

// maxRecordedSpace is the exclusive upper bound for a 0-based Space index
// (macOS hard-caps at 16 Desktops → indices 0..15).
const maxRecordedSpace = 16

// SpaceIndexResolver maps an iTerm/CG window id to a 0-based Desktop index.
type SpaceIndexResolver func(windowID uint64) (int, error)

// SpaceBackend is Create / Switch / Highest for restore placement.
// Matches space.Backend (Create/Switch/List/Highest); tests typically inject
// space.MockBackend.
type SpaceBackend = space.Backend

var (
	spaceHookMu sync.Mutex

	// spaceIndexForWindowFn resolves window id → space index (injectable).
	spaceIndexForWindowFn SpaceIndexResolver = defaultSpaceIndexForWindow

	// spaceBackend is optional inject for Create/Switch/Highest (tests).
	// When nil, production uses live space package wrappers.
	spaceBackend SpaceBackend
)

func defaultSpaceIndexForWindow(windowID uint64) (int, error) {
	return space.SpaceIndexForWindow(windowID)
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
//   - s==0: always Switch(Desktop 1)
//   - s>0: Create until Highest >= s+1; on max-cap fail → warn, Switch(1);
//     else Switch(s+1)
//
// Warnings are returned for soft placement fallbacks (caller prints them).
func ensureSpacePlacement(spaceIdx int) (warnings []string, err error) {
	s, clampWarn := clampSpaceIndex(spaceIdx)
	if clampWarn != "" {
		warnings = append(warnings, clampWarn)
	}
	b := currentSpaceBackend()
	if s == 0 {
		if err := b.Switch(1); err != nil {
			return warnings, fmt.Errorf("switch to Desktop 1: %w", err)
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
				if err := b.Switch(1); err != nil {
					return warnings, fmt.Errorf("switch to Desktop 1 after max Desktops: %w", err)
				}
				return warnings, nil
			}
			return warnings, fmt.Errorf("create Desktop: %w", cerr)
		}
	}
	if err := b.Switch(desktop); err != nil {
		return warnings, fmt.Errorf("switch to Desktop %d: %w", desktop, err)
	}
	return warnings, nil
}
