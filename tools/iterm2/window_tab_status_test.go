package iterm2

import (
	"bytes"
	"strings"
	"testing"

	lib "github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func testStatusEnv() TestRun {
	refs := []lib.SessionRef{
		{WindowID: "23473", WindowName: "Project", TabIndex: 1, SessionID: "E209EDD0-6149-4972-94B0-3816091267B0", TTY: "/dev/ttys143", Name: "grok"},
		{WindowID: "23473", WindowName: "Project", TabIndex: 2, SessionID: "C057E2A3-1111-2222-3333-444444444444", TTY: "/dev/ttys163", Name: "bash"},
	}
	return TestRun{
		CurrentStatus: &lib.CurrentStatusConfig{
			SessionID:    func() string { return "w0t0p0:E209EDD0-6149-4972-94B0-3816091267B0" },
			ListSessions: func() ([]lib.SessionRef, error) { return refs, nil },
		},
	}
}

func TestWindowStatus_CLI(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := RunForTestEnv([]string{"window", "status"}, &stdout, &stderr, testStatusEnv())
	if code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "window 23473") || !strings.Contains(out, "* [1]") || !strings.Contains(out, "  [2]") {
		t.Fatalf("stdout=%q", out)
	}
	if !strings.Contains(out, "E209EDD0-6149-4972-94B0-3816091267B0") {
		t.Fatalf("want full session id: %q", out)
	}
	if strings.Contains(out, "E209EDD0…") {
		t.Fatalf("session id must not be truncated: %q", out)
	}
}

func TestTabStatus_CLI(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := RunForTestEnv([]string{"tab", "status"}, &stdout, &stderr, testStatusEnv())
	if code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "tab 1 of window 23473") || !strings.Contains(out, "/dev/ttys143") {
		t.Fatalf("stdout=%q", out)
	}
}

func TestWindowStatus_NotInSession(t *testing.T) {
	env := TestRun{
		CurrentStatus: &lib.CurrentStatusConfig{
			SessionID:      func() string { return "" },
			ListSessions:   func() ([]lib.SessionRef, error) { return nil, nil },
			ControllingTTY: func() string { return "" },
			AncestorTTYs:   func() []string { return nil },
		},
	}
	var stdout, stderr bytes.Buffer
	code := RunForTestEnv([]string{"window", "status"}, &stdout, &stderr, env)
	if code != 1 {
		t.Fatalf("exit=%d", code)
	}
	if !strings.Contains(stderr.String(), "Error:") || !strings.Contains(stderr.String(), "not inside") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestWindow_HelpAndUnknown(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := RunForTestEnv([]string{"window", "--help"}, &stdout, &stderr, TestRun{}); code != 0 {
		t.Fatalf("help exit=%d", code)
	}
	if !strings.Contains(stdout.String(), "window status") {
		t.Fatalf("help=%q", stdout.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := RunForTestEnv([]string{"window", "foo"}, &stdout, &stderr, TestRun{}); code != 1 {
		t.Fatalf("unknown exit=%d", code)
	}
	if !strings.Contains(stderr.String(), `unknown command "foo"`) {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestTab_HelpAndUnknown(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := RunForTestEnv([]string{"tab", "-h"}, &stdout, &stderr, TestRun{}); code != 0 {
		t.Fatalf("help exit=%d", code)
	}
	if !strings.Contains(stdout.String(), "tab status") {
		t.Fatalf("help=%q", stdout.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := RunForTestEnv([]string{"tab", "foo"}, &stdout, &stderr, TestRun{}); code != 1 {
		t.Fatalf("unknown exit=%d", code)
	}
}

func TestRootHelp_ListsWindowTabStatus(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := RunForTestEnv([]string{"--help"}, &stdout, &stderr, TestRun{}); code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "window status") || !strings.Contains(out, "tab status") {
		t.Fatalf("root help missing window/tab status: %q", out)
	}
}
