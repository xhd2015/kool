package iterm2

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestForegroundSaveAndAutoBackupCLI(t *testing.T) {
	InstallPhasedFixtureCollectorForTest(t, PhasedFixtureOpts{
		ITermRunning: true,
		Windows: []SnapshotWindow{{Index: 1, Name: "Commands", Tabs: []SnapshotTab{
			{Index: 1, Sessions: []SnapshotSession{{Index: 1, ID: "SERVER", TTY: "/dev/ttys010", Foreground: &ForegroundCommand{Cwd: "/work", Argv: []string{"python3", "-m", "http.server", "8000"}}}}},
			{Index: 2, Sessions: []SnapshotSession{{Index: 1, ID: "REVIEW", TTY: "/dev/ttys011", Foreground: &ForegroundCommand{Cwd: "/work", Argv: []string{"./deploy.sh"}, Observed: "./deploy.sh"}}}},
			{Index: 3, Sessions: []SnapshotSession{{Index: 1, ID: "IDLE", TTY: "/dev/ttys012"}}},
		}}},
		BusyTTYs: []string{"ttys010", "ttys011"}, IdleTTYs: []string{"ttys012"},
		BusyLeafByTTY: map[string]string{"ttys010": "python3 -m http.server 8000", "ttys011": "./deploy.sh"},
		CwdByTTY:      map[string]string{"ttys010": "/work", "ttys011": "/work"},
	})
	for _, command := range []string{"save", "auto-backup"} {
		t.Run(command, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "commands.json")
			args := []string{command, "--file", path, "--ignore-macos-space", "--no-color"}
			if command == "auto-backup" {
				args = append(args, "--once")
			}
			var stdout, stderr bytes.Buffer
			if err := runSessions(append(append([]string{}, args...), "--dry-run"), &stdout, &stderr); err != nil {
				t.Fatalf("%v %s", err, stderr.String())
			}
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatal("dry-run wrote checkpoint")
			}
			for _, want := range []string{"2 command", "restart", "review required", "./deploy.sh"} {
				if !strings.Contains(stdout.String(), want) {
					t.Fatalf("missing %s:\n%s", want, stdout.String())
				}
			}
			stdout.Reset()
			stderr.Reset()
			if err := runSessions(args, &stdout, &stderr); err != nil {
				t.Fatalf("%v %s", err, stderr.String())
			}
			doc, err := ReadSaveDocument(path)
			if err != nil {
				t.Fatal(err)
			}
			if doc.Version != 2 || doc.Summary.Sessions != 2 || doc.Summary.ByKind["command"] != 2 {
				t.Fatalf("%+v", doc)
			}
			if doc.Windows[0].Tabs[0].Command.RestorePolicy != "restart" || doc.Windows[0].Tabs[1].Command.RestorePolicy != "review" {
				t.Fatalf("%+v", doc.Windows[0].Tabs)
			}
			if !strings.Contains(stderr.String(), "warning:") {
				t.Fatal("missing review warning")
			}
		})
	}
}

func TestCommandMixedRestoreConsumesAfterResolve(t *testing.T) {
	installEmptyLiveCollectorForRestoreTest(t)
	path := filepath.Join(t.TempDir(), "commands.json")
	doc := &SaveDocument{Version: 2, Summary: SaveSummary{Sessions: 2}, Windows: []SaveWindow{{Tabs: []SaveTab{
		restartTestTab("/work", "python3", "-m", "http.server", "8000"),
		{Kind: "command", Cwd: "/work", SourceCmdLine: "./deploy.sh", Command: &SavedCommand{RestorePolicy: "review"}},
	}}}}
	if err := WriteSaveDocument(path, doc); err != nil {
		t.Fatal(err)
	}
	prev := sessionsRunRestoreAS
	calls := 0
	sessionsRunRestoreAS = func(script string) (string, error) {
		calls++
		if strings.Contains(script, "deploy") || !strings.Contains(script, "http.server") {
			t.Fatal(script)
		}
		return "", nil
	}
	t.Cleanup(func() { sessionsRunRestoreAS = prev })
	var stdout, stderr bytes.Buffer
	if err := runSessions([]string{"restore", "--file", path, "--ignore-macos-space"}, &stdout, &stderr); err != nil {
		t.Fatalf("%v %s", err, stderr.String())
	}
	if calls != 1 || !strings.Contains(stdout.String(), "1 windows / 1 tabs") {
		t.Fatalf("calls=%d stdout=%s", calls, stdout.String())
	}
	got, err := ReadSaveDocument(path)
	if err != nil || !got.IsConsumed() || len(got.Windows[0].Tabs) != 2 {
		t.Fatalf("%+v %v", got, err)
	}
}
