package iterm2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

// ForegroundCommand is command evidence, not an instruction to replay. Observed
// is display-only ps text. Only Argv contains exact argument boundaries. Complex
// or Reason nonempty requires review; unavailable argv/cwd must never be guessed.
type ForegroundCommand struct {
	PID      int      `json:"pid"`
	Cwd      string   `json:"cwd,omitempty"`
	Argv     []string `json:"argv,omitempty"`
	Observed string   `json:"observed,omitempty"`
	Reason   string   `json:"reason,omitempty"`
	Complex  bool     `json:"complex,omitempty"`
}

// resolveForegroundCommand is pure: '+' identifies the terminal foreground
// process group, and ancestry identifies its launcher rather than its leaf.
func resolveForegroundCommand(s SnapshotSession) *ForegroundCommand {
	byPID := make(map[int]SnapshotProc, len(s.Processes))
	for _, p := range s.Processes {
		byPID[p.PID] = p
	}
	shell := 0
	if s.ShellPID != nil {
		shell = *s.ShellPID
	}
	roots := map[int]SnapshotProc{}
	var observed []string
	// If the session shell is itself in the foreground group, job control is
	// absent or the snapshot raced. Do not mistake shared-group background work
	// for an independently launched foreground job.
	shellProc, shellKnown := byPID[shell]
	uncertain := !shellKnown || strings.Contains(shellProc.Stat, "+")
	for _, p := range s.Processes {
		if !strings.Contains(p.Stat, "+") || p.PID == shell {
			continue
		}
		observed = append(observed, p.Command)
		if shell == 0 {
			roots[p.PID] = p
			uncertain = true
			continue
		}
		seen := map[int]bool{}
		root := p
		for root.PPID != shell {
			if seen[root.PID] {
				uncertain = true
				break
			}
			seen[root.PID] = true
			parent, ok := byPID[root.PPID]
			if !ok || parent.PID == root.PID {
				uncertain = true
				break
			}
			root = parent
		}
		if !strings.Contains(root.Stat, "+") {
			uncertain = true
		}
		roots[root.PID] = root
	}
	if len(roots) == 0 {
		return nil
	}
	if len(roots) != 1 {
		return &ForegroundCommand{Observed: strings.Join(observed, " | "), Complex: true, Reason: "multiple foreground launchers (possible pipeline)"}
	}
	var root SnapshotProc
	for _, p := range roots {
		root = p
	}
	out := &ForegroundCommand{PID: root.PID, Observed: root.Command}
	if uncertain {
		out.Complex = true
		out.Reason = "foreground launcher ancestry is incomplete or ambiguous"
	}
	return out
}

// foregroundProcessIdentity is comparable and includes the kernel's subsecond
// birth time, so a recycled PID cannot mix argv and cwd from different processes.
type foregroundProcessIdentity struct {
	PID, PPID, PGID, TPGID, TTY int
	StartSec, StartUsec         int64
}

type foregroundCommandReads struct {
	argv     func(int) ([]string, error)
	cwd      func(int) (string, error)
	identity func(int) (foregroundProcessIdentity, error)
}

func enrichForegroundCommands(win *SnapshotWindow) {
	enrichForegroundCommandsWith(win, foregroundCommandReads{argv: readForegroundArgv, cwd: readForegroundCwd, identity: readForegroundIdentity})
}

func enrichForegroundCommandsWith(win *SnapshotWindow, reads foregroundCommandReads) {
	if win == nil {
		return
	}
	for ti := range win.Tabs {
		for si := range win.Tabs[ti].Sessions {
			s := &win.Tabs[ti].Sessions[si]
			s.Foreground = nil
			if s.Agent != nil && s.Agent.SessionID != "" && (s.Agent.Kind == "grok" || s.Agent.Kind == "codex") {
				continue
			}
			if _, ok := detectMarkSession(s); ok {
				continue
			}
			f := resolveForegroundCommand(*s)
			s.Foreground = f
			if f == nil || f.Complex {
				continue
			}
			var before foregroundProcessIdentity
			if reads.identity != nil {
				var err error
				before, err = reads.identity(f.PID)
				if err != nil || !foregroundIdentityMatches(*s, f.PID, before) {
					f.Complex = true
					f.Reason = "foreground launcher identity changed or unavailable"
					continue
				}
			}
			argv, err := reads.argv(f.PID)
			if err != nil || len(argv) == 0 || argv[0] == "" {
				f.Reason = "exact foreground argv unavailable"
			} else {
				if foregroundBareShell(argv) {
					// Nested interactive shells are not a command to save either.
					s.Foreground = nil
					continue
				}
				f.Argv = append([]string(nil), argv...)
			}
			cwd, err := reads.cwd(f.PID)
			if err != nil || cwd == "" {
				if f.Reason != "" {
					f.Reason += "; "
				}
				f.Reason += "foreground launcher cwd unavailable"
			} else {
				f.Cwd = cwd
			}
			if reads.identity != nil {
				// exec preserves a PID's birth time, so also check exact argv on
				// both sides of cwd collection before accepting replay evidence.
				changed := false
				if len(f.Argv) > 0 {
					again, err := reads.argv(f.PID)
					changed = err != nil || !slices.Equal(f.Argv, again)
				}
				after, err := reads.identity(f.PID)
				if err != nil || before != after || changed {
					f.Argv, f.Cwd = nil, ""
					f.Complex = true
					f.Reason = "foreground launcher changed during evidence collection"
				}
			}
		}
	}
}

func foregroundIdentityMatches(s SnapshotSession, pid int, identity foregroundProcessIdentity) bool {
	if identity.PID != pid || identity.StartSec <= 0 || identity.PGID <= 0 || identity.PGID != identity.TPGID {
		return false
	}
	for _, p := range s.Processes {
		if p.PID != pid {
			continue
		}
		if identity.PPID != p.PPID {
			return false
		}
		// lstart has second precision; the dependency can also derive it from
		// elapsed seconds, so allow one second of rounding, never a new birth.
		if p.StartTimeUnix != nil {
			delta := identity.StartSec - *p.StartTimeUnix
			if delta < -1 || delta > 1 {
				return false
			}
		}
		return true
	}
	return false
}

// foregroundBareShell uses only exact argv; ps display strings are never split.
func foregroundBareShell(argv []string) bool {
	if len(argv) == 0 {
		return false
	}
	switch strings.TrimPrefix(filepath.Base(argv[0]), "-") {
	case "sh", "bash", "zsh", "fish", "ksh", "tcsh", "csh", "dash":
	default:
		return false
	}
	for _, arg := range argv[1:] {
		switch arg {
		case "-i", "-l", "-il", "-li", "--login", "--interactive":
		default:
			return false
		}
	}
	return true
}

// parseForegroundProcArgs parses Darwin KERN_PROCARGS2: argc, executable path,
// NUL padding, then exactly argc NUL-terminated arguments. Environment entries
// after argv are deliberately never decoded or retained.
func parseForegroundProcArgs(data []byte) ([]string, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("short procargs header")
	}
	argc := int(binary.NativeEndian.Uint32(data[:4]))
	if argc <= 0 || argc > len(data)-4 {
		return nil, fmt.Errorf("invalid procargs argc")
	}
	rest := data[4:]
	end := bytes.IndexByte(rest, 0)
	if end < 0 {
		return nil, fmt.Errorf("unterminated executable path")
	}
	rest = rest[end+1:]
	for len(rest) > 0 && rest[0] == 0 {
		rest = rest[1:]
	}
	argv := make([]string, 0, argc)
	for i := 0; i < argc; i++ {
		end = bytes.IndexByte(rest, 0)
		if end < 0 {
			return nil, fmt.Errorf("truncated procargs argv")
		}
		argv = append(argv, string(rest[:end]))
		rest = rest[end+1:]
	}
	return argv, nil
}

// The dependency snapshot model intentionally has no kool-specific evidence.
// Restore only explicit fixture evidence after crossing that model boundary.
func (c *SnapshotCollector) enrichWindowForeground(win *SnapshotWindow) {
	if !c.fixtureEnabled {
		enrichForegroundCommands(win)
		return
	}
	for _, fw := range c.fixtureWindows {
		if fw.Index != win.Index {
			continue
		}
		for ti := range win.Tabs {
			for si := range win.Tabs[ti].Sessions {
				s := &win.Tabs[ti].Sessions[si]
				for _, ft := range fw.Tabs {
					for _, fs := range ft.Sessions {
						if (s.ID != "" && s.ID == fs.ID) || (s.ID == "" && win.Tabs[ti].Index == ft.Index && s.Index == fs.Index) {
							if fs.Foreground != nil {
								f := *fs.Foreground
								f.Argv = append([]string(nil), f.Argv...)
								s.Foreground = &f
							}
						}
					}
				}
			}
		}
	}
}
