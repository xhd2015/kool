// Package codex implements the kool codex subcommand (install / ensure Codex CLI).
package codex

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	codexinstall "github.com/xhd2015/dot-pkgs/go-pkgs/shell/codex/install"
	"github.com/xhd2015/kool/pkgs/errs"
	lessflags "github.com/xhd2015/less-flags"
)

const help = `codex — OpenAI Codex CLI helpers

Usage:
  kool codex install [flags]
  kool codex install -h|--help
  kool codex -h|--help

Commands:
  install    ensure Codex CLI is installed and up to date

Options:
  -h, --help    show this help

Run kool codex install --help for install flags.
`

const installHelp = `codex install — ensure OpenAI Codex CLI is installed/updated

Usage:
  kool codex install [flags]
  kool codex install -h|--help

Ensures the Codex CLI is present and current via
github.com/xhd2015/dot-pkgs/go-pkgs/shell/codex/install:

  missing  → install (official curl | sh recipe)
  outdated → update  (codex update)
  current  → noop

Flags:
  --dry-run         print plan only (no install/update shell)
  --check-update    report local vs latest status (exit 0; no shell)
  -h, --help        show this help

Examples:
  kool codex install
  kool codex install --dry-run
  kool codex install --check-update
`

// Deps are injectable for L2 doctests (nil fields → production defaults).
type Deps struct {
	LookPath    func(file string) (string, error)
	RunShell    func(ctx context.Context, cmd string) error
	RunVersion  func(ctx context.Context, bin string) (string, error)
	FetchLatest func(ctx context.Context) (string, error)
}

// packageDeps is process-global; tests serialize via SetDepsForTest + mutex.
var packageDeps Deps

// SetDepsForTest installs package-level deps for one test invocation.
// Returns restore; callers should defer restore under a process mutex.
func SetDepsForTest(d Deps) (restore func()) {
	prev := packageDeps
	packageDeps = d
	return func() {
		packageDeps = prev
	}
}

// Handle runs the kool codex subcommand.
func Handle(args []string) error {
	return run(args, os.Stdout, os.Stderr)
}

// RunForTest runs the codex handler in-process (args after "kool codex").
// Example: RunForTest([]string{"install", "--dry-run"}, stdout, stderr)
func RunForTest(args []string, stdout, stderr io.Writer) int {
	if err := run(args, stdout, stderr); err != nil {
		if se, ok := errs.IsSilenceExitCode(err); ok {
			return se.SilenceExitCode()
		}
		msg := err.Error()
		if msg != "" {
			fmt.Fprint(stderr, msg)
			if !strings.HasSuffix(msg, "\n") {
				fmt.Fprintln(stderr)
			}
		}
		return 1
	}
	return 0
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stdout, strings.TrimSpace(help)+"\n")
		return nil
	}

	switch args[0] {
	case "install":
		return runInstall(args[1:], stdout, stderr)
	case "help", "-h", "--help":
		fmt.Fprint(stdout, strings.TrimSpace(help)+"\n")
		return nil
	default:
		// Allow root -h/--help without a subcommand.
		remain, err := lessflags.
			HelpFunc("-h,--help", func() {}).
			HelpNoExit().
			Parse(args)
		if err != nil {
			if err == lessflags.ErrHelp {
				fmt.Fprint(stdout, strings.TrimSpace(help)+"\n")
				return nil
			}
			fmt.Fprint(stderr, err.Error())
			return errs.NewSilenceExitCode(1)
		}
		if len(remain) == 0 {
			fmt.Fprint(stdout, strings.TrimSpace(help)+"\n")
			return nil
		}
		fmt.Fprintf(stderr, "unrecognized command: %s\nRun 'kool codex --help' for usage.\n", remain[0])
		return errs.NewSilenceExitCode(1)
	}
}

func runInstall(args []string, stdout, stderr io.Writer) error {
	var dryRun bool
	var checkUpdate bool

	remain, err := lessflags.Bool("--dry-run", &dryRun).
		Bool("--check-update", &checkUpdate).
		HelpFunc("-h,--help", func() {}).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if err == lessflags.ErrHelp {
			fmt.Fprint(stdout, strings.TrimSpace(installHelp)+"\n")
			return nil
		}
		fmt.Fprint(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}
	if len(remain) > 0 {
		fmt.Fprintf(stderr, "Error: unexpected arguments: %s\nRun 'kool codex install --help' for usage.\n", strings.Join(remain, " "))
		return errs.NewSilenceExitCode(1)
	}

	ctx := context.Background()

	if checkUpdate {
		return runCheckUpdate(ctx, stdout, stderr)
	}
	if dryRun {
		return runDryRun(ctx, stdout, stderr)
	}
	return runEnsure(ctx, stdout, stderr)
}

func lookPathFn() func(file string) (string, error) {
	if packageDeps.LookPath != nil {
		return packageDeps.LookPath
	}
	return exec.LookPath
}

func runShellFn(stdout, stderr io.Writer) func(ctx context.Context, cmd string) error {
	if packageDeps.RunShell != nil {
		return packageDeps.RunShell
	}
	return func(ctx context.Context, cmd string) error {
		c := exec.CommandContext(ctx, "sh", "-c", cmd)
		c.Stdout = stdout
		c.Stderr = stderr
		return c.Run()
	}
}

func runVersionFn() func(ctx context.Context, bin string) (string, error) {
	if packageDeps.RunVersion != nil {
		return packageDeps.RunVersion
	}
	return nil // library default
}

func fetchLatestFn() func(ctx context.Context) (string, error) {
	if packageDeps.FetchLatest != nil {
		return packageDeps.FetchLatest
	}
	return nil // library default
}

// statusProbe mirrors Ensure decisioning without shell mutation.
// When bin is missing it never calls FetchLatest.
type statusProbe struct {
	Present       bool
	BinPath       string
	LocalVersion  string
	LatestVersion string
	NeedsUpdate   bool
	// Action: install | update | noop | missing
	Action string
}

func probeStatus(ctx context.Context) (statusProbe, error) {
	var s statusProbe
	look := lookPathFn()
	binPath, err := look("codex")
	if err != nil {
		s.Present = false
		s.Action = "missing"
		return s, nil
	}
	s.Present = true
	s.BinPath = binPath

	runVer := runVersionFn()
	if runVer == nil {
		// Match library default path: bin --version.
		runVer = func(ctx context.Context, bin string) (string, error) {
			c := exec.CommandContext(ctx, bin, "--version")
			out, err := c.Output()
			if err != nil {
				return "", err
			}
			return strings.TrimSpace(string(out)), nil
		}
	}
	rawLocal, localErr := runVer(ctx, binPath)
	if localErr == nil {
		if parsed, perr := codexinstall.ParseVersion(rawLocal); perr == nil {
			s.LocalVersion = parsed
		} else {
			s.LocalVersion = strings.TrimSpace(rawLocal)
		}
	}

	fetch := fetchLatestFn()
	if fetch == nil {
		fetch = func(ctx context.Context) (string, error) {
			return codexinstall.LatestVersion(ctx, codexinstall.LatestVersionOpts{})
		}
	}
	latest, latestErr := fetch(ctx)
	if latestErr != nil {
		// Soft: treat as noop / unknown latest.
		s.Action = "noop"
		return s, nil
	}
	s.LatestVersion = latest

	if localErr != nil {
		s.Action = "noop"
		return s, nil
	}

	localForCompare := s.LocalVersion
	if localForCompare == "" {
		localForCompare = rawLocal
	}
	s.NeedsUpdate = codexinstall.NeedsUpdate(localForCompare, latest)
	if s.NeedsUpdate {
		s.Action = "update"
	} else {
		s.Action = "noop"
	}
	return s, nil
}

func runDryRun(ctx context.Context, stdout, stderr io.Writer) error {
	s, err := probeStatus(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return errs.NewSilenceExitCode(1)
	}
	switch s.Action {
	case "missing":
		fmt.Fprintf(stdout, "dry-run: would install codex\n")
		fmt.Fprintf(stdout, "  cmd: %s\n", codexinstall.InstallCmd)
	case "update":
		fmt.Fprintf(stdout, "dry-run: would update codex (%s → %s)\n", s.LocalVersion, s.LatestVersion)
		fmt.Fprintf(stdout, "  cmd: %s\n", codexinstall.UpdateCmd)
	default: // noop
		ver := s.LocalVersion
		if ver == "" {
			ver = s.LatestVersion
		}
		if ver != "" {
			fmt.Fprintf(stdout, "dry-run: noop — codex %s is up to date\n", ver)
		} else {
			fmt.Fprintf(stdout, "dry-run: noop — codex is up to date\n")
		}
	}
	return nil
}

func runCheckUpdate(ctx context.Context, stdout, stderr io.Writer) error {
	s, err := probeStatus(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return errs.NewSilenceExitCode(1)
	}
	switch s.Action {
	case "missing":
		fmt.Fprintln(stdout, "codex: missing (not installed)")
	case "update":
		fmt.Fprintf(stdout, "codex: update available (%s → %s)\n", s.LocalVersion, s.LatestVersion)
	default:
		ver := s.LocalVersion
		if ver == "" {
			ver = s.LatestVersion
		}
		if ver != "" {
			fmt.Fprintf(stdout, "codex: up to date (%s)\n", ver)
		} else {
			fmt.Fprintln(stdout, "codex: up to date")
		}
	}
	return nil
}

func runEnsure(ctx context.Context, stdout, stderr io.Writer) error {
	opts := codexinstall.EnsureOpts{
		LookPath:    lookPathFn(),
		RunShell:    runShellFn(stdout, stderr),
		RunVersion:  runVersionFn(),
		FetchLatest: fetchLatestFn(),
		Stdout:      stdout,
		Stderr:      stderr,
	}
	result, err := codexinstall.Ensure(ctx, opts)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return errs.NewSilenceExitCode(1)
	}
	switch result.Action {
	case "install":
		fmt.Fprintln(stdout, "codex: installed")
	case "update":
		fmt.Fprintf(stdout, "codex: updated (%s → %s)\n", result.LocalVersion, result.LatestVersion)
	default:
		if result.LocalVersion != "" {
			fmt.Fprintf(stdout, "codex: up to date (%s)\n", result.LocalVersion)
		} else {
			fmt.Fprintln(stdout, "codex: up to date")
		}
	}
	return nil
}
