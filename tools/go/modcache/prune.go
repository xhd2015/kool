package modcache

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/xhd2015/dot-pkgs/go-pkgs/terminal/color"
)

type pruneJSON struct {
	DryRun   bool     `json:"dryRun"`
	Versions int      `json:"versions"`
	Bytes    int64    `json:"bytes"`
	Paths    []string `json:"paths"`
}

func runPrune(stdout, stderr io.Writer, rep *Report, dryRun, jsonOut bool) error {
	var paths []string
	seen := map[string]bool{}
	for _, cv := range rep.legacy {
		for _, p := range cv.removePaths() {
			if p == "" || seen[p] {
				continue
			}
			seen[p] = true
			paths = append(paths, p)
		}
	}
	sort.Strings(paths)

	if jsonOut {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(pruneJSON{
			DryRun:   dryRun,
			Versions: rep.LegacyVersions,
			Bytes:    rep.LegacyBytes,
			Paths:    paths,
		})
	}

	verb := "removed"
	if dryRun {
		verb = "would remove"
	}
	summary := fmt.Sprintf("%s %d versions, %s", verb, rep.LegacyVersions, formatBytes(rep.LegacyBytes))
	style := color.Style{Enabled: color.EnabledFor(color.Auto, stdout)}
	if !dryRun && rep.LegacyVersions > 0 {
		summary = style.Green(summary)
	}
	fmt.Fprintln(stdout, summary)
	if dryRun {
		for _, p := range paths {
			fmt.Fprintf(stdout, "  rm  %s\n", p)
		}
		return nil
	}

	for _, p := range paths {
		if err := removePath(p); err != nil {
			return fail(stderr, "remove %s: %s", p, err.Error())
		}
	}
	return nil
}
