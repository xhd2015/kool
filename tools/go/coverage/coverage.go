package coverage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/xhd2015/kool/pkgs/errs"
	"github.com/xhd2015/kool/pkgs/flag"
	"golang.org/x/tools/cover"
)

const rootHelp = `kool go coverage - package coverage reports from coverprofiles

Usage:
  kool go coverage package-table [OPTIONS] <coverage.out>
  kool go coverage -h|--help

Subcommands:
  package-table                    print markdown/JSON package coverage table

Examples:
  kool go coverage package-table coverage.out
  kool go coverage package-table --module example.com/mod coverage.out
  kool go coverage package-table --json coverage.out
`

const packageTableHelp = `kool go coverage package-table - print package coverage table from a coverprofile

Usage:
  kool go coverage package-table [OPTIONS] <coverage.out>

Arguments:
  coverage.out                     classic go coverprofile (mode: + file blocks)

Options:
  --module PATH                    module path prefix (default: from go.mod under --dir)
  --dir DIR                        directory for go.mod lookup (default: .)
  --skip-prefix LIST               comma-separated path prefixes relative to module
                                   (default: script/,cmd/); replaces defaults when set
  --skip-contains LIST             comma-separated substrings (default: /legacy_);
                                   replaces defaults when set
  --all                            do not filter by module; package path is dir of full file path
  --json                           JSON rows instead of markdown
  -h,--help                        show help message

Examples:
  kool go coverage package-table coverage.out
  kool go coverage package-table --dir . --module example.com/mod coverage.out
  kool go coverage package-table --skip-prefix tmp/ --skip-contains /x_drop coverage.out
  kool go coverage package-table --all --json coverage.out
`

// Handle is production entry: HandleWith(args, HandleOpts{}).
func Handle(args []string) error {
	return HandleWith(args, HandleOpts{})
}

// HandleOpts injects writers for tests and production defaults.
type HandleOpts struct {
	// Stdout/Stderr nil → os.Stdout / os.Stderr.
	Stdout io.Writer
	Stderr io.Writer
}

// HandleWith is the injectable entry used by doctests.
// args are the argv after "kool go coverage".
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
		fmt.Fprintln(stderr, "requires subcommand: package-table (try --help)")
		return errs.NewSilenceExitCode(1)
	}

	if args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		fmt.Fprint(stdout, rootHelp)
		if !strings.HasSuffix(rootHelp, "\n") {
			fmt.Fprintln(stdout)
		}
		return nil
	}

	cmd := args[0]
	rest := args[1:]
	switch cmd {
	case "package-table":
		return handlePackageTable(rest, stdout, stderr)
	default:
		return fmt.Errorf("unknown subcommand: %s (try package-table or --help)", cmd)
	}
}

func handlePackageTable(args []string, stdout, stderr io.Writer) error {
	var (
		module          string
		moduleSet       bool
		dir             string
		dirSet          bool
		skipPrefix      string
		skipPrefixSet   bool
		skipContains    string
		skipContainsSet bool
		all             bool
		jsonOut         bool
	)

	var remain []string
	for i := 0; i < len(args); i++ {
		f, value := flag.ParseFlag(args, &i)
		if f == "" {
			remain = append(remain, args[i])
			continue
		}
		switch f {
		case "-h", "--help":
			fmt.Fprint(stdout, packageTableHelp)
			if !strings.HasSuffix(packageTableHelp, "\n") {
				fmt.Fprintln(stdout)
			}
			return nil
		case "--module":
			v, ok := value()
			if !ok {
				return usageExit(stderr, "--module requires a value")
			}
			module = v
			moduleSet = true
		case "--dir":
			v, ok := value()
			if !ok {
				return usageExit(stderr, "--dir requires a value")
			}
			dir = v
			dirSet = true
		case "--skip-prefix":
			v, ok := value()
			if !ok {
				return usageExit(stderr, "--skip-prefix requires a value")
			}
			skipPrefix = v
			skipPrefixSet = true
		case "--skip-contains":
			v, ok := value()
			if !ok {
				return usageExit(stderr, "--skip-contains requires a value")
			}
			skipContains = v
			skipContainsSet = true
		case "--all":
			all = true
		case "--json":
			jsonOut = true
		default:
			return usageExit(stderr, "unrecognized flag: "+f)
		}
	}

	if len(remain) == 0 {
		return usageExit(stderr, "requires coverage profile path: kool go coverage package-table [OPTIONS] <coverage.out>")
	}
	if len(remain) > 1 {
		return usageExit(stderr, "unrecognized extra args: "+strings.Join(remain[1:], " "))
	}
	profilePath := remain[0]

	st, err := os.Stat(profilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Error: profile file not found: %s", profilePath)
		}
		return fmt.Errorf("Error: cannot read profile: %v", err)
	}
	if st.IsDir() {
		return fmt.Errorf("Error: profile path is a directory: %s", profilePath)
	}

	if !dirSet || dir == "" {
		dir = "."
	}

	modulePath := module
	if !all && !moduleSet {
		mp, err := readModulePath(dir)
		if err != nil {
			return err
		}
		modulePath = mp
	} else if moduleSet {
		modulePath = module
	}

	skipPrefixes := defaultSkipPrefixes
	if skipPrefixSet {
		skipPrefixes = splitCSV(skipPrefix)
	}
	skipContainsList := defaultSkipContains
	if skipContainsSet {
		skipContainsList = splitCSV(skipContains)
	}

	rows, err := buildPackageRows(profilePath, modulePath, all, skipPrefixes, skipContainsList)
	if err != nil {
		return err
	}

	if len(rows) == 0 {
		fmt.Fprintln(stderr, "warning: no packages matched after filters")
	}

	if jsonOut {
		return writeJSON(stdout, rows)
	}
	return writeMarkdown(stdout, rows)
}

func usageExit(stderr io.Writer, msg string) error {
	fmt.Fprintln(stderr, msg)
	fmt.Fprintln(stderr, "usage: kool go coverage package-table [OPTIONS] <coverage.out>")
	fmt.Fprintln(stderr, "try: kool go coverage package-table --help")
	return errs.NewSilenceExitCode(2)
}

var defaultSkipPrefixes = []string{"script/", "cmd/"}
var defaultSkipContains = []string{"/legacy_"}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func readModulePath(dir string) (string, error) {
	// Resolve without Chdir: relative dir is interpreted from process cwd via Abs.
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve --dir: %w", err)
	}
	data, err := os.ReadFile(filepath.Join(absDir, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("read go.mod under %s: %w", absDir, err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			mp := strings.TrimSpace(strings.TrimPrefix(line, "module"))
			mp = strings.Trim(mp, `"'`)
			if mp == "" {
				return "", fmt.Errorf("empty module path in go.mod under %s", absDir)
			}
			return mp, nil
		}
	}
	return "", fmt.Errorf("no module directive in go.mod under %s", absDir)
}

type pkgStats struct {
	covered int
	total   int
}

type packageRow struct {
	Coverage float64 `json:"coverage"`
	Package  string  `json:"package"`
}

func buildPackageRows(profilePath, modulePath string, all bool, skipPrefixes, skipContains []string) ([]packageRow, error) {
	profiles, err := cover.ParseProfiles(profilePath)
	if err != nil {
		return nil, fmt.Errorf("parse coverprofile: %w", err)
	}

	agg := make(map[string]*pkgStats)
	modulePrefix := ""
	if !all && modulePath != "" {
		modulePrefix = strings.TrimSuffix(modulePath, "/") + "/"
	}

	for _, p := range profiles {
		file := filepath.ToSlash(p.FileName)

		var pkgPath string
		if all {
			pkgPath = dirOfCoverPath(file)
			if shouldSkip(pkgPath, skipPrefixes, skipContains) || shouldSkip(file, skipPrefixes, skipContains) {
				continue
			}
		} else {
			if modulePrefix == "" || !strings.HasPrefix(file, modulePrefix) {
				continue
			}
			rel := strings.TrimPrefix(file, modulePrefix)
			if shouldSkip(rel, skipPrefixes, skipContains) {
				continue
			}
			pkgPath = dirOfCoverPath(rel)
		}

		st := agg[pkgPath]
		if st == nil {
			st = &pkgStats{}
			agg[pkgPath] = st
		}
		for _, b := range p.Blocks {
			st.total += b.NumStmt
			if b.Count > 0 {
				st.covered += b.NumStmt
			}
		}
	}

	rows := make([]packageRow, 0, len(agg))
	for name, st := range agg {
		if st.total <= 0 {
			continue
		}
		pct := 100.0 * float64(st.covered) / float64(st.total)
		rows = append(rows, packageRow{Coverage: pct, Package: name})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Coverage != rows[j].Coverage {
			return rows[i].Coverage < rows[j].Coverage
		}
		return rows[i].Package < rows[j].Package
	})
	return rows, nil
}

func dirOfCoverPath(file string) string {
	file = filepath.ToSlash(file)
	i := strings.LastIndex(file, "/")
	if i < 0 {
		return "."
	}
	return file[:i]
}

func shouldSkip(path string, prefixes, contains []string) bool {
	path = filepath.ToSlash(path)
	for _, p := range prefixes {
		if p != "" && strings.HasPrefix(path, p) {
			return true
		}
	}
	for _, c := range contains {
		if c != "" && strings.Contains(path, c) {
			return true
		}
	}
	return false
}

func writeMarkdown(w io.Writer, rows []packageRow) error {
	var b strings.Builder
	b.WriteString("| Coverage | Package |\n")
	b.WriteString("|----------|---------|\n")
	for _, r := range rows {
		pct := formatPercent(r.Coverage)
		b.WriteString("| ")
		b.WriteString(pct)
		b.WriteString(" | `")
		b.WriteString(r.Package)
		b.WriteString("` |\n")
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func formatPercent(pct float64) string {
	return strconv.FormatFloat(pct, 'f', 1, 64) + "%"
}

func writeJSON(w io.Writer, rows []packageRow) error {
	data, err := json.Marshal(rows)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}
