package iterm2

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestForegroundCheckpointClassification(t *testing.T) {
	for _, tc := range []struct {
		name   string
		fg     *ForegroundCommand
		policy string
	}{
		{"vite", &ForegroundCommand{Cwd: "/work", Argv: []string{"vite", "--host", "127.0.0.1"}}, "restart"},
		{"python", &ForegroundCommand{Cwd: "/work", Argv: []string{"python3", "-m", "http.server", "8000"}}, "restart"},
		{"unknown", &ForegroundCommand{Cwd: "/work", Argv: []string{"./deploy.sh"}}, "review"},
		{"pipeline", &ForegroundCommand{Cwd: "/work", Argv: []string{"vite"}, Complex: true}, "review"},
		{"missing cwd", &ForegroundCommand{Argv: []string{"vite"}}, "review"},
		{"ps only", &ForegroundCommand{Cwd: "/work", Observed: "vite --host"}, "review"},
		{"newline", &ForegroundCommand{Cwd: "/work", Argv: []string{"vite", "--host\necho bad"}}, "review"},
		{"cwd newline", &ForegroundCommand{Cwd: "/work\nother", Argv: []string{"vite"}}, "review"},
		{"relative cwd", &ForegroundCommand{Cwd: "work", Argv: []string{"vite"}}, "review"},
		{"vite build after options", &ForegroundCommand{Cwd: "/work", Argv: []string{"vite", "--mode", "production", "build"}}, "review"},
		{"vite preview after options", &ForegroundCommand{Cwd: "/work", Argv: []string{"vite", "--mode=production", "preview"}}, "review"},
		{"vite dev options", &ForegroundCommand{Cwd: "/work", Argv: []string{"vite", "dev", "--mode", "production", "--port", "4000"}}, "restart"},
		{"uncertain", &ForegroundCommand{Cwd: "/work", Argv: []string{"vite"}, Reason: "process changed"}, "review"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := SnapshotSession{ID: "pane", Foreground: tc.fg}
			tab, warning := classifyCriticalTab(SnapshotWindow{}, SnapshotTab{}, s)
			if tab == nil || tab.Kind != "command" || tab.Command.RestorePolicy != tc.policy {
				t.Fatalf("tab=%+v warning=%s", tab, warning)
			}
			if (warning != "") != (tc.policy == "review") {
				t.Fatalf("warning=%q", warning)
			}
			if commandNeedsReview(*tab) != (tc.policy == "review") {
				t.Fatal("policy mismatch")
			}
		})
	}
	s := SnapshotSession{Cwd: strPtr("/work"), Foreground: &ForegroundCommand{Cwd: "/work", Argv: []string{"vite"}}, Agent: &SessionAgent{Kind: "grok", SessionID: "g1"}}
	tab, _ := classifyCriticalTab(SnapshotWindow{}, SnapshotTab{}, s)
	if tab.Kind != "grok" {
		t.Fatal("agent must win")
	}
	s.Agent = nil
	s.Processes = []SnapshotProc{{Command: "mark waiting"}}
	tab, _ = classifyCriticalTab(SnapshotWindow{}, SnapshotTab{}, s)
	if tab.Kind != "mark" {
		t.Fatal("mark must win")
	}
	s = SnapshotSession{Idle: boolPtr(true)}
	if tab, _ := classifyCriticalTab(SnapshotWindow{}, SnapshotTab{}, s); tab != nil {
		t.Fatal("idle pane saved")
	}
}

func TestServerPackageScriptPolicy(t *testing.T) {
	dir := t.TempDir()
	for _, tc := range []struct {
		script, pre string
		want        bool
	}{
		{"vite", "", true}, {"next dev --port 3000", "", true},
		{"./deploy.sh", "", false}, {"vite && ./deploy.sh", "", false},
		{"vite", "./migrate.sh", false}, {"NODE_ENV=dev vite", "", false},
	} {
		data, _ := json.Marshal(map[string]any{"scripts": map[string]string{"dev": tc.script, "predev": tc.pre}})
		if err := os.WriteFile(filepath.Join(dir, "package.json"), data, 0600); err != nil {
			t.Fatal(err)
		}
		for _, argv := range [][]string{{"pnpm", "dev"}, {"npm", "run", "dev"}, {"node", "/opt/pnpm.cjs", "dev"}} {
			if got := serverInvocation(argv, dir); got != tc.want {
				t.Fatalf("%v script=%q pre=%q got=%v", argv, tc.script, tc.pre, got)
			}
		}
	}
}

func restartTestTab(cwd string, argv ...string) SaveTab {
	return SaveTab{Kind: "command", Cwd: cwd, Command: &SavedCommand{Argv: argv, Source: "process_argv", Category: "dev-server", RestorePolicy: "restart"}, ResumeCmd: "DO_NOT_EXECUTE_RESUME_TEXT"}
}

func TestCommandRestoreSafetyAndMultiplicity(t *testing.T) {
	tab := restartTestTab("/work with spaces", "python3", "-m", "http.server", "8000")
	review := SaveTab{Kind: "command", Cwd: "/work", ResumeCmd: "DANGEROUS_DEPLOY", Command: &SavedCommand{Argv: []string{"./deploy.sh"}, RestorePolicy: "review"}}
	doc := &SaveDocument{Windows: []SaveWindow{{Tabs: []SaveTab{tab, review, tab}}}}
	script := BuildSessionsRestoreScript(doc)
	if strings.Contains(script, "DANGEROUS") || strings.Contains(script, "DO_NOT_EXECUTE") {
		t.Fatal(script)
	}
	if !strings.Contains(script, "cd '/work with spaces' && python3 -m http.server 8000") {
		t.Fatal(script)
	}
	if strings.Count(script, "write text") != 2 {
		t.Fatal(script)
	}
	idx := newLiveCriticalIndex()
	idx.add(liveCriticalHit{Kind: "command", SemanticKey: criticalMatchKey(tab)}, "live")
	skips, n := matchCheckpointSkips(doc, idx, &bytes.Buffer{})
	if n != 1 || !skips[0][0] || skips[0][2] {
		t.Fatalf("skips=%v n=%d", skips, n)
	}
	w, n := countRemainingWouldCreate(doc, skips)
	if w != 1 || n != 1 {
		t.Fatalf("remaining=%d/%d", w, n)
	}
	other := tab
	other.Cwd = "/other"
	if criticalMatchKey(other) == criticalMatchKey(tab) {
		t.Fatal("cwd not in identity")
	}
}

func TestCommandCheckpointVersionAndRoundTrip(t *testing.T) {
	fg := &ForegroundCommand{Cwd: "/work", Argv: []string{"vite"}}
	snap := &Snapshot{Windows: []SnapshotWindow{{Tabs: []SnapshotTab{{Sessions: []SnapshotSession{{Foreground: fg}}}}}}}
	doc, warnings := BuildSaveDocument(snap, time.Now(), "test")
	if len(warnings) != 0 || doc.Version != 2 || doc.Summary.ByKind["command"] != 1 {
		t.Fatalf("%+v warnings=%v", doc, warnings)
	}
	path := filepath.Join(t.TempDir(), "checkpoint.json")
	if err := WriteSaveDocument(path, doc); err != nil {
		t.Fatal(err)
	}
	read, err := ReadSaveDocument(path)
	if err != nil || read.Windows[0].Tabs[0].Command.Argv[0] != "vite" {
		t.Fatalf("%+v %v", read, err)
	}
	// New readers keep legacy agent/mark checkpoints readable.
	doc.Version = 1
	doc.Windows = nil
	if err := WriteSaveDocument(path, doc); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadSaveDocument(path); err != nil {
		t.Fatal(err)
	}
	doc.Version = 99
	if err := WriteSaveDocument(path, doc); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadSaveDocument(path); err == nil {
		t.Fatal("future version accepted")
	}
}

func TestCommandReviewRestoreAutoDeniesNonExecutable(t *testing.T) {
	installEmptyLiveCollectorForRestoreTest(t)
	path := filepath.Join(t.TempDir(), "checkpoint.json")
	doc := &SaveDocument{Version: 2, Summary: SaveSummary{Sessions: 1}, Windows: []SaveWindow{{Tabs: []SaveTab{{Kind: "command", Cwd: "/work", SourceCmdLine: "./deploy.sh", ResumeCmd: "./deploy.sh", Command: &SavedCommand{RestorePolicy: "review", Reason: "unknown command"}}}}}}
	if err := WriteSaveDocument(path, doc); err != nil {
		t.Fatal(err)
	}
	prev := sessionsRunRestoreAS
	sessionsRunRestoreAS = func(string) (string, error) { t.Fatal("review entry executed AppleScript"); return "", nil }
	t.Cleanup(func() { sessionsRunRestoreAS = prev })

	// Dry-run: plan only, nothing consumed, no prompt needed (non-executable).
	var stdout, stderr bytes.Buffer
	if err := runSessions([]string{"restore", "--dry-run", "--file", path, "--ignore-macos-space"}, &stdout, &stderr); err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	if !strings.Contains(stderr.String(), "command skipped") {
		t.Fatal(stderr.String())
	}
	if got, _ := ReadSaveDocument(path); got.IsConsumed() {
		t.Fatal("dry-run must not consume")
	}

	// Live restore: no argv to run, so the entry is auto-denied and the
	// checkpoint is consumed (nothing actionable remains).
	for _, mode := range [][]string{{}, {"--force"}} {
		stdout.Reset()
		stderr.Reset()
		args := append([]string{"restore", "--file", path, "--ignore-macos-space"}, mode...)
		if err := runSessions(args, &stdout, &stderr); err != nil {
			t.Fatalf("%v %s", err, stderr.String())
		}
		if !strings.Contains(stderr.String(), "command skipped") {
			t.Fatal(stderr.String())
		}
		got, err := ReadSaveDocument(path)
		if err != nil || !got.IsConsumed() {
			t.Fatalf("expected consumed: %+v %v", got, err)
		}
	}
}

func TestCommandArgvShellQuoting(t *testing.T) {
	for _, tc := range []struct {
		name string
		argv []string
		want string
	}{
		{"simple", []string{"go", "run", "./script/ai-workshop/dev/"}, "go run ./script/ai-workshop/dev/"},
		{"space", []string{"vite", "--config", "config files/vite.config.ts"}, "vite --config 'config files/vite.config.ts'"},
		{"single quote", []string{"echo", "it's"}, "echo 'it'\\''s'"},
		{"shell expansion", []string{"echo", "$(touch nope); echo bad"}, "echo '$(touch nope); echo bad'"},
		{"empty", []string{"echo", ""}, "echo ''"},
		{"glob", []string{"echo", "*.go"}, "echo '*.go'"},
		{"tilde", []string{"echo", "~user"}, "echo '~user'"},
		{"hash", []string{"echo", "#comment"}, "echo '#comment'"},
		{"safe flag", []string{"vite", "--port=3000"}, "vite --port=3000"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := quoteCommandArgv(tc.argv); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestShellQuoteArgIfNeeded(t *testing.T) {
	for _, tc := range []struct {
		in, want string
	}{
		{"", "''"},
		{"go", "go"},
		{"/work/my-app", "/work/my-app"},
		{"/work my app", "'/work my app'"},
		{"a=b", "a=b"},
		{"-n", "-n"},
		{"--port=3000", "--port=3000"},
		{"it's", "'it'\\''s'"},
		{"a;b", "'a;b'"},
		{"$HOME", "'$HOME'"},
		{"`x`", "'`x`'"},
		{"*.go", "'*.go'"},
		{"~", "'~'"},
		{"#x", "'#x'"},
	} {
		if got := shellQuoteArgIfNeeded(tc.in); got != tc.want {
			t.Fatalf("shellQuoteArgIfNeeded(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
