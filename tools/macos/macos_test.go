package macos

import (
	"bytes"
	"strings"
	"testing"

	lib "github.com/xhd2015/dot-pkgs/go-pkgs/computer-use/macos/space"
	"github.com/xhd2015/kool/tools/macos/space"
)

func TestMacosHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := RunForTest([]string{"--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit %d stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "space") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestMacosSpaceHelp(t *testing.T) {
	space.ResetTestHooks()
	defer space.ResetTestHooks()
	var stdout, stderr bytes.Buffer
	code := RunForTest([]string{"space", "--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit %d stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "--run") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestMacosSpaceCreateMock(t *testing.T) {
	space.ResetTestHooks()
	defer space.ResetTestHooks()
	mock := &lib.MockBackend{Desktops: []int{1}}
	space.SetBackendForTest(mock)
	space.SetSettleMSForTest(0)

	var stdout, stderr bytes.Buffer
	code := RunForTest([]string{"space", "create"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit %d stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "desktop=2") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestMacosUnknown(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := RunForTest([]string{"nope"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected failure")
	}
	if !strings.Contains(stderr.String(), "unrecognized") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}
