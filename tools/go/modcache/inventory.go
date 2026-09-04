package modcache

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

const toolchainPath = "golang.org/toolchain"

var downloadSuffixes = []string{".ziphash", ".partial", ".info", ".mod", ".lock", ".zip"}

type cachedVersion struct {
	Path           string
	Version        string
	ExtractedDir   string
	ExtractedBytes int64
	DownloadFiles  []string
	DownloadBytes  int64
}

func (v *cachedVersion) totalBytes() int64 {
	if v == nil {
		return 0
	}
	return v.ExtractedBytes + v.DownloadBytes
}

func (v *cachedVersion) removePaths() []string {
	if v == nil {
		return nil
	}
	out := make([]string, 0, 1+len(v.DownloadFiles))
	if v.ExtractedDir != "" {
		out = append(out, v.ExtractedDir)
	}
	out = append(out, v.DownloadFiles...)
	return out
}

type cachedModule struct {
	Path     string
	Versions map[string]*cachedVersion
}

type inventory struct {
	ModCache       string
	Modules        map[string]*cachedModule
	ExtractedBytes int64
	DownloadBytes  int64
	VCSBytes       int64
}

type extractedHit struct {
	path    string
	version string
	dir     string
}

func inventoryCache(modcache string, prog *stageProgress) (*inventory, error) {
	inv := &inventory{
		ModCache: modcache,
		Modules:  map[string]*cachedModule{},
	}
	if err := inv.walkExtracted(prog); err != nil {
		return nil, err
	}
	if err := inv.walkDownload(prog); err != nil {
		return nil, err
	}
	inv.walkVCS(prog)
	return inv, nil
}

func (inv *inventory) walkExtracted(prog *stageProgress) error {
	prog.start("extracted", "walking")
	hits, err := inv.discoverExtracted()
	if err != nil {
		return err
	}
	n := len(hits)
	prog.line("extracted", fmt.Sprintf("sizing %d versions", n))
	for i, h := range hits {
		cv := inv.ensure(h.path, h.version)
		cv.ExtractedDir = h.dir
		cv.ExtractedBytes = dirSize(h.dir)
		inv.ExtractedBytes += cv.ExtractedBytes
		if shouldHeartbeat(i+1, n) {
			prog.detail("%d/%d  %s@%s", i+1, n, h.path, h.version)
		}
	}
	prog.ok("extracted", formatBytes(inv.ExtractedBytes))
	return nil
}

func (inv *inventory) discoverExtracted() ([]extractedHit, error) {
	cacheDir := filepath.Join(inv.ModCache, "cache")
	var hits []extractedHit
	err := filepath.WalkDir(inv.ModCache, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if path == inv.ModCache {
				return err
			}
			return nil
		}
		if path == cacheDir {
			return filepath.SkipDir
		}
		if path == inv.ModCache || !d.IsDir() {
			return nil
		}
		name := d.Name()
		at := strings.LastIndex(name, "@")
		if at <= 0 || at == len(name)-1 {
			return nil
		}
		rel, err := filepath.Rel(inv.ModCache, path)
		if err != nil {
			return nil
		}
		escapedPath := name[:at]
		if dir := filepath.Dir(rel); dir != "." {
			escapedPath = filepath.ToSlash(filepath.Join(dir, escapedPath))
		}
		pathName, uerr := module.UnescapePath(escapedPath)
		if uerr != nil {
			return filepath.SkipDir
		}
		ver, uerr := module.UnescapeVersion(name[at+1:])
		if uerr != nil {
			return filepath.SkipDir
		}
		hits = append(hits, extractedHit{path: pathName, version: ver, dir: path})
		return filepath.SkipDir
	})
	return hits, err
}

func (inv *inventory) walkDownload(prog *stageProgress) error {
	prog.start("download", "walking")
	root := filepath.Join(inv.ModCache, "cache", "download")
	st, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			prog.ok("download", formatBytes(0))
			return nil
		}
		return err
	}
	if !st.IsDir() {
		prog.ok("download", formatBytes(0))
		return nil
	}
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, lerr := os.Lstat(path)
		if lerr != nil {
			return nil
		}
		inv.DownloadBytes += info.Size()
		if filepath.Base(filepath.Dir(path)) != "@v" {
			return nil
		}
		ver, _, ok := parseDownloadFile(d.Name())
		if !ok {
			return nil
		}
		rel, rerr := filepath.Rel(root, filepath.Dir(filepath.Dir(path)))
		if rerr != nil {
			return nil
		}
		pathName, uerr := module.UnescapePath(filepath.ToSlash(rel))
		if uerr != nil {
			return nil
		}
		ver, uerr = module.UnescapeVersion(ver)
		if uerr != nil {
			return nil
		}
		cv := inv.ensure(pathName, ver)
		cv.DownloadFiles = append(cv.DownloadFiles, path)
		cv.DownloadBytes += info.Size()
		return nil
	})
	if err != nil {
		return err
	}
	prog.ok("download", formatBytes(inv.DownloadBytes))
	return nil
}

func (inv *inventory) walkVCS(prog *stageProgress) {
	prog.start("vcs", "walking")
	vcs := filepath.Join(inv.ModCache, "cache", "vcs")
	if st, err := os.Stat(vcs); err == nil && st.IsDir() {
		inv.VCSBytes = dirSize(vcs)
	}
	prog.ok("vcs", formatBytes(inv.VCSBytes))
}

func (inv *inventory) ensure(path, version string) *cachedVersion {
	mod := inv.Modules[path]
	if mod == nil {
		mod = &cachedModule{Path: path, Versions: map[string]*cachedVersion{}}
		inv.Modules[path] = mod
	}
	cv := mod.Versions[version]
	if cv == nil {
		cv = &cachedVersion{Path: path, Version: version}
		mod.Versions[version] = cv
	}
	return cv
}

func parseDownloadFile(name string) (version, kind string, ok bool) {
	for _, suf := range downloadSuffixes {
		if strings.HasSuffix(name, suf) {
			return strings.TrimSuffix(name, suf), strings.TrimPrefix(suf, "."), true
		}
	}
	return "", "", false
}

func newestVersion(versions []string) string {
	if len(versions) == 0 {
		return ""
	}
	best := versions[0]
	for _, v := range versions[1:] {
		if semver.Compare(v, best) > 0 {
			best = v
		}
	}
	return best
}

func isToolchain(path string) bool {
	return path == toolchainPath
}

func sortedVersions(m map[string]*cachedVersion) []string {
	out := make([]string, 0, len(m))
	for v := range m {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool {
		c := semver.Compare(out[i], out[j])
		if c != 0 {
			return c < 0
		}
		return out[i] < out[j]
	})
	return out
}

func sortedPaths(mods map[string]*cachedModule) []string {
	out := make([]string, 0, len(mods))
	for p := range mods {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}
