package iterm2

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	decisionsVersion = 1
	decisionAllow    = "allow"
	decisionDeny     = "deny"
)

// CommandDecision is one remembered allow/deny for an exact command (cwd + argv).
type CommandDecision struct {
	Key       string   `json:"key"`
	Cwd       string   `json:"cwd"`
	Argv      []string `json:"argv"`
	Decision  string   `json:"decision"` // allow | deny
	DecidedAt string   `json:"decided_at"`
}

// DecisionsDocument is the persisted restore-decision store. It lives outside
// checkpoints because checkpoints are overwritten on every save.
type DecisionsDocument struct {
	Version   int               `json:"version"`
	Decisions []CommandDecision `json:"decisions"`
}

// Injectable hooks for tests.
var (
	sessionsDecisionsPathForTest string // if set, overrides default path
)

// DefaultDecisionsPath returns ~/.config/iterm2/restore-decisions.json.
func DefaultDecisionsPath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = "."
	}
	return filepath.Join(home, ".config", "iterm2", "restore-decisions.json")
}

// effectiveDecisionsPath returns --file-less default (test override aware).
func effectiveDecisionsPath() string {
	if sessionsDecisionsPathForTest != "" {
		return sessionsDecisionsPathForTest
	}
	return DefaultDecisionsPath()
}

// ReadDecisionsDocument loads the store. Missing or corrupt files yield an
// empty store plus warnings; restore never blocks on the store being absent.
func ReadDecisionsDocument(path string) (*DecisionsDocument, []string) {
	var warnings []string
	doc := &DecisionsDocument{Version: decisionsVersion}
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			warnings = append(warnings, fmt.Sprintf("could not read decisions file %s: %v", path, err))
		}
		return doc, warnings
	}
	if err := json.Unmarshal(data, doc); err != nil {
		warnings = append(warnings, fmt.Sprintf("invalid decisions file %s: %v", path, err))
		return &DecisionsDocument{Version: decisionsVersion}, warnings
	}
	if doc.Version == 0 {
		doc.Version = decisionsVersion
	}
	if doc.Decisions == nil {
		doc.Decisions = []CommandDecision{}
	}
	return doc, warnings
}

// WriteDecisionsDocument writes the store atomically with 0600 — argv can
// contain credentials, so the file must never be world-readable.
func WriteDecisionsDocument(path string, doc *DecisionsDocument) error {
	if doc == nil {
		return fmt.Errorf("nil decisions document")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp, err := os.CreateTemp(filepath.Dir(path), ".restore-decisions-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// decisionFor returns the recorded decision for tab and whether one exists.
// Identity is the exact command (cwd + argv) via commandMatchKey.
func decisionFor(store *DecisionsDocument, tab SaveTab) (CommandDecision, bool) {
	if store == nil {
		return CommandDecision{}, false
	}
	key := commandMatchKey(tab)
	if key == "" {
		return CommandDecision{}, false
	}
	for _, d := range store.Decisions {
		if d.Key == key {
			return d, true
		}
	}
	return CommandDecision{}, false
}

// recordDecision upserts an allow/deny record for the exact command.
func recordDecision(store *DecisionsDocument, key, cwd string, argv []string, decision string, now time.Time) {
	if store == nil {
		return
	}
	rec := CommandDecision{
		Key:       key,
		Cwd:       cwd,
		Argv:      append([]string(nil), argv...),
		Decision:  decision,
		DecidedAt: now.Format("2006-01-02T15:04:05-0700"),
	}
	for i := range store.Decisions {
		if store.Decisions[i].Key == key {
			store.Decisions[i] = rec
			return
		}
	}
	store.Decisions = append(store.Decisions, rec)
}

func removeDecision(store *DecisionsDocument, key string) bool {
	if store == nil {
		return false
	}
	for i := range store.Decisions {
		if store.Decisions[i].Key == key {
			store.Decisions = append(store.Decisions[:i], store.Decisions[i+1:]...)
			return true
		}
	}
	return false
}

// promptCommandDecision asks the user how to treat one review command.
// Returns the decision and whether it should be remembered ("always").
func promptCommandDecision(tab SaveTab, stdout, stderr io.Writer) (string, bool, error) {
	prompt := fmt.Sprintf("  %s   (cwd %s)\n"+
		"    1) deny once    2) allow once    3) deny always    4) allow always   [default 1] ",
		commandDisplay(tab), tab.Cwd)
	for attempt := 0; attempt < 3; attempt++ {
		ans, err := sessionsReadConfirm(prompt, stdout, stderr)
		if err != nil {
			return "", false, err
		}
		switch strings.ToLower(strings.TrimSpace(ans)) {
		case "", "1", "deny once", "deny":
			return decisionDeny, false, nil
		case "2", "allow once", "allow":
			return decisionAllow, false, nil
		case "3", "deny always":
			return decisionDeny, true, nil
		case "4", "allow always":
			return decisionAllow, true, nil
		default:
			fmt.Fprintf(stderr, "    invalid choice %q (1-4)\n", ans)
		}
	}
	return "", false, fmt.Errorf("too many invalid choices")
}

// sortDecisionsByDisplay orders records by display command then cwd (list UX).
func sortDecisionsByDisplay(doc *DecisionsDocument) {
	if doc == nil {
		return
	}
	sort.SliceStable(doc.Decisions, func(i, j int) bool {
		di := strings.Join(doc.Decisions[i].Argv, " ")
		dj := strings.Join(doc.Decisions[j].Argv, " ")
		if di != dj {
			return di < dj
		}
		return doc.Decisions[i].Cwd < doc.Decisions[j].Cwd
	})
}

// listDecisions prints the decision store (numbered, sorted by display).
func listDecisions(stdout io.Writer, store *DecisionsDocument) {
	sortDecisionsByDisplay(store)
	if store == nil || len(store.Decisions) == 0 {
		fmt.Fprintln(stdout, "no saved decisions")
		return
	}
	for i, d := range store.Decisions {
		disp := strings.Join(d.Argv, " ")
		if disp == "" {
			disp = d.Key
		}
		fmt.Fprintf(stdout, "  %d. %-5s  %s   (cwd %s, saved %s)\n", i+1, d.Decision, disp, d.Cwd, d.DecidedAt)
	}
	fmt.Fprintf(stdout, "%d decisions\n", len(store.Decisions))
}
