package iterm2

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SavedCommand describes evidence, not arbitrary shell text to execute. ResumeCmd
// is display-only for command entries; restore regenerates it from verified argv.
type SavedCommand struct {
	Argv          []string `json:"argv,omitempty"`
	Source        string   `json:"source"`
	Category      string   `json:"category,omitempty"`
	RestorePolicy string   `json:"restore_policy"`
	Reason        string   `json:"reason,omitempty"`
}

func quoteCommandArgv(argv []string) string {
	parts := make([]string, len(argv))
	for i, arg := range argv {
		parts[i] = shellQuoteArgIfNeeded(arg)
	}
	return strings.Join(parts, " ")
}

// shellQuoteArgIfNeeded quotes a token only when it is empty or contains
// anything outside the POSIX safe set [A-Za-z0-9_@%+=:,./-]; otherwise the
// token is emitted verbatim. Unquoted tokens expand exactly as written, so
// this preserves argv semantics while keeping resume_cmd human-readable.
func shellQuoteArgIfNeeded(s string) string {
	if s == "" {
		return "''"
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case strings.ContainsRune("_@%+=:,./-", r):
		default:
			return shellSingleQuote(s)
		}
	}
	return s
}

func validCommandCwd(cwd string) bool {
	return filepath.IsAbs(cwd) && validCommandArgv([]string{cwd})
}

func validCommandArgv(argv []string) bool {
	if len(argv) == 0 || argv[0] == "" || strings.HasPrefix(argv[0], "-") {
		return false
	}
	for _, arg := range argv {
		// Terminal write-text is not a binary argv transport. Reject control bytes,
		// including newline, even though ordinary shell metacharacters can be quoted.
		for _, r := range arg {
			if r < 32 || r == 127 {
				return false
			}
		}
	}
	return true
}

// serverInvocation recognizes a deliberately narrow set of restartable servers.
// A script named "dev" alone is not evidence: inspect the script and lifecycle
// hooks rather than automatically re-running an arbitrary deploy/migration.
func serverInvocation(argv []string, cwd string) bool {
	if !validCommandArgv(argv) {
		return false
	}
	args := argv
	base := filepath.Base(args[0])
	if base == "node" || base == "nodejs" {
		if len(args) < 2 || strings.HasPrefix(args[1], "-") {
			return false
		}
		args = args[1:]
		base = filepath.Base(args[0])
	}
	switch base {
	case "vite", "vite.js":
		return viteDevArguments(args[1:])
	case "next", "next.js", "astro":
		return len(args) >= 2 && args[1] == "dev"
	case "python", "python3":
		return len(args) >= 3 && args[1] == "-m" && args[2] == "http.server"
	case "npm", "npm-cli.js", "pnpm", "pnpm.cjs", "yarn", "yarn.js":
		rest := args[1:]
		if len(rest) > 0 && (rest[0] == "run" || rest[0] == "run-script") {
			rest = rest[1:]
		}
		if len(rest) == 0 || (rest[0] != "dev" && rest[0] != "start" && rest[0] != "serve") {
			return false
		}
		script := rest[0]
		// Reject workspace/config selectors and arbitrary extra shell fragments.
		for _, arg := range rest[1:] {
			if arg != "--" && !strings.HasPrefix(arg, "--host") && !strings.HasPrefix(arg, "--port=") {
				return false
			}
		}
		data, err := os.ReadFile(filepath.Join(cwd, "package.json"))
		if err != nil {
			return false
		}
		var pkg struct {
			Scripts map[string]string `json:"scripts"`
		}
		if json.Unmarshal(data, &pkg) != nil || pkg.Scripts["pre"+script] != "" || pkg.Scripts["post"+script] != "" {
			return false
		}
		text := pkg.Scripts[script]
		// Only simple script tokens; never parse shell syntax into trusted argv.
		if text == "" || strings.ContainsAny(text, "\"'`$;&|<>\\\n\r(){}*?!") {
			return false
		}
		words := strings.Fields(text)
		if len(words) == 0 {
			return false
		}
		switch filepath.Base(words[0]) {
		case "vite", "next", "astro", "python", "python3":
			return serverInvocation(words, cwd)
		}
	}
	return false
}

// viteDevArguments rejects command positionals (notably build/preview) even
// after options. A first token starting with -- alone does not imply a server.
func viteDevArguments(args []string) bool {
	if len(args) > 0 && (args[0] == "dev" || args[0] == "serve") {
		args = args[1:]
	}
	for i := 0; i < len(args); i++ {
		flag, _, inline := strings.Cut(args[i], "=")
		switch flag {
		case "--config", "-c", "--base", "--mode", "-m", "--logLevel", "-l", "--port":
			if !inline {
				i++
				if i >= len(args) || strings.HasPrefix(args[i], "-") {
					return false
				}
			}
		case "--host", "--open":
			if !inline && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
			}
		case "--strictPort", "--force", "--cors":
			if inline {
				return false
			}
		case "--clearScreen":
			if !inline && i+1 < len(args) && (args[i+1] == "true" || args[i+1] == "false") {
				i++
			}
		default:
			return false
		}
	}
	return true
}

func classifyForegroundTab(base SaveTab, s SnapshotSession) (*SaveTab, string) {
	fg := s.Foreground
	if fg == nil {
		return nil, ""
	}
	base.Kind = "command"
	base.Cwd = fg.Cwd
	base.SourceCmdLine = fg.Observed
	base.Command = &SavedCommand{Argv: append([]string(nil), fg.Argv...), Source: "process_argv", RestorePolicy: "review", Reason: fg.Reason}
	if len(fg.Argv) == 0 {
		base.Command.Source = "process_listing"
	}
	if validCommandArgv(fg.Argv) {
		base.ResumeCmd = quoteCommandArgv(fg.Argv)
	}
	switch {
	case !validCommandCwd(base.Cwd):
		base.Command.Reason = "launcher cwd unavailable or not terminal-safe"
	case fg.Complex:
		base.Command.Reason = "compound foreground job requires review"
	case !validCommandArgv(fg.Argv):
		base.Command.Reason = "reliable terminal-safe argv unavailable"
	case fg.Reason != "":
	case serverInvocation(fg.Argv, base.Cwd):
		base.Command.Category = "dev-server"
		base.Command.RestorePolicy = "restart"
		return &base, ""
	default:
		base.Command.Reason = "unrecognized foreground command; automatic restart disabled"
	}
	return &base, fmt.Sprintf("warning: saved command pane %s for review: %s", shortID(s.ID), base.Command.Reason)
}

func commandReviewReason(tab SaveTab) string {
	if tab.Kind != "command" {
		return ""
	}
	c := tab.Command
	if c == nil {
		return "missing command evidence"
	}
	if c.RestorePolicy != "restart" {
		return firstNonEmpty(c.Reason, "automatic restart disabled")
	}
	if c.Source != "process_argv" || c.Reason != "" || !validCommandCwd(tab.Cwd) || !validCommandArgv(c.Argv) {
		return "incomplete or unsafe command evidence"
	}
	if !serverInvocation(c.Argv, tab.Cwd) {
		return "server invocation or package script no longer qualifies for restart"
	}
	return ""
}

func commandNeedsReview(tab SaveTab) bool { return commandReviewReason(tab) != "" }

func commandDisplay(tab SaveTab) string {
	if tab.Command != nil && validCommandArgv(tab.Command.Argv) {
		return quoteCommandArgv(tab.Command.Argv)
	}
	return tab.SourceCmdLine
}

func commandMatchKey(tab SaveTab) string {
	if tab.Command == nil || tab.Cwd == "" || !validCommandArgv(tab.Command.Argv) {
		return ""
	}
	key, _ := json.Marshal(struct {
		Cwd  string
		Argv []string
	}{filepath.Clean(tab.Cwd), tab.Command.Argv})
	return "command:" + string(key)
}

// commandRestartable reports whether a command tab has exact terminal-safe argv
// + cwd and can therefore be executed when auto-restarted or explicitly allowed.
// Restart-policy commands additionally need serverInvocation; user-approved
// review commands do not.
func commandRestartable(tab SaveTab) bool {
	c := tab.Command
	return c != nil && c.Source == "process_argv" && validCommandCwd(tab.Cwd) && validCommandArgv(c.Argv)
}

// commandAction returns how a command tab is treated on this restore:
// "" (non-command), "restart" (auto), "allow" (user resolved), "deny" (skip),
// or "ask" (needs confirmation). ResolvedAllow/ResolvedDeny are set by the
// decision preflight; recorded decisions win over --deny-unknown.
func commandAction(tab SaveTab) string {
	if tab.Kind != "command" || tab.Command == nil {
		return ""
	}
	if tab.ResolvedDeny {
		return "deny"
	}
	if tab.ResolvedAllow {
		return "allow"
	}
	if tab.Command.RestorePolicy == "restart" && commandReviewReason(tab) == "" {
		return "restart"
	}
	if commandRestartable(tab) {
		return "ask"
	}
	return "deny"
}

// commandExecutable reports whether the tab will be executed on this restore
// (non-command tabs always are; commands only when restart or user-allowed).
func commandExecutable(tab SaveTab) bool {
	if tab.Kind != "command" {
		return true
	}
	a := commandAction(tab)
	return a == "restart" || a == "allow"
}

func leaveCommandReviewsPending(path string, doc *SaveDocument) error {
	if !doc.IsConsumed() {
		return nil
	}
	// --force may reopen an old checkpoint whose package scripts have changed.
	doc.RestoredAt = nil
	return WriteSaveDocument(path, doc)
}
