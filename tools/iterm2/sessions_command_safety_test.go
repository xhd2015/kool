package iterm2

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestCommandCheckpointPrivatePermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "checkpoint.json")
	for i := 0; i < 2; i++ {
		if i == 1 {
			if err := os.Chmod(path, 0644); err != nil {
				t.Fatal(err)
			}
		}
		if err := WriteSaveDocument(path, &SaveDocument{Version: 2}); err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0600 {
			t.Fatalf("checkpoint mode=%o", info.Mode().Perm())
		}
	}
}

func TestCommandScriptChangeDuringRestoreStaysPending(t *testing.T) {
	installEmptyLiveCollectorForRestoreTest(t)
	dir := t.TempDir()
	pkgPath := filepath.Join(dir, "package.json")
	if err := os.WriteFile(pkgPath, []byte(`{"scripts":{"dev":"vite"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "commands.json")
	doc := &SaveDocument{Version: 2, Summary: SaveSummary{Sessions: 1}, Windows: []SaveWindow{{Tabs: []SaveTab{restartTestTab(dir, "pnpm", "dev")}}}}
	if err := WriteSaveDocument(path, doc); err != nil {
		t.Fatal(err)
	}
	prev := sessionsRunRestoreAS
	sessionsRunRestoreAS = func(string) (string, error) {
		return "", os.WriteFile(pkgPath, []byte(`{"scripts":{"dev":"./deploy.sh"}}`), 0600)
	}
	t.Cleanup(func() { sessionsRunRestoreAS = prev })
	var stdout, stderr bytes.Buffer
	if err := runSessions([]string{"restore", "--file", path, "--ignore-macos-space"}, &stdout, &stderr); err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	got, err := ReadSaveDocument(path)
	if err != nil || got.IsConsumed() {
		t.Fatalf("script changed but consumed: %+v %v", got, err)
	}
}

func TestCommandForcedReviewConsumesResolvedCheckpoint(t *testing.T) {
	installEmptyLiveCollectorForRestoreTest(t)
	path := filepath.Join(t.TempDir(), "commands.json")
	consumed := "2026-09-20T10:00:00Z"
	doc := &SaveDocument{Version: 2, RestoredAt: &consumed, Summary: SaveSummary{Sessions: 1}, Windows: []SaveWindow{{Tabs: []SaveTab{{Kind: "command", Command: &SavedCommand{RestorePolicy: "review"}}}}}}
	if err := WriteSaveDocument(path, doc); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if err := runSessions([]string{"restore", "--force", "--file", path, "--ignore-macos-space"}, &stdout, &stderr); err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	// Non-executable review commands are auto-denied; nothing actionable
	// remains, so the checkpoint is consumed again.
	got, err := ReadSaveDocument(path)
	if err != nil || !got.IsConsumed() {
		t.Fatalf("expected consumed: %+v %v", got, err)
	}
}
