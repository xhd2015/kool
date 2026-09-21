//go:build ignore

package run

import (
	"strings"
	"testing"
)

func TestSkillShowContainsServerWorkflow(t *testing.T) {
	out := captureStdout(t, func() {
		if err := handleSkill([]string{"--show"}); err != nil {
			t.Fatalf("skill --show: %v", err)
		}
	})
	for _, want := range []string{"name: __PROJECT_NAME__", "__PROJECT_NAME__ server", "get", "post"} {
		if !strings.Contains(out, want) {
			t.Fatalf("skill content missing %q", want)
		}
	}
}

func TestSkillListPrintsName(t *testing.T) {
	out := captureStdout(t, func() {
		if err := handleSkill([]string{"--list"}); err != nil {
			t.Fatalf("skill --list: %v", err)
		}
	})
	if strings.TrimSpace(out) != "__PROJECT_NAME__" {
		t.Fatalf("skill --list output = %q", out)
	}
}

func TestSkillShowHeaderPrintsFrontmatter(t *testing.T) {
	out := captureStdout(t, func() {
		if err := handleSkill([]string{"--show", "--header"}); err != nil {
			t.Fatalf("skill --show --header: %v", err)
		}
	})
	if !strings.Contains(out, "name: __PROJECT_NAME__") {
		t.Fatalf("expected frontmatter header, got:\n%s", out)
	}
	if strings.Contains(out, "## Talk to the server") {
		t.Fatal("--header should print frontmatter only, not the body")
	}
}

func TestSkillHelp(t *testing.T) {
	out := captureStdout(t, func() {
		if err := handleSkill([]string{"--help"}); err != nil {
			t.Fatalf("skill --help: %v", err)
		}
	})
	if !strings.Contains(out, "--show") || !strings.Contains(out, "--install") {
		t.Fatalf("skill help should list actions, got:\n%s", out)
	}
}
