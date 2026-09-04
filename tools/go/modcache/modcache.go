package modcache

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/kool/pkgs/errs"
	lessflags "github.com/xhd2015/less-flags"
)

const rootHelp = `Usage: kool go modcache <inspect|prune|seed> [OPTIONS]

Inspect $GOMODCACHE and prune legacy (non-newest) versions of each module,
or seed the cache from a local tagged git module.

Commands:
  inspect            report module cache versions and reclaimable legacy copies
  prune              delete legacy versions of each module
  seed               download module@version into GOMODCACHE from this local git repo

Options (inspect/prune):
  --modcache PATH    module cache root (default: go env GOMODCACHE)
  --root PATH        live-set / upgrade hints via git repo scan (repeatable)
  --json             JSON on stdout
  --top N            max rows in inspect tables (default 20; 0 = all)
  --include-toolchain  include golang.org/toolchain in legacy/prune
  --dry-run          prune: print plan, do not delete
  --no-cache         do not read or write the git-repo-scan cache
  --cache-dir PATH   git-repo-scan cache root
  -h, --help         show help message

Run kool go modcache seed --help for seed options.
Progress for inspect/prune is printed on stderr as [n/total] stage markers.
`

const inspectHelp = `Usage: kool go modcache inspect [OPTIONS]

Report versions stored in $GOMODCACHE, grouped by module path. Non-newest
versions are legacy and would be deleted by prune. Inspect prints SAVE: as
the estimated space reclaimed if prune keeps the newest version of each module.
Progress is printed on stderr as [n/total] stage markers while the cache is sized.

Options:
  --modcache PATH    module cache root (default: go env GOMODCACHE)
  --root PATH        live-set / upgrade hints via git repo scan (repeatable)
  --json             JSON on stdout
  --top N            max rows in inspect tables (default 20; 0 = all)
  --include-toolchain  include golang.org/toolchain in legacy/prune
  --no-cache         do not read or write the git-repo-scan cache
  --cache-dir PATH   git-repo-scan cache root
  -h, --help         show help message
`

const pruneHelp = `Usage: kool go modcache prune [OPTIONS]

Delete legacy (non-newest) versions of each module from $GOMODCACHE.
Keeps the newest version of each module path. Does not delete
golang.org/toolchain unless --include-toolchain is set.

Options:
  --modcache PATH    module cache root (default: go env GOMODCACHE)
  --root PATH        warn about local go.mod/go.sum still on legacy versions
  --json             JSON on stdout
  --include-toolchain  include golang.org/toolchain in prune
  --dry-run          print plan, do not delete
  --no-cache         do not read or write the git-repo-scan cache
  --cache-dir PATH   git-repo-scan cache root
  -h, --help         show help message

Progress is printed on stderr as [n/total] stage markers while the cache is sized.
`

// Handle is production entry: HandleWith(args, HandleOpts{}).
func Handle(args []string) error {
	return HandleWith(args, HandleOpts{})
}

// HandleOpts injects writers for tests and production defaults.
type HandleOpts struct {
	Stdout io.Writer
	Stderr io.Writer
}

// HandleWith is the injectable entry used by doctests.
// args are the argv after "kool go modcache".
func HandleWith(args []string, opts HandleOpts) error {
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	if len(args) == 0 {
		fmt.Fprintln(stderr, "requires subcommand: inspect, prune, or seed (try --help)")
		return errs.NewSilenceExitCode(1)
	}

	if args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		fmt.Fprint(stdout, rootHelp)
		if !strings.HasSuffix(rootHelp, "\n") {
			fmt.Fprintln(stdout)
		}
		return nil
	}

	switch args[0] {
	case "inspect":
		return handleInspect(args[1:], stdout, stderr)
	case "prune":
		return handlePrune(args[1:], stdout, stderr)
	case "seed":
		return handleSeed(args[1:], stdout, stderr)
	default:
		return fail(stderr, "unknown subcommand: %s (try inspect, prune, seed, or --help)", args[0])
	}
}

type options struct {
	modcache         string
	roots            []string
	json             bool
	top              int
	includeToolchain bool
	dryRun           bool
	noCache          bool
	cacheDir         string
}

func handleInspect(args []string, stdout, stderr io.Writer) error {
	opts := options{top: 20}
	remain, err := lessflags.String("--modcache", &opts.modcache).
		StringSlice("--root", &opts.roots).
		Bool("--json", &opts.json).
		Int("--top", &opts.top).
		Bool("--include-toolchain", &opts.includeToolchain).
		Bool("--no-cache", &opts.noCache).
		String("--cache-dir", &opts.cacheDir).
		HelpFunc("-h,--help", func() {
			fmt.Fprint(stdout, inspectHelp)
			if !strings.HasSuffix(inspectHelp, "\n") {
				fmt.Fprintln(stdout)
			}
		}).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if err == lessflags.ErrHelp {
			return nil
		}
		return fail(stderr, "%s", err.Error())
	}
	if len(remain) > 0 {
		return fail(stderr, "unrecognized extra args: %s", strings.Join(remain, " "))
	}

	modcache, err := resolveModCache(opts.modcache)
	if err != nil {
		return fail(stderr, "%s", err.Error())
	}
	total := 3
	if len(opts.roots) > 0 {
		total = 4
	}
	prog := newStageProgress(stderr, total)
	inv, err := inventoryCache(modcache, prog)
	if err != nil {
		return fail(stderr, "%s", err.Error())
	}
	live, err := collectLiveSet(opts.roots, modcache, opts.noCache, opts.cacheDir, stderr, prog)
	if err != nil {
		return fail(stderr, "%s", err.Error())
	}
	rep := buildReport(inv, live, opts.includeToolchain)
	return renderInspect(stdout, stderr, rep, opts)
}

func handlePrune(args []string, stdout, stderr io.Writer) error {
	var opts options
	remain, err := lessflags.String("--modcache", &opts.modcache).
		StringSlice("--root", &opts.roots).
		Bool("--json", &opts.json).
		Bool("--include-toolchain", &opts.includeToolchain).
		Bool("--dry-run", &opts.dryRun).
		Bool("--no-cache", &opts.noCache).
		String("--cache-dir", &opts.cacheDir).
		HelpFunc("-h,--help", func() {
			fmt.Fprint(stdout, pruneHelp)
			if !strings.HasSuffix(pruneHelp, "\n") {
				fmt.Fprintln(stdout)
			}
		}).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if err == lessflags.ErrHelp {
			return nil
		}
		return fail(stderr, "%s", err.Error())
	}
	if len(remain) > 0 {
		return fail(stderr, "unrecognized extra args: %s", strings.Join(remain, " "))
	}

	modcache, err := resolveModCache(opts.modcache)
	if err != nil {
		return fail(stderr, "%s", err.Error())
	}
	total := 3
	if len(opts.roots) > 0 {
		total = 4
	}
	prog := newStageProgress(stderr, total)
	inv, err := inventoryCache(modcache, prog)
	if err != nil {
		return fail(stderr, "%s", err.Error())
	}
	live, err := collectLiveSet(opts.roots, modcache, opts.noCache, opts.cacheDir, stderr, prog)
	if err != nil {
		return fail(stderr, "%s", err.Error())
	}
	rep := buildReport(inv, live, opts.includeToolchain)
	warnLiveLegacy(stderr, rep, live)
	return runPrune(stdout, stderr, rep, opts.dryRun, opts.json)
}

func fail(stderr io.Writer, format string, args ...interface{}) error {
	fmt.Fprintf(stderr, "Error: "+format+"\n", args...)
	return errs.NewSilenceExitCode(1)
}

func resolveModCache(explicit string) (string, error) {
	raw := strings.TrimSpace(explicit)
	if raw == "" {
		if v := os.Getenv("GOMODCACHE"); v != "" {
			raw = v
		} else {
			out, err := exec.Command("go", "env", "GOMODCACHE").Output()
			if err != nil {
				return "", fmt.Errorf("go env GOMODCACHE: %w", err)
			}
			raw = strings.TrimSpace(string(out))
		}
	}
	if raw == "" {
		return "", fmt.Errorf("GOMODCACHE is empty")
	}
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", err
	}
	st, err := os.Stat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("modcache not a directory: %s", abs)
		}
		return "", err
	}
	if !st.IsDir() {
		return "", fmt.Errorf("modcache not a directory: %s", abs)
	}
	return abs, nil
}
