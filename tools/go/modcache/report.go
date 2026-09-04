package modcache

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/terminal/color"
	"golang.org/x/mod/semver"
)

type VersionJSON struct {
	Version        string `json:"version"`
	Legacy         bool   `json:"legacy"`
	ExtractedBytes int64  `json:"extractedBytes"`
	DownloadBytes  int64  `json:"downloadBytes"`
	ExtractedDir   string `json:"extractedDir,omitempty"`
}

type ModuleJSON struct {
	Path      string        `json:"path"`
	Newest    string        `json:"newest"`
	Toolchain bool          `json:"toolchain,omitempty"`
	Versions  []VersionJSON `json:"versions"`
}

type SuggestionJSON struct {
	Path    string   `json:"path"`
	Current string   `json:"current"`
	Newest  string   `json:"newest"`
	Repos   []string `json:"repos"`
}

type Report struct {
	ModCache         string           `json:"modCache"`
	TotalBytes       int64            `json:"totalBytes"`
	ExtractedBytes   int64            `json:"extractedBytes"`
	DownloadBytes    int64            `json:"downloadBytes"`
	VCSBytes         int64            `json:"vcsBytes"`
	ToolchainBytes   int64            `json:"toolchainBytes"`
	ToolchainVers    int              `json:"toolchainVersions"`
	ModulePaths      int              `json:"modulePaths"`
	Versions         int              `json:"versions"`
	MultiVersion     int              `json:"multiVersion"`
	LegacyVersions   int              `json:"legacyVersions"`
	LegacyBytes      int64            `json:"legacyBytes"`
	SaveBytes        int64            `json:"saveBytes"`
	ReposScanned     int              `json:"reposScanned,omitempty"`
	GoModules        int              `json:"goModules,omitempty"`
	Modules          []ModuleJSON     `json:"modules"`
	Suggestions      []SuggestionJSON `json:"suggestions"`
	legacy           []*cachedVersion
	includeToolchain bool
}

type legacyRow struct {
	Path     string
	Newest   string
	Versions int
	Bytes    int64
}

func buildReport(inv *inventory, live *liveSet, includeToolchain bool) *Report {
	rep := &Report{
		ModCache:         inv.ModCache,
		ExtractedBytes:   inv.ExtractedBytes,
		DownloadBytes:    inv.DownloadBytes,
		VCSBytes:         inv.VCSBytes,
		Modules:          []ModuleJSON{},
		Suggestions:      []SuggestionJSON{},
		includeToolchain: includeToolchain,
	}
	rep.TotalBytes = inv.ExtractedBytes + inv.DownloadBytes + inv.VCSBytes
	if live != nil {
		rep.ReposScanned = live.Repos
		rep.GoModules = live.GoModules
	}

	for _, path := range sortedPaths(inv.Modules) {
		mod := inv.Modules[path]
		vers := sortedVersions(mod.Versions)
		newest := newestVersion(vers)
		toolchain := isToolchain(path)
		mj := ModuleJSON{
			Path:      path,
			Newest:    newest,
			Toolchain: toolchain,
			Versions:  make([]VersionJSON, 0, len(vers)),
		}
		for _, ver := range vers {
			cv := mod.Versions[ver]
			legacy := ver != newest
			if toolchain && !includeToolchain {
				legacy = false
			}
			mj.Versions = append(mj.Versions, VersionJSON{
				Version:        ver,
				Legacy:         legacy,
				ExtractedBytes: cv.ExtractedBytes,
				DownloadBytes:  cv.DownloadBytes,
				ExtractedDir:   cv.ExtractedDir,
			})
			if toolchain {
				rep.ToolchainBytes += cv.totalBytes()
				rep.ToolchainVers++
			}
			if !toolchain || includeToolchain {
				rep.Versions++
				if legacy {
					rep.LegacyVersions++
					rep.LegacyBytes += cv.totalBytes()
					rep.legacy = append(rep.legacy, cv)
				}
			}
		}
		if toolchain && !includeToolchain {
			rep.Modules = append(rep.Modules, mj)
			continue
		}
		rep.ModulePaths++
		if len(vers) >= 2 {
			rep.MultiVersion++
		}
		rep.Modules = append(rep.Modules, mj)
	}

	rep.Suggestions = buildSuggestions(inv, live)
	rep.SaveBytes = rep.LegacyBytes
	return rep
}

func buildSuggestions(inv *inventory, live *liveSet) []SuggestionJSON {
	if live == nil || len(live.Refs) == 0 {
		return []SuggestionJSON{}
	}
	type key struct{ path, current, newest string }
	grouped := map[key]map[string]bool{}
	for _, path := range sortedPaths(inv.Modules) {
		if isToolchain(path) {
			continue
		}
		mod := inv.Modules[path]
		vers := sortedVersions(mod.Versions)
		newest := newestVersion(vers)
		if newest == "" {
			continue
		}
		for _, ver := range vers {
			if ver == newest || semver.Compare(ver, newest) >= 0 {
				continue
			}
			for _, c := range live.consumers(path, ver) {
				k := key{path: path, current: ver, newest: newest}
				if grouped[k] == nil {
					grouped[k] = map[string]bool{}
				}
				grouped[k][c.ModuleDir] = true
			}
		}
	}
	out := make([]SuggestionJSON, 0, len(grouped))
	for k, repos := range grouped {
		s := SuggestionJSON{Path: k.path, Current: k.current, Newest: k.newest, Repos: make([]string, 0, len(repos))}
		for r := range repos {
			s.Repos = append(s.Repos, r)
		}
		sort.Strings(s.Repos)
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Path != out[j].Path {
			return out[i].Path < out[j].Path
		}
		return out[i].Current < out[j].Current
	})
	return out
}

func renderInspect(stdout, stderr io.Writer, rep *Report, opts options) error {
	if opts.json {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(rep); err != nil {
			return err
		}
		return nil
	}

	style := color.Style{Enabled: color.EnabledFor(color.Auto, stdout)}
	label := func(s string) string { return style.Gray(s) }

	fmt.Fprintf(stdout, "%s  %s\n", label("GOMODCACHE:"), rep.ModCache)
	fmt.Fprintf(stdout, "%s  %s  (extracted %s, download %s, vcs %s)\n",
		label("TOTAL:"), formatBytes(rep.TotalBytes),
		formatBytes(rep.ExtractedBytes), formatBytes(rep.DownloadBytes), formatBytes(rep.VCSBytes))
	fmt.Fprintf(stdout, "%s  %d paths, %d versions (%d with 2+ versions)\n",
		label("MODULES:"), rep.ModulePaths, rep.Versions, rep.MultiVersion)
	fmt.Fprintf(stdout, "%s  %d versions, %s\n",
		label("LEGACY:"), rep.LegacyVersions, formatBytes(rep.LegacyBytes))
	saveAmt := formatBytes(rep.SaveBytes)
	if style.Enabled && rep.SaveBytes > 0 {
		saveAmt = style.Green(saveAmt)
	}
	fmt.Fprintf(stdout, "%s  %s  if prune keeps newest of each module%s\n",
		label("SAVE:"), saveAmt, savePercent(rep.SaveBytes, rep.TotalBytes))
	if rep.ToolchainVers > 0 && !opts.includeToolchain {
		fmt.Fprintf(stdout, "%s  %d versions, %s  (excluded from prune)\n",
			label("TOOLCHAIN:"), rep.ToolchainVers, formatBytes(rep.ToolchainBytes))
	}
	if len(opts.roots) > 0 {
		fmt.Fprintf(stdout, "%s  %d git repos, %d go modules\n",
			label("REPOS:"), rep.ReposScanned, rep.GoModules)
	}

	rows := aggregateLegacy(rep)
	if len(rows) > 0 {
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "TOP legacy by size")
		fmt.Fprintf(stdout, "  %-8s %-16s %s\n", "SIZE", "KEEP", "PATH")
		limit := opts.top
		if limit <= 0 || limit > len(rows) {
			limit = len(rows)
		}
		for _, row := range rows[:limit] {
			path := row.Path
			if row.Versions > 1 {
				path = fmt.Sprintf("%s (%d versions)", row.Path, row.Versions)
			}
			fmt.Fprintf(stdout, "  %-8s %-16s %s\n", formatBytes(row.Bytes), row.Newest, path)
		}
	}

	if len(rep.Suggestions) > 0 {
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "upgrade suggestions")
		for _, s := range rep.Suggestions {
			repos := strings.Join(s.Repos, ", ")
			fmt.Fprintf(stdout, "  %s  %s  %s -> %s\n", repos, s.Path, s.Current, s.Newest)
		}
	}

	if len(opts.roots) == 0 && rep.LegacyVersions > 0 {
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, style.Gray("notice: pass --root ~ to list local go.mod/go.sum still on legacy versions"))
	}
	return nil
}

func savePercent(save, total int64) string {
	if save <= 0 || total <= 0 {
		return ""
	}
	pct := 100 * float64(save) / float64(total)
	if pct < 1 {
		return " (<1% of total)"
	}
	return fmt.Sprintf(" (%.0f%% of total)", pct)
}

func aggregateLegacy(rep *Report) []legacyRow {
	type agg struct {
		newest   string
		versions int
		bytes    int64
	}
	byPath := map[string]*agg{}
	for _, mj := range rep.Modules {
		if mj.Toolchain && !rep.includeToolchain {
			continue
		}
		a := byPath[mj.Path]
		if a == nil {
			a = &agg{newest: mj.Newest}
			byPath[mj.Path] = a
		}
		for _, v := range mj.Versions {
			if !v.Legacy {
				continue
			}
			a.versions++
			a.bytes += v.ExtractedBytes + v.DownloadBytes
		}
		if a.versions == 0 {
			delete(byPath, mj.Path)
		}
	}
	rows := make([]legacyRow, 0, len(byPath))
	for path, a := range byPath {
		rows = append(rows, legacyRow{Path: path, Newest: a.newest, Versions: a.versions + 1, Bytes: a.bytes})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Bytes != rows[j].Bytes {
			return rows[i].Bytes > rows[j].Bytes
		}
		return rows[i].Path < rows[j].Path
	})
	return rows
}

func warnLiveLegacy(stderr io.Writer, rep *Report, live *liveSet) {
	if live == nil || len(live.Refs) == 0 {
		return
	}
	type line struct {
		dir, ref string
	}
	var lines []line
	for _, cv := range rep.legacy {
		for _, c := range live.consumers(cv.Path, cv.Version) {
			lines = append(lines, line{dir: c.ModuleDir, ref: cv.Path + "@" + cv.Version})
		}
	}
	if len(lines) == 0 {
		return
	}
	sort.Slice(lines, func(i, j int) bool {
		if lines[i].dir != lines[j].dir {
			return lines[i].dir < lines[j].dir
		}
		return lines[i].ref < lines[j].ref
	})
	dirs := map[string]bool{}
	for _, l := range lines {
		dirs[l.dir] = true
	}
	fmt.Fprintf(stderr, "warning: %d local modules still require versions that will be removed\n", len(dirs))
	for _, l := range lines {
		fmt.Fprintf(stderr, "  %s  %s\n", l.dir, l.ref)
	}
}
