package scaffold

import (
	"bytes"
	"strings"
	"testing"
)

func TestHandleList(t *testing.T) {
	var buf bytes.Buffer
	if err := HandleWithWriter(&buf, []string{"--list"}); err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(buf.String())
	names := strings.Split(got, "\n")
	wantNames := map[string]bool{
		"go-cmd-run-lib": true,
		"github-publish": true,
	}
	if len(names) != len(wantNames) {
		t.Fatalf("list output has %d entries, want %d:\n%s", len(names), len(wantNames), got)
	}
	for _, name := range names {
		if !wantNames[name] {
			t.Fatalf("unexpected scaffold in list: %q", name)
		}
	}
}

func TestHandleGoCmdRunLib(t *testing.T) {
	var buf bytes.Buffer
	if err := HandleWithWriter(&buf, []string{"go-cmd-run-lib"}); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	for _, want := range []string{
		"# cmd/__NAME__/main.go",
		"# run/__NAME__/run.go",
		"# pkgs/__NAME__/__NAME__.go",
		`__NAME__ "__MODULE__/run/__NAME__"`,
		`core "__MODULE__/pkgs/__NAME__"`,
		"github.com/xhd2015/less-gen/flags",
		"func Run(args []string) error",
		"func Run(config Config) error",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("scaffold output missing %q:\n%s", want, got)
		}
	}
}

func TestHandleGitHubPublish(t *testing.T) {
	var buf bytes.Buffer
	if err := HandleWithWriter(&buf, []string{"github-publish"}); err != nil {
		t.Fatal(err)
	}

	got := buf.String()

	wantPaths := []string{
		"# script/release/main.go",
		"# script/lib/build_release.go",
		"# install.sh",
		"# README.md",
		"# .upload-credentials.json",
	}
	for _, want := range wantPaths {
		if !strings.Contains(got, want) {
			t.Fatalf("scaffold output missing file header %q:\n%s", want, got)
		}
	}

	wantContent := []string{
		`"github.com/xhd2015/kool/pkgs/github"`,
		`"github.com/xhd2015/kool/pkgs/release"`,
		"__NAME__",
		"__OWNER__",
		"__REPO__",
		"github.com/xhd2015/less-gen/flags",
		"func BuildRelease",
		"--dry-run",
	}
	for _, want := range wantContent {
		if !strings.Contains(got, want) {
			t.Fatalf("scaffold output missing %q:\n%s", want, got)
		}
	}
}
