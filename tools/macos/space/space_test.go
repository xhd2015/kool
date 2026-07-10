package space

import (
	"bytes"
	"strings"
	"testing"

	lib "github.com/xhd2015/dot-pkgs/go-pkgs/computer-use/macos/space"
)

func TestSplitRun(t *testing.T) {
	left, right, has := splitRun([]string{"create", "--run", "echo", "hi"})
	if !has || strings.Join(left, " ") != "create" || strings.Join(right, " ") != "echo hi" {
		t.Fatalf("got left=%v right=%v has=%v", left, right, has)
	}
}

func TestHelpLevels(t *testing.T) {
	cases := []struct {
		args    []string
		wantSub string
	}{
		{nil, "macos space"},
		{[]string{"--help"}, "macos space"},
		{[]string{"create", "--help"}, "macos space create"},
		{[]string{"switch", "--help"}, "macos space switch"},
		{[]string{"list", "--help"}, "macos space list"},
	}
	for _, tc := range cases {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			ResetTestHooks()
			var stdout, stderr bytes.Buffer
			code := RunForTest(tc.args, &stdout, &stderr)
			if code != 0 {
				t.Fatalf("exit %d stderr=%q", code, stderr.String())
			}
			if !strings.Contains(stdout.String(), tc.wantSub) {
				t.Fatalf("stdout missing %q:\n%s", tc.wantSub, stdout.String())
			}
		})
	}
}

func TestCreateOnly(t *testing.T) {
	ResetTestHooks()
	defer ResetTestHooks()
	mock := &lib.MockBackend{Desktops: []int{1, 2}}
	SetBackendForTest(mock)
	SetSettleMSForTest(0)
	runner := &RecordingRunner{}
	SetRunnerForTest(runner)

	var stdout, stderr bytes.Buffer
	code := RunForTest([]string{"create"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit %d stderr=%q", code, stderr.String())
	}
	if mock.Created != 1 {
		t.Fatalf("created=%d", mock.Created)
	}
	if len(mock.Switched) != 0 {
		t.Fatalf("should not switch without --run: %v", mock.Switched)
	}
	if len(runner.Calls) != 0 {
		t.Fatalf("unexpected run: %v", runner.Calls)
	}
	if !strings.Contains(stdout.String(), "created=true") || !strings.Contains(stdout.String(), "desktop=3") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestCreateRun(t *testing.T) {
	ResetTestHooks()
	defer ResetTestHooks()
	mock := &lib.MockBackend{Desktops: []int{1, 2}}
	SetBackendForTest(mock)
	SetSettleMSForTest(0)
	runner := &RecordingRunner{}
	SetRunnerForTest(runner)

	var stdout, stderr bytes.Buffer
	code := RunForTest([]string{"create", "--run", "kool", "iterm2", "-n", "."}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit %d stderr=%q", code, stderr.String())
	}
	if mock.Created != 1 {
		t.Fatalf("created=%d", mock.Created)
	}
	if len(mock.Switched) != 1 || mock.Switched[0] != 3 {
		t.Fatalf("switched=%v want [3]", mock.Switched)
	}
	got := strings.Join(runner.Calls[0], " ")
	if got != "kool iterm2 -n ." {
		t.Fatalf("run=%q", got)
	}
}

func TestSwitchRun(t *testing.T) {
	ResetTestHooks()
	defer ResetTestHooks()
	mock := &lib.MockBackend{Desktops: []int{1, 5, 12}}
	SetBackendForTest(mock)
	SetSettleMSForTest(0)
	runner := &RecordingRunner{}
	SetRunnerForTest(runner)

	var stdout, stderr bytes.Buffer
	code := RunForTest([]string{"switch", "12", "--run", "open", "-a", "Safari"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit %d stderr=%q", code, stderr.String())
	}
	got := strings.Join(runner.Calls[0], " ")
	if got != "open -a Safari" {
		t.Fatalf("run=%q", got)
	}
}

func TestList(t *testing.T) {
	ResetTestHooks()
	defer ResetTestHooks()
	SetBackendForTest(&lib.MockBackend{Desktops: []int{3, 1, 2}})
	var stdout, stderr bytes.Buffer
	code := RunForTest([]string{"list"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit %d stderr=%q", code, stderr.String())
	}
	if stdout.String() != "1\n2\n3\ncount=3\n" {
		t.Fatalf("got %q", stdout.String())
	}
}

func TestListRejectsRun(t *testing.T) {
	ResetTestHooks()
	defer ResetTestHooks()
	SetBackendForTest(&lib.MockBackend{Desktops: []int{1}})
	var stdout, stderr bytes.Buffer
	code := RunForTest([]string{"list", "--run", "true"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected failure")
	}
}

func TestNonDarwin(t *testing.T) {
	ResetTestHooks()
	defer ResetTestHooks()
	// no mock → hits library with GOOS override
	SetGOOSForTest("linux")
	var stdout, stderr bytes.Buffer
	code := RunForTest([]string{"list"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected failure")
	}
	if !strings.Contains(stderr.String(), "only supported on macOS") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestCreateRunEmpty(t *testing.T) {
	ResetTestHooks()
	defer ResetTestHooks()
	SetBackendForTest(&lib.MockBackend{Desktops: []int{1}})
	var stdout, stderr bytes.Buffer
	code := RunForTest([]string{"create", "--run"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected failure")
	}
}
