package iterm2

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// reviewAskTab builds an executable review command (valid argv + cwd) that
// needs a confirmation decision at restore time.
func reviewAskTab(cwd string, argv ...string) SaveTab {
	return SaveTab{
		Kind:          "command",
		Cwd:           cwd,
		SourceCmdLine: strings.Join(argv, " "),
		Command: &SavedCommand{
			Argv:          argv,
			Source:        "process_argv",
			RestorePolicy: "review",
			Reason:        "unrecognized foreground command; automatic restart disabled",
		},
	}
}

func decisionsTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	prev := sessionsDecisionsPathForTest
	sessionsDecisionsPathForTest = filepath.Join(dir, "restore-decisions.json")
	t.Cleanup(func() { sessionsDecisionsPathForTest = prev })
	return dir
}

func installTTYPrompt(t *testing.T, answers []string) {
	t.Helper()
	prevTTY := sessionsIsStdinTTY
	sessionsIsStdinTTY = func() bool { return true }
	t.Cleanup(func() { sessionsIsStdinTTY = prevTTY })
	prevConfirm := sessionsReadConfirm
	i := 0
	sessionsReadConfirm = func(prompt string, stdout, stderr io.Writer) (string, error) {
		if i >= len(answers) {
			return "", io.EOF
		}
		ans := answers[i]
		i++
		return ans, nil
	}
	t.Cleanup(func() { sessionsReadConfirm = prevConfirm })
}

func installNonTTY(t *testing.T) {
	t.Helper()
	prevTTY := sessionsIsStdinTTY
	sessionsIsStdinTTY = func() bool { return false }
	t.Cleanup(func() { sessionsIsStdinTTY = prevTTY })
}

// restoreFixture writes a checkpoint with one restart command and one review
// command, mocking AppleScript capture and installing an empty live collector
// so the already-running scan never touches real iTerm2.
func restoreFixture(t *testing.T, dir string, review SaveTab) (path string, scripts *[]string) {
	t.Helper()
	installEmptyLiveCollectorForRestoreTest(t)
	path = filepath.Join(dir, "checkpoint.json")
	doc := &SaveDocument{Version: 2, Summary: SaveSummary{Sessions: 2}, Windows: []SaveWindow{{Tabs: []SaveTab{restartTestTab("/work", "python3", "-m", "http.server", "8000"), review}}}}
	if err := WriteSaveDocument(path, doc); err != nil {
		t.Fatal(err)
	}
	prev := sessionsRunRestoreAS
	var gotScripts []string
	sessionsRunRestoreAS = func(script string) (string, error) {
		gotScripts = append(gotScripts, script)
		return "", nil
	}
	t.Cleanup(func() { sessionsRunRestoreAS = prev })
	return path, &gotScripts
}

func storeContents(t *testing.T) *DecisionsDocument {
	t.Helper()
	store, warns := ReadDecisionsDocument(effectiveDecisionsPath())
	if len(warns) > 0 {
		t.Fatalf("store warnings: %v", warns)
	}
	return store
}

func TestRestorePromptDenyOnce(t *testing.T) {
	dir := decisionsTestDir(t)
	installTTYPrompt(t, []string{"1"})
	path, _ := restoreFixture(t, dir, reviewAskTab("/work", "./deploy.sh"))
	var stdout, stderr bytes.Buffer
	if err := runSessions([]string{"restore", "--file", path, "--ignore-macos-space"}, &stdout, &stderr); err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	if strings.Contains(stdout.String(), "./deploy.sh") {
		t.Fatalf("deny once must not execute review command:\n%s", stdout.String())
	}
	if got, _ := ReadSaveDocument(path); !got.IsConsumed() {
		t.Fatal("expected consumed")
	}
	if n := len(storeContents(t).Decisions); n != 0 {
		t.Fatalf("deny once must not record: %d", n)
	}
}

func TestRestorePromptAllowOnce(t *testing.T) {
	dir := decisionsTestDir(t)
	installTTYPrompt(t, []string{"2"})
	path, scripts := restoreFixture(t, dir, reviewAskTab("/work", "./deploy.sh"))
	var stdout, stderr bytes.Buffer
	if err := runSessions([]string{"restore", "--file", path, "--ignore-macos-space"}, &stdout, &stderr); err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	if len(*scripts) != 1 || !strings.Contains((*scripts)[0], "./deploy.sh") {
		t.Fatalf("allow once must execute review command: %v", *scripts)
	}
	if got, _ := ReadSaveDocument(path); !got.IsConsumed() {
		t.Fatal("expected consumed")
	}
	if n := len(storeContents(t).Decisions); n != 0 {
		t.Fatalf("allow once must not record: %d", n)
	}
}

func TestRestorePromptDenyAlwaysRecords(t *testing.T) {
	dir := decisionsTestDir(t)
	installTTYPrompt(t, []string{"3"})
	path, _ := restoreFixture(t, dir, reviewAskTab("/work", "./deploy.sh"))
	var stdout, stderr bytes.Buffer
	if err := runSessions([]string{"restore", "--file", path, "--ignore-macos-space"}, &stdout, &stderr); err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	if !strings.Contains(stderr.String(), "recorded deny") {
		t.Fatal(stderr.String())
	}
	store := storeContents(t)
	if len(store.Decisions) != 1 || store.Decisions[0].Decision != decisionDeny {
		t.Fatalf("store=%+v", store)
	}

	// Second restore: recorded decision auto-applies with a notice, no prompt.
	prev := sessionsRunRestoreAS
	asCalls := 0
	var secondScripts []string
	sessionsRunRestoreAS = func(script string) (string, error) {
		asCalls++
		secondScripts = append(secondScripts, script)
		return "", nil
	}
	t.Cleanup(func() { sessionsRunRestoreAS = prev })
	// A prompt would return EOF and abort the restore; auto-apply must not prompt.
	installTTYPrompt(t, []string{}) // answers exhausted -> EOF on any prompt
	stdout.Reset()
	stderr.Reset()
	if err := runSessions([]string{"restore", "--force", "--file", path, "--ignore-macos-space"}, &stdout, &stderr); err != nil {
		t.Fatalf("second restore: %v %s", err, stderr.String())
	}
	if !strings.Contains(stderr.String(), "auto-denied") {
		t.Fatalf("missing auto-denied notice:\n%s", stderr.String())
	}
	if asCalls != 1 || strings.Contains(secondScripts[0], "./deploy.sh") {
		t.Fatalf("auto-denied command must not execute (restart tab still runs): calls=%d scripts=%v", asCalls, secondScripts)
	}
}

func TestRestorePromptAllowAlwaysRecordsAndRuns(t *testing.T) {
	dir := decisionsTestDir(t)
	installTTYPrompt(t, []string{"4"})
	path, _ := restoreFixture(t, dir, reviewAskTab("/work", "./deploy.sh"))
	var stdout, stderr bytes.Buffer
	if err := runSessions([]string{"restore", "--file", path, "--ignore-macos-space"}, &stdout, &stderr); err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	if !strings.Contains(stderr.String(), "recorded allow") {
		t.Fatal(stderr.String())
	}
	store := storeContents(t)
	if len(store.Decisions) != 1 || store.Decisions[0].Decision != decisionAllow {
		t.Fatalf("store=%+v", store)
	}
	// Second restore: auto-allowed and executed, no prompt.
	prev := sessionsRunRestoreAS
	asCalls := 0
	var secondScripts []string
	sessionsRunRestoreAS = func(script string) (string, error) {
		asCalls++
		secondScripts = append(secondScripts, script)
		return "", nil
	}
	t.Cleanup(func() { sessionsRunRestoreAS = prev })
	stdout.Reset()
	stderr.Reset()
	if err := runSessions([]string{"restore", "--force", "--file", path, "--ignore-macos-space"}, &stdout, &stderr); err != nil {
		t.Fatalf("second restore: %v %s", err, stderr.String())
	}
	if !strings.Contains(stderr.String(), "auto-allowed") {
		t.Fatalf("missing auto-allowed notice:\n%s", stderr.String())
	}
	if asCalls != 1 || !strings.Contains(secondScripts[0], "./deploy.sh") {
		t.Fatalf("auto-allowed command must execute: calls=%d scripts=%v", asCalls, secondScripts)
	}
}

func TestRestorePromptAbortLeavesUnconsumed(t *testing.T) {
	dir := decisionsTestDir(t)
	installNonTTY(t)
	path, _ := restoreFixture(t, dir, reviewAskTab("/work", "./deploy.sh"))
	var stdout, stderr bytes.Buffer
	err := runSessions([]string{"restore", "--file", path, "--ignore-macos-space"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected non-TTY confirmation error")
	}
	if !strings.Contains(stderr.String(), "not a TTY") {
		t.Fatal(stderr.String())
	}
	if !strings.Contains(stderr.String(), "--deny-unknown") {
		t.Fatal(stderr.String())
	}
	if got, _ := ReadSaveDocument(path); got.IsConsumed() {
		t.Fatal("must not consume on error")
	}
}

func TestRestoreDenyUnknownFlag(t *testing.T) {
	dir := decisionsTestDir(t)
	installNonTTY(t)
	path, _ := restoreFixture(t, dir, reviewAskTab("/work", "./deploy.sh"))
	var stdout, stderr bytes.Buffer
	if err := runSessions([]string{"restore", "--deny-unknown", "--file", path, "--ignore-macos-space"}, &stdout, &stderr); err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	if !strings.Contains(stderr.String(), "--deny-unknown skipped it") {
		t.Fatalf("missing deny-unknown notice:\n%s", stderr.String())
	}
	if strings.Contains(stdout.String(), "./deploy.sh") {
		t.Fatalf("deny-unknown must skip review command:\n%s", stdout.String())
	}
	if got, _ := ReadSaveDocument(path); !got.IsConsumed() {
		t.Fatal("expected consumed")
	}
	if n := len(storeContents(t).Decisions); n != 0 {
		t.Fatalf("--deny-unknown must not record: %d", n)
	}
}

func TestDecisionsListAndRm(t *testing.T) {
	decisionsTestDir(t)
	store := &DecisionsDocument{Version: 1}
	recordDecision(store, "command:{\"Cwd\":\"/work/a\",\"Argv\":[\"go\",\"run\",\"a\"]}", "/work/a", []string{"go", "run", "a"}, decisionAllow, sessionsNowFn())
	recordDecision(store, "command:{\"Cwd\":\"/work/b\",\"Argv\":[\"go\",\"run\",\"b\"]}", "/work/b", []string{"go", "run", "b"}, decisionDeny, sessionsNowFn())
	if err := WriteDecisionsDocument(effectiveDecisionsPath(), store); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	if err := runSessions([]string{"decisions", "list"}, &stdout, &stderr); err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "go run a") || !strings.Contains(stdout.String(), "go run b") || !strings.Contains(stdout.String(), "2 decisions") {
		t.Fatalf("list output:\n%s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if err := runSessions([]string{"decisions", "rm", "1"}, &stdout, &stderr); err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "removed 1 decision") {
		t.Fatal(stdout.String())
	}
	got := storeContents(t)
	if len(got.Decisions) != 1 || got.Decisions[0].Argv[2] != "b" {
		t.Fatalf("after rm: %+v", got)
	}

	// rm by exact key
	stdout.Reset()
	stderr.Reset()
	key := commandMatchKey(SaveTab{Cwd: "/work/b", Command: &SavedCommand{Argv: []string{"go", "run", "b"}}})
	if err := runSessions([]string{"decisions", "rm", key}, &stdout, &stderr); err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	got = storeContents(t)
	if len(got.Decisions) != 0 {
		t.Fatalf("after rm key: %+v", got)
	}

	// rm of missing -> error
	stderr.Reset()
	if err := runSessions([]string{"decisions", "rm", "99"}, &stdout, &stderr); err == nil {
		t.Fatal("expected rm error for missing index")
	}
	if !strings.Contains(stderr.String(), "no decision matches") {
		t.Fatal(stderr.String())
	}
}

func TestDecisionsFilePermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "restore-decisions.json")
	for i := 0; i < 2; i++ {
		if i == 1 {
			if err := os.Chmod(path, 0644); err != nil {
				t.Fatal(err)
			}
		}
		if err := WriteDecisionsDocument(path, &DecisionsDocument{Version: 1}); err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0600 {
			t.Fatalf("decisions mode=%o", info.Mode().Perm())
		}
	}
}

func TestRestoreDryRunShowsAskWithoutPrompt(t *testing.T) {
	dir := decisionsTestDir(t)
	installNonTTY(t) // even non-TTY, dry-run must not error
	path, _ := restoreFixture(t, dir, reviewAskTab("/work", "./deploy.sh"))
	var stdout, stderr bytes.Buffer
	if err := runSessions([]string{"restore", "--dry-run", "--file", path, "--ignore-macos-space"}, &stdout, &stderr); err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "would ask") || !strings.Contains(stdout.String(), "./deploy.sh") {
		t.Fatalf("dry-run must list ask marker:\n%s", stdout.String())
	}
	if got, _ := ReadSaveDocument(path); got.IsConsumed() {
		t.Fatal("dry-run must not consume")
	}
}
