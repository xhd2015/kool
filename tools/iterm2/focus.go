package iterm2

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	lib "github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

const focusHelp = `Usage: kool iterm2 focus <dir> [--index N]
  --index N   select candidate N when multiple sessions match
`

// FocusCandidate identifies one candidate session. Its order is presented to
// users as the stable zero-based --index order.
type FocusCandidate struct {
	WindowID      string
	WindowTitle   string
	TabIndex      int
	SessionID     string
	Path          string
	KoolTargetDir string
}

type focusBoundary interface {
	Discover(string) ([]FocusCandidate, error)
	Focus(FocusCandidate) error
}

// FocusFake is the deterministic injected iTerm boundary used by L2 callers.
type FocusFake struct {
	Candidates    []FocusCandidate
	Focused       []string
	DiscoverCalls int
}

func (f *FocusFake) Discover(_ string) ([]FocusCandidate, error) {
	f.DiscoverCalls++
	return append([]FocusCandidate(nil), f.Candidates...), nil
}
func (f *FocusFake) Focus(candidate FocusCandidate) error {
	f.Focused = append(f.Focused, candidate.SessionID)
	return nil
}

type liveFocusBoundary struct{}

func (liveFocusBoundary) Discover(target string) ([]FocusCandidate, error) {
	candidates, err := lib.FindDirectoryFocusCandidates(target)
	if err != nil {
		return nil, err
	}
	out := make([]FocusCandidate, len(candidates))
	for i, c := range candidates {
		out[i] = FocusCandidate{WindowID: c.WindowID, WindowTitle: c.WindowTitle, TabIndex: c.TabIndex, SessionID: c.SessionID, Path: c.Path, KoolTargetDir: c.KoolTargetDir}
	}
	return out, nil
}
func (liveFocusBoundary) Focus(c FocusCandidate) error {
	return lib.FocusDirectoryCandidate(lib.DirectoryFocusCandidate{WindowID: c.WindowID, WindowTitle: c.WindowTitle, TabIndex: c.TabIndex, SessionID: c.SessionID, Path: c.Path, KoolTargetDir: c.KoolTargetDir})
}

// RunFocusForTest runs the iterm2 focus route against an injected boundary.
// args are the full arguments after "kool iterm2", mirroring the public
// handler rather than requiring callers to strip the focus subcommand.
func RunFocusForTest(args []string, stdout, stderr io.Writer, fake *FocusFake) int {
	if len(args) > 0 && args[0] == "focus" {
		return runFocus(args[1:], stdout, stderr, fake)
	}
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		// Keep this injectable seam narrowly scoped to the focus route.
		fmt.Fprint(stdout, "  iterm2 focus <dir> [--index N]\n")
		return 0
	}
	return focusError(stderr, "expected focus subcommand")
}

func runFocus(args []string, stdout, stderr io.Writer, boundary focusBoundary) int {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Fprint(stdout, focusHelp)
		return 0
	}
	var index *int
	var positional []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--index" {
			if i+1 >= len(args) {
				return focusError(stderr, "--index requires an integer")
			}
			n, err := strconv.Atoi(args[i+1])
			if err != nil {
				return focusError(stderr, "--index must be an integer")
			}
			index = &n
			i++
			continue
		}
		if strings.HasPrefix(args[i], "--index=") {
			n, err := strconv.Atoi(strings.TrimPrefix(args[i], "--index="))
			if err != nil {
				return focusError(stderr, "--index must be an integer")
			}
			index = &n
			continue
		}
		positional = append(positional, args[i])
	}
	if len(positional) == 0 {
		fmt.Fprint(stderr, "Error: missing directory argument\nRun 'kool iterm2 focus --help' for usage.\n")
		return 1
	}
	if len(positional) > 1 {
		return focusError(stderr, "unexpected arguments: "+strings.Join(positional[1:], " "))
	}
	target, err := canonicalDirectory(positional[0])
	if err != nil {
		return focusError(stderr, err.Error())
	}
	candidates, err := boundary.Discover(target)
	if err != nil {
		return focusError(stderr, err.Error())
	}
	if len(candidates) == 0 {
		return focusError(stderr, "no iTerm2 session found for: "+target)
	}
	if index == nil && len(candidates) > 1 {
		fmt.Fprintf(stderr, "Error: multiple iTerm2 sessions found for: %s\n\n", target)
		printFocusCandidates(stderr, candidates)
		fmt.Fprintf(stderr, "\nSpecify one with:\n  kool iterm2 focus %s --index <%s>\n", target, validIndexes(candidates))
		return 1
	}
	selected := 0
	if index != nil {
		selected = *index
		if selected < 0 || selected >= len(candidates) {
			fmt.Fprintf(stderr, "Error: --index %d is out of range (valid indexes: %s)\n\n", selected, validIndexes(candidates))
			printFocusCandidates(stderr, candidates)
			return 1
		}
	}
	if err := boundary.Focus(candidates[selected]); err != nil {
		return focusError(stderr, err.Error())
	}
	fmt.Fprintf(stdout, "focused: window %s, tab %d\n", candidates[selected].WindowID, candidates[selected].TabIndex)
	return 0
}

func canonicalDirectory(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("not a directory: %s", abs)
		}
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("not a directory: %s", abs)
	}
	if real, err := filepath.EvalSymlinks(abs); err == nil {
		return real, nil
	}
	return abs, nil
}
func focusError(stderr io.Writer, message string) int {
	fmt.Fprintf(stderr, "Error: %s\n", message)
	return 1
}
func printFocusCandidates(w io.Writer, candidates []FocusCandidate) {
	for i, c := range candidates {
		fmt.Fprintf(w, "  [%d] window %s (%q) tab %d session %s\n", i, c.WindowID, c.WindowTitle, c.TabIndex, c.SessionID)
	}
}
func validIndexes(candidates []FocusCandidate) string {
	indexes := make([]string, len(candidates))
	for i := range candidates {
		indexes[i] = strconv.Itoa(i)
	}
	return strings.Join(indexes, "|")
}
