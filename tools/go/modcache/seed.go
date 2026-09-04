package modcache

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/seed"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/update"
	lessflags "github.com/xhd2015/less-flags"
	"golang.org/x/mod/modfile"
)

const seedHelp = `Usage: kool go modcache seed [--tag TAG] [--dry-run] [--json] [-v] <dir>

Seed GOMODCACHE for the Go module at <dir> from this repo's git tags
(GOPROXY=direct + url.file:// insteadOf → repo root).

  <dir>              module directory; <dir>/go.mod required
  --tag TAG          latest (default) | HEAD | vX.Y.Z
                     nested modules use git tag <rel>/vX.Y.Z
  --dry-run          resolve for real; gate download with would:
  --json             JSON on stdout (no ANSI)
  -v, --verbose      kind-aligned notice: under download
  -h, --help         show this help
`

var explicitVersionRE = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

type seedOpts struct {
	tag     string
	dryRun  bool
	jsonOut bool
	verbose bool
}

type seedJSON struct {
	Path    string `json:"path"`
	Version string `json:"version"`
	Tag     string `json:"tag"`
	Zip     string `json:"zip,omitempty"`
	Sum     string `json:"sum,omitempty"`
	DryRun  bool   `json:"dryRun,omitempty"`
}

func handleSeed(args []string, stdout, stderr io.Writer) error {
	opts := seedOpts{tag: "latest"}
	remain, err := lessflags.String("--tag", &opts.tag).
		Bool("--dry-run", &opts.dryRun).
		Bool("--json", &opts.jsonOut).
		Bool("-v,--verbose", &opts.verbose).
		HelpFunc("-h,--help", func() {
			fmt.Fprint(stdout, seedHelp)
			if !strings.HasSuffix(seedHelp, "\n") {
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
	if len(remain) != 1 {
		return fail(stderr, "kool go modcache seed requires exactly one <dir>")
	}
	modDir, err := filepath.Abs(remain[0])
	if err != nil {
		return fail(stderr, "resolve dir: %v", err)
	}

	goModPath := filepath.Join(modDir, "go.mod")
	if _, err := os.Stat(goModPath); err != nil {
		if os.IsNotExist(err) {
			return fail(stderr, "%s: no such file", goModPath)
		}
		return fail(stderr, "%s: %v", goModPath, err)
	}

	repoDir, err := gitToplevel(modDir)
	if err != nil {
		return fail(stderr, "%s is not inside a git work tree: %v", modDir, err)
	}

	modPath, err := readModulePath(goModPath)
	if err != nil {
		return fail(stderr, "%v", err)
	}
	if !seed.IsWellKnown(modPath) {
		return fail(stderr, "module %s is not a well-known VCS host (github.com/…)", modPath)
	}

	prefix, err := update.CalculateVersionPrefix(modDir, modPath)
	if err != nil {
		return fail(stderr, "version prefix: %v", err)
	}

	prog := newStageProgress(stderr, 2)

	fullTag, goVer, tagMode, err := resolveSeedTag(modDir, repoDir, prefix, opts.tag)
	if err != nil {
		prog.start("resolve", "tag "+opts.tag)
		return fail(stderr, "%v", err)
	}
	prog.start("resolve", fmt.Sprintf("tag %s (%s)", fullTag, tagMode))

	spec := modPath + "@" + goVer
	prog.start("download", spec)

	if opts.dryRun {
		prog.detail("would: go mod download -json %s", spec)
		fmt.Fprintln(stderr)
		if opts.jsonOut {
			enc := json.NewEncoder(stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(seedJSON{Path: modPath, Version: goVer, Tag: fullTag, DryRun: true})
		}
		fmt.Fprintf(stdout, "would: seed %s\n", spec)
		return nil
	}

	if opts.verbose {
		prog.detail("notice: go mod download -json %s", spec)
	}

	res, err := seed.Download(context.Background(), seed.Request{
		RepoDir: repoDir,
		Modules: []seed.Module{{Path: modPath, Version: goVer}},
	})
	if err != nil {
		return fail(stderr, "%v", err)
	}
	if len(res.Modules) != 1 {
		return fail(stderr, "internal: expected 1 module result")
	}
	mr := res.Modules[0]
	if mr.Skipped {
		return fail(stderr, "module %s skipped (not well-known)", modPath)
	}
	if mr.Err != nil {
		return fail(stderr, "seed %s: %v", spec, mr.Err)
	}
	if opts.verbose {
		prog.detail("ok")
	}
	fmt.Fprintln(stderr)
	if opts.jsonOut {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(seedJSON{
			Path:    modPath,
			Version: goVer,
			Tag:     fullTag,
			Zip:     mr.Zip,
			Sum:     mr.Sum,
		})
	}
	fmt.Fprintf(stdout, "seeded %s\n", spec)
	return nil
}

func resolveSeedTag(modDir, repoDir, prefix, tagFlag string) (fullTag, goVer, mode string, err error) {
	tagFlag = strings.TrimSpace(tagFlag)
	if tagFlag == "" {
		tagFlag = "latest"
	}
	switch {
	case tagFlag == "latest":
		full, err := update.GetLatestVersionTag(modDir, prefix)
		if err != nil {
			return "", "", "", fmt.Errorf("no latest release tag for prefix %q: %w", prefix, err)
		}
		return full, update.StripVersionPrefix(prefix, full), "latest", nil
	case tagFlag == "HEAD":
		full, err := tagAtHEAD(repoDir, prefix)
		if err != nil {
			return "", "", "", err
		}
		return full, update.StripVersionPrefix(prefix, full), "HEAD", nil
	case explicitVersionRE.MatchString(tagFlag):
		full := prefix + strings.TrimPrefix(tagFlag, "v")
		if !gitTagExists(repoDir, full) {
			return "", "", "", fmt.Errorf("tag %s does not exist", full)
		}
		return full, tagFlag, "explicit", nil
	default:
		return "", "", "", fmt.Errorf("--tag must be latest, HEAD, or vX.Y.Z (got %q)", tagFlag)
	}
}

func tagAtHEAD(repoDir, prefix string) (string, error) {
	out, err := exec.Command("git", "-C", repoDir, "tag", "--points-at", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("git tag --points-at HEAD: %w", err)
	}
	var match string
	for _, line := range strings.Split(string(out), "\n") {
		tag := strings.TrimSpace(line)
		if tag == "" || !strings.HasPrefix(tag, prefix) {
			continue
		}
		basic := strings.TrimPrefix(tag, prefix)
		if strings.Contains(basic, "/") {
			continue
		}
		if !explicitVersionRE.MatchString("v" + basic) {
			continue
		}
		match = tag
		break
	}
	if match == "" {
		want := prefix + "X.Y.Z"
		if prefix == "v" {
			want = "vX.Y.Z"
		} else {
			want = prefix + "X.Y.Z"
		}
		return "", fmt.Errorf("HEAD is not tagged with a release tag for this module (want %s)", want)
	}
	return match, nil
}

func gitTagExists(repoDir, tag string) bool {
	err := exec.Command("git", "-C", repoDir, "rev-parse", "-q", "--verify", "refs/tags/"+tag).Run()
	return err == nil
}

func gitToplevel(dir string) (string, error) {
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func readModulePath(goModPath string) (string, error) {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}
	f, err := modfile.Parse(goModPath, data, nil)
	if err != nil {
		return "", fmt.Errorf("parse %s: %w", goModPath, err)
	}
	if f.Module == nil || f.Module.Mod.Path == "" {
		return "", fmt.Errorf("%s: missing module path", goModPath)
	}
	return f.Module.Mod.Path, nil
}
