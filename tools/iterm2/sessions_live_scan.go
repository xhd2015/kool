package iterm2

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
)

// liveCriticalPane is the minimum iTerm metadata needed to associate a live
// terminal process with a saved checkpoint tab.
type liveCriticalPane struct {
	WindowID uint64
	ID       string
	TTY      string
	Name     string
}

// scanLiveCriticalAcrossApps scans every running iTerm installation for the
// checkpoint identities without constructing a full enriched Snapshot.
func scanLiveCriticalAcrossApps(doc *SaveDocument, strict bool) (*liveCriticalIndex, []string, error) {
	idx := newLiveCriticalIndex()
	wanted := wantedLiveCriticalKeys(doc)
	if len(wanted) == 0 {
		return idx, nil, nil
	}

	base := activeCollector()
	if base == nil {
		base = defaultCollector()
	}
	if base.fixtureEnabled {
		snap, warnings, err := base.capture(nil, CaptureOpts{NoEnrich: false})
		if err != nil {
			return nil, warnings, err
		}
		return indexLiveCriticalForKeys(snap, wanted), warnings, nil
	}
	if base.ITermRunning != nil && !base.ITermRunning() {
		return nil, nil, fmt.Errorf("Error: iTerm2 is not running")
	}

	pf := resolveMultiAppPreflight()
	sources := multiAppCaptureSources(pf)
	var warnings []string
	var panes []liveCriticalPane
	seenWindows := map[uint64]bool{}
	for si, src := range sources {
		col := *base
		col.AppTell = src.Tell
		sourcePanes, err := col.listLiveCriticalPanes()
		if err != nil {
			if si == 0 || strict {
				if si == 0 {
					return nil, warnings, err
				}
				return nil, warnings, fmt.Errorf("Error: failed to capture %s: %w", src.App, err)
			}
			warnings = append(warnings, "warning: failed to capture "+src.App+": "+err.Error())
			continue
		}
		skipWindows := map[uint64]bool{}
		sourceWindows := map[uint64]bool{}
		for _, pane := range sourcePanes {
			if pane.WindowID == 0 || sourceWindows[pane.WindowID] {
				continue
			}
			sourceWindows[pane.WindowID] = true
			if seenWindows[pane.WindowID] {
				skipWindows[pane.WindowID] = true
				continue
			}
			seenWindows[pane.WindowID] = true
		}
		for _, pane := range sourcePanes {
			if pane.WindowID != 0 && skipWindows[pane.WindowID] {
				continue
			}
			panes = append(panes, pane)
		}
	}
	if len(panes) == 0 {
		return idx, warnings, nil
	}

	ttySet := map[string]bool{}
	for _, pane := range panes {
		if tty := normalizeLiveTTY(pane.TTY); tty != "" {
			ttySet[tty] = true
		}
	}
	if len(ttySet) == 0 {
		return idx, warnings, nil
	}
	ttys := make([]string, 0, len(ttySet))
	for tty := range ttySet {
		ttys = append(ttys, tty)
	}
	sort.Strings(ttys)
	listProcs := base.ListTTYProcs
	if listProcs == nil {
		listProcs = defaultListTTYProcs
	}
	liveProcs, err := listProcs(ttys)
	if err != nil {
		return nil, warnings, fmt.Errorf("Error: failed to list live iTerm processes: %w", err)
	}

	byTTY := map[string][]rawProc{}
	allProcs := make([]procresolve.Proc, 0, len(liveProcs))
	for _, proc := range liveProcs {
		allProcs = append(allProcs, procresolve.Proc{PID: proc.PID, PPID: proc.PPID, Cmd: proc.Command})
		short := normalizeLiveTTY(proc.TTY)
		if short != "" {
			byTTY[short] = append(byTTY[short], proc.rawProc)
		}
	}
	resolveAgent := liveCriticalAgentResolver(base, allProcs)
	wantedITerm := wantedLiveITermIDs(doc)
	wantedAgents := wantedLiveAgentCounts(doc)
	foundAgents := map[string]int{}
	var fallback []liveCriticalCandidate

	for _, pane := range panes {
		procs := byTTY[normalizeLiveTTY(pane.TTY)]
		idle, _, chosen, _, _ := enrichFromProcs(procs, nil, baseNow(base))
		if idle == nil || *idle || chosen == nil {
			continue
		}
		candidate := liveCriticalCandidate{
			pane:     pane,
			procs:    procs,
			chosen:   *chosen,
			mayAgent: paneMayContainAgent(chosen.PID, allProcs),
		}
		candidate.mark, candidate.hasMark = matchingLiveMark(procs, wanted)
		if !candidate.mayAgent {
			idx.addMarkCandidate(candidate)
			continue
		}
		if wantedITerm[normalizeITermSessionID(pane.ID)] || candidate.hasMark {
			if idx.resolveAgentCandidate(candidate, resolveAgent, wanted, foundAgents, wantedAgents) {
				continue
			}
			idx.addMarkCandidate(candidate)
			continue
		}
		fallback = append(fallback, candidate)
	}

	if liveAgentMatchesMissing(foundAgents, wantedAgents) {
		for _, candidate := range fallback {
			idx.resolveAgentCandidate(candidate, resolveAgent, wanted, foundAgents, wantedAgents)
			if !liveAgentMatchesMissing(foundAgents, wantedAgents) {
				break
			}
		}
	}
	return idx, warnings, nil
}

func wantedLiveCriticalKeys(doc *SaveDocument) map[string]bool {
	wanted := map[string]bool{}
	if doc == nil {
		return wanted
	}
	for _, win := range doc.Windows {
		for _, tab := range win.Tabs {
			if key := criticalMatchKey(tab); key != "" {
				wanted[key] = true
			}
		}
	}
	return wanted
}

type liveCriticalCandidate struct {
	pane     liveCriticalPane
	procs    []rawProc
	chosen   rawProc
	mayAgent bool
	mark     rawProc
	hasMark  bool
}

func wantedLiveITermIDs(doc *SaveDocument) map[string]bool {
	ids := map[string]bool{}
	if doc == nil {
		return ids
	}
	for _, win := range doc.Windows {
		for _, tab := range win.Tabs {
			if id := normalizeITermSessionID(tab.ItermSessionID); id != "" {
				ids[id] = true
			}
		}
	}
	return ids
}

func wantedLiveAgentCounts(doc *SaveDocument) map[string]int {
	counts := map[string]int{}
	if doc == nil {
		return counts
	}
	for _, win := range doc.Windows {
		for _, tab := range win.Tabs {
			key := criticalMatchKey(tab)
			if strings.HasPrefix(key, "grok:") || strings.HasPrefix(key, "codex:") {
				counts[key]++
			}
		}
	}
	return counts
}

func liveAgentMatchesMissing(found, wanted map[string]int) bool {
	for key, count := range wanted {
		if found[key] < count {
			return true
		}
	}
	return false
}

func matchingLiveMark(procs []rawProc, wanted map[string]bool) (rawProc, bool) {
	for i := len(procs) - 1; i >= 0; i-- {
		if !isMarkCmdline(procs[i].Command) {
			continue
		}
		if wanted["mark:"+markMessageFromCmdline(procs[i].Command)] {
			return procs[i], true
		}
		return rawProc{}, false
	}
	return rawProc{}, false
}

func (idx *liveCriticalIndex) addMarkCandidate(candidate liveCriticalCandidate) {
	if !candidate.hasMark {
		return
	}
	message := markMessageFromCmdline(candidate.mark.Command)
	idx.add(liveCriticalHit{
		Kind:        "mark",
		ID:          message,
		SemanticKey: "mark:" + message,
		Name:        candidate.pane.Name,
		PID:         candidate.mark.PID,
		WindowID:    candidate.pane.WindowID,
		ITermID:     candidate.pane.ID,
	}, candidate.pane.ID)
}

// resolveAgentCandidate records a wanted hard agent identity. Its return value
// reports whether the pane has an agent identity, which must win over mark.
func (idx *liveCriticalIndex) resolveAgentCandidate(candidate liveCriticalCandidate, resolve func(int) (*procresolve.Result, error), wanted map[string]bool, found, needed map[string]int) bool {
	res, _ := resolve(candidate.chosen.PID)
	if res == nil || (res.Kind != "grok" && res.Kind != "codex") || res.SessionID == "" {
		return false
	}
	key := res.Kind + ":" + res.SessionID
	if wanted[key] && found[key] < needed[key] {
		idx.add(liveCriticalHit{
			Kind:        res.Kind,
			ID:          res.SessionID,
			SemanticKey: key,
			Name:        candidate.pane.Name,
			PID:         firstPositive(res.RunnerPID, candidate.chosen.PID),
			WindowID:    candidate.pane.WindowID,
			ITermID:     candidate.pane.ID,
		}, candidate.pane.ID)
		found[key]++
	}
	return true
}

func indexLiveCriticalForKeys(snap *Snapshot, wanted map[string]bool) *liveCriticalIndex {
	all := indexLiveCritical(snap)
	out := newLiveCriticalIndex()
	for _, hit := range all.hits {
		if wanted[hit.SemanticKey] {
			out.add(hit, hit.ITermID)
		}
	}
	return out
}

func (idx *liveCriticalIndex) add(hit liveCriticalHit, itermSessionID string) {
	if idx == nil || hit.SemanticKey == "" {
		return
	}
	hitIndex := len(idx.hits)
	idx.hits = append(idx.hits, hit)
	if key := normalizeITermSessionID(itermSessionID); key != "" {
		idx.byITerm[key] = append(idx.byITerm[key], hitIndex)
	}
	idx.bySemantic[hit.SemanticKey] = append(idx.bySemantic[hit.SemanticKey], hitIndex)
}

func (c *SnapshotCollector) listLiveCriticalPanes() ([]liveCriticalPane, error) {
	runAS := c.RunAppleScript
	if runAS == nil {
		runAS = defaultRunAppleScript
	}
	raw, err := runAS(listLiveCriticalPanesAppleScript(c.AppTell))
	if err != nil {
		return nil, fmt.Errorf("Error: failed to query iTerm2: %w", err)
	}
	panes, err := parseLiveCriticalPanes(raw)
	if err != nil {
		return nil, err
	}
	return panes, nil
}

func listLiveCriticalPanesAppleScript(appTell string) string {
	return fmt.Sprintf(`
tell application %s
  set out to ""
  set windowCount to count of windows
  repeat with wi from 1 to windowCount
    try
      set wid to id of window wi
    on error
      set wid to 0
    end try
    try
      set tabCount to count of tabs of window wi
    on error
      set out to out & "###E###window " & wi & " changed while scanning" & linefeed
      set tabCount to 0
    end try
    repeat with ti from 1 to tabCount
      try
        set sessionCount to count of sessions of tab ti of window wi
      on error
        set out to out & "###E###window " & wi & ", tab " & ti & " changed while scanning" & linefeed
        set sessionCount to 0
      end try
      repeat with si from 1 to sessionCount
        try
          set ttyn to tty of session si of tab ti of window wi
        on error
          set out to out & "###E###window " & wi & ", tab " & ti & ", session " & si & " tty unavailable" & linefeed
          set ttyn to ""
        end try
        try
          set uid to unique ID of session si of tab ti of window wi
        on error
          set uid to ""
        end try
        try
          set nm to name of session si of tab ti of window wi
        on error
          set nm to ""
        end try
        set out to out & "###P###" & wid & "###" & ttyn & "###" & uid & "###" & nm & linefeed
      end repeat
    end repeat
  end repeat
  return out
end tell
`, appleScriptAppLiteral(appTell))
}

func parseLiveCriticalPanes(raw string) ([]liveCriticalPane, error) {
	var panes []liveCriticalPane
	for _, row := range strings.Split(raw, "\n") {
		row = strings.TrimRight(row, "\r")
		if strings.HasPrefix(row, "###E###") {
			return nil, fmt.Errorf("Error: iTerm2 changed while scanning live panes: %s", strings.TrimPrefix(row, "###E###"))
		}
		if !strings.HasPrefix(row, "###P###") {
			continue
		}
		parts := strings.SplitN(strings.TrimPrefix(row, "###P###"), "###", 4)
		if len(parts) != 4 {
			continue
		}
		windowID, err := strconv.ParseUint(strings.TrimSpace(parts[0]), 10, 64)
		if err != nil {
			windowID = 0
		}
		panes = append(panes, liveCriticalPane{
			WindowID: windowID,
			TTY:      parts[1],
			ID:       parts[2],
			Name:     parts[3],
		})
	}
	return panes, nil
}

func normalizeLiveTTY(tty string) string {
	return strings.TrimSpace(strings.TrimPrefix(tty, "/dev/"))
}

func baseNow(c *SnapshotCollector) time.Time {
	if c != nil && c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

func paneMayContainAgent(rootPID int, procs []procresolve.Proc) bool {
	if rootPID <= 0 {
		return false
	}
	byPID := make(map[int]procresolve.Proc, len(procs))
	children := make(map[int][]int, len(procs))
	for _, proc := range procs {
		byPID[proc.PID] = proc
		children[proc.PPID] = append(children[proc.PPID], proc.PID)
	}
	type candidate struct {
		pid   int
		depth int
	}
	queue := []candidate{{pid: rootPID}}
	seen := map[int]bool{rootPID: true}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		proc, ok := byPID[current.pid]
		if !ok {
			continue
		}
		switch commandBase(proc.Cmd) {
		case "grok", "codex":
			return true
		}
		if current.depth >= 16 {
			continue
		}
		for _, child := range children[current.pid] {
			if !seen[child] {
				seen[child] = true
				queue = append(queue, candidate{pid: child, depth: current.depth + 1})
			}
		}
	}
	return false
}

func liveCriticalAgentResolver(c *SnapshotCollector, procs []procresolve.Proc) func(int) (*procresolve.Result, error) {
	if c != nil && c.ResolveFromPID != nil {
		return c.ResolveFromPID
	}
	lsofCache := map[int][]string{}
	return func(pid int) (*procresolve.Result, error) {
		return procresolve.ResolveFromPID(pid, procresolve.Options{
			ListProcs: func() []procresolve.Proc { return procs },
			Lsof: func(pid int) []string {
				if cached, ok := lsofCache[pid]; ok {
					return cached
				}
				paths := procresolve.LiveLsof(pid)
				lsofCache[pid] = paths
				return paths
			},
			EnrichInfo: false,
		})
	}
}

func firstPositive(values ...int) int {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}
