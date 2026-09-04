package modcache

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
	scanpkg "github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan"
)

type consumer struct {
	Repo      string
	ModuleDir string
	Path      string
	Version   string
}

type liveSet struct {
	Repos     int
	GoModules int
	// key is path@version
	Refs map[string][]consumer
}

func collectLiveSet(roots []string, modcache string, noCache bool, cacheDir string, stderr io.Writer, prog *stageProgress) (*liveSet, error) {
	live := &liveSet{Refs: map[string][]consumer{}}
	if len(roots) == 0 {
		return live, nil
	}

	prog.start("live", "scanning")
	result, err := scan_repo.Scan(context.Background(), scan_repo.Options{
		Roots:     roots,
		NoCache:   noCache,
		CacheRoot: cacheDir,
		Stderr:    stderr,
	})
	if err != nil {
		return nil, err
	}
	for _, re := range result.RootErrors {
		fmt.Fprintf(stderr, "warning: %s: %s\n", re.Root, re.Error)
	}

	modcacheAbs, err := filepath.Abs(modcache)
	if err != nil {
		modcacheAbs = modcache
	}

	for _, repo := range result.Repos {
		if repo.Error != "" {
			fmt.Fprintf(stderr, "warning: skipped unreadable repo %s: %s\n", repo.Path, repo.Error)
			continue
		}
		if isUnder(repo.Path, modcacheAbs) {
			continue
		}
		live.Repos++
		err := scanpkg.ScanStream(repo.Path, scanpkg.Options{}, func(m scanpkg.Module) error {
			modDir := repo.Path
			if m.Dir != "." && m.Dir != "" {
				modDir = filepath.Join(repo.Path, filepath.FromSlash(m.Dir))
			}
			live.GoModules++
			addLiveRequires(live, repo.Path, modDir, m)
			if err := addLiveGoSum(live, repo.Path, modDir); err != nil {
				fmt.Fprintf(stderr, "warning: %s: %v\n", filepath.Join(modDir, "go.sum"), err)
			}
			return nil
		})
		if err != nil {
			fmt.Fprintf(stderr, "warning: scanning go modules in %s: %v\n", repo.Path, err)
		}
	}
	prog.ok("live", fmt.Sprintf("%d git repos, %d go modules", live.Repos, live.GoModules))
	return live, nil
}

func addLiveRequires(live *liveSet, repo, modDir string, m scanpkg.Module) {
	for _, req := range m.Requires {
		if req.Path == "" || req.Version == "" {
			continue
		}
		live.add(repo, modDir, req.Path, req.Version)
	}
	for _, repl := range m.Replaces {
		if repl.NewPath == "" || repl.NewVersion == "" {
			continue
		}
		live.add(repo, modDir, repl.NewPath, repl.NewVersion)
	}
}

func addLiveGoSum(live *liveSet, repo, modDir string) error {
	data, err := os.ReadFile(filepath.Join(modDir, "go.sum"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		ver := strings.TrimSuffix(fields[1], "/go.mod")
		if fields[0] == "" || ver == "" {
			continue
		}
		live.add(repo, modDir, fields[0], ver)
	}
	return nil
}

func (l *liveSet) add(repo, modDir, path, version string) {
	if l == nil || l.Refs == nil {
		return
	}
	key := path + "@" + version
	for _, c := range l.Refs[key] {
		if c.ModuleDir == modDir && c.Path == path && c.Version == version {
			return
		}
	}
	l.Refs[key] = append(l.Refs[key], consumer{
		Repo:      repo,
		ModuleDir: modDir,
		Path:      path,
		Version:   version,
	})
}

func (l *liveSet) consumers(path, version string) []consumer {
	if l == nil {
		return nil
	}
	return l.Refs[path+"@"+version]
}

func isUnder(path, root string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		absRoot = root
	}
	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	sep := string(filepath.Separator)
	return rel != ".." && !strings.HasPrefix(rel, ".."+sep)
}
