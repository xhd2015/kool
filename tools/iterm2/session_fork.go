package iterm2

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	lib "github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
	"github.com/xhd2015/kool/pkgs/errs"
	lessflags "github.com/xhd2015/less-flags"
)

const sessionForkHelp = `iterm2 session fork — fork a grok session in a new iTerm2 window

Usage: kool iterm2 session <session-id> fork [--dry-run]

Opens a new iTerm2 window at the pane's cwd and runs agent-run to branch the
resolved grok session via native grok --fork-session:

  agent-run run --agent-runner=grok-tty --dir=<cwd> \
    --resume-from-grok-session=<id> --fork --open

--fork skips agent-run's already-mapped import gate and always creates a new
agent-run session for the branch.

Requirements:
  - session must resolve to a unique live iTerm pane
  - pane must have a resolved grok agent session id (enrich)
  - pane cwd must be non-empty
  - agent-run must be available on PATH

Options:
  --dry-run     print plan only; do not open a window
  -h, --help    show this help

Session id: same as status (iTerm unique ID, UUID prefix ≥8, tty, or pid).

Examples:
  kool iterm2 session D922B298 fork --dry-run
  kool iterm2 session D922B298 fork
`

// ForkPlan is the resolved action for session fork (dry-run and live).
type ForkPlan struct {
	SourceItermID  string
	Cwd            string
	GrokSessionID  string
	FollowUp       string
	AgentRunBinary string
}

// Injectable for tests: open new terminal with follow-up (default: ForceNew + write text).
var sessionOpenInNewTerminal = defaultOpenInNewTerminal

// Injectable LookPath for agent-run (tests).
var sessionLookPath = exec.LookPath

// defaultOpenInNewTerminal opens a new iTerm2 window at dir and runs followUp
// (same ModeForceNew pattern as agent-run --new-terminal / agentrunapi.OpenInNewTerminal).
func defaultOpenInNewTerminal(dir, followUp string) error {
	return lib.OpenConfig(dir, &lib.Config{
		Mode:             lib.ModeForceNew,
		FollowUpCommands: []string{followUp},
	})
}

// BuildForkFollowUpCommand builds the agent-run line for ForceNew FollowUpCommands.
// Binary defaults to "agent-run". Does not include --new-terminal (parent opens the window).
func BuildForkFollowUpCommand(agentRunBinary, cwd, grokSessionID string) string {
	return buildForkFollowUp(agentRunBinary, cwd, grokSessionID)
}

func buildForkFollowUp(bin, cwd, grokSessionID string) string {
	if strings.TrimSpace(bin) == "" {
		bin = "agent-run"
	}
	// agent-run --fork → grok --resume <id> --fork-session (native fork).
	// Parent may already be mapped; --fork skips that gate and allocates a new
	// agent-run session. --open keeps the TTY interactive in the new window.
	tokens := []string{
		bin,
		"run",
		"--agent-runner=grok-tty",
		"--dir=" + cwd,
		"--resume-from-grok-session=" + grokSessionID,
		"--fork",
		"--open",
	}
	quoted := make([]string, 0, len(tokens))
	for _, t := range tokens {
		quoted = append(quoted, shellQuote(t))
	}
	return strings.Join(quoted, " ")
}

// shellQuote wraps s for POSIX shell if needed (single-quote safe).
func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	// Safe unquoted charset: alnum and common path/flag chars.
	safe := true
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			continue
		}
		switch r {
		case '-', '_', '.', '/', '=', ':', '+', '@', '%':
			continue
		default:
			safe = false
		}
	}
	if safe {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// BuildForkPlan validates a snapshot session for fork and builds the plan.
// Returns an error message suitable for WriteError (without "Error: " prefix).
func BuildForkPlan(s *SnapshotSession, agentRunBinary string) (*ForkPlan, error) {
	if s == nil {
		return nil, fmt.Errorf("session is nil")
	}
	if s.Agent == nil || strings.ToLower(strings.TrimSpace(s.Agent.Kind)) != "grok" || strings.TrimSpace(s.Agent.SessionID) == "" {
		kind := "none"
		if s.Agent != nil && s.Agent.Kind != "" {
			kind = s.Agent.Kind
		}
		return nil, fmt.Errorf("session %s is not a grok session (kind=%s); only grok fork is supported", shortID(s.ID), kind)
	}
	cwd := strings.TrimSpace(derefStr(s.Cwd))
	if cwd == "" {
		return nil, fmt.Errorf("session %s has empty cwd; cwd is required for fork", shortID(s.ID))
	}
	// Prefer absolute realpath for --dir stability.
	if abs, err := filepath.Abs(cwd); err == nil {
		cwd = abs
		if real, err := filepath.EvalSymlinks(cwd); err == nil {
			cwd = real
		}
	}
	bin := strings.TrimSpace(agentRunBinary)
	if bin == "" {
		bin = "agent-run"
	}
	return &ForkPlan{
		SourceItermID:  s.ID,
		Cwd:            cwd,
		GrokSessionID:  strings.TrimSpace(s.Agent.SessionID),
		FollowUp:       buildForkFollowUp(bin, cwd, strings.TrimSpace(s.Agent.SessionID)),
		AgentRunBinary: bin,
	}, nil
}

func runSessionFork(sessionRef string, args []string, stdout, stderr io.Writer) error {
	var dryRun bool
	remain, err := lessflags.Bool("--dry-run", &dryRun).
		HelpFunc("-h,--help", func() {}).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if err == lessflags.ErrHelp {
			fmt.Fprint(stdout, strings.TrimSpace(sessionForkHelp)+"\n")
			return nil
		}
		WriteError(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}
	if len(remain) > 0 {
		fmt.Fprintf(stderr, "Error: session fork: unexpected arguments: %s\n", strings.Join(remain, " "))
		return errs.NewSilenceExitCode(1)
	}

	snap, warnings, err := CaptureSnapshotWith(CaptureOpts{NoEnrich: false})
	if err != nil {
		WriteError(stderr, strings.TrimPrefix(err.Error(), "Error: "))
		return errs.NewSilenceExitCode(1)
	}
	for _, w := range warnings {
		WriteWarning(stderr, strings.TrimPrefix(w, "warning: "))
	}

	matches := FindSessionsByRef(snap, sessionRef)
	if len(matches) == 0 {
		WriteError(stderr, fmt.Sprintf("session not found: %s", sessionRef))
		return errs.NewSilenceExitCode(1)
	}
	if len(matches) > 1 {
		WriteError(stderr, fmt.Sprintf("ambiguous session id %q (matched %d); use full unique id", sessionRef, len(matches)))
		return errs.NewSilenceExitCode(1)
	}
	s := matches[0]

	// Resolve agent-run binary on PATH (live path); dry-run still prefers PATH
	// but falls back to bare name with a warning.
	bin, lookErr := sessionLookPath("agent-run")
	if lookErr != nil || strings.TrimSpace(bin) == "" {
		if dryRun {
			WriteWarning(stderr, "agent-run not found on PATH; plan uses bare agent-run")
			bin = "agent-run"
		} else {
			WriteError(stderr, "agent-run not found on PATH (install agent-run)")
			return errs.NewSilenceExitCode(1)
		}
	}

	plan, err := BuildForkPlan(s, bin)
	if err != nil {
		WriteError(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}

	if dryRun {
		return renderForkDryRun(stdout, plan)
	}

	if err := sessionOpenInNewTerminal(plan.Cwd, plan.FollowUp); err != nil {
		WriteError(stderr, fmt.Sprintf("session fork: open new window: %v", err))
		return errs.NewSilenceExitCode(1)
	}

	fmt.Fprintf(stdout, "Opened new window; agent-run will fork grok %s (--fork-session)\n", plan.GrokSessionID)
	return nil
}

func renderForkDryRun(w io.Writer, plan *ForkPlan) error {
	if plan == nil {
		return nil
	}
	fmt.Fprintln(w, "Would open new window and fork grok session via agent-run")
	fmt.Fprintf(w, "  iterm id:     %s\n", plan.SourceItermID)
	fmt.Fprintf(w, "  cwd:          %s\n", plan.Cwd)
	fmt.Fprintf(w, "  grok id:      %s\n", plan.GrokSessionID)
	fmt.Fprintf(w, "  follow-up:    %s\n", plan.FollowUp)
	return nil
}
