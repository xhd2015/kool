//go:build ignore

package skill

import (
	"strings"
	"testing"
)

func TestGetSkillContentHasFrontmatter(t *testing.T) {
	content := GetSkillContent()
	if !strings.HasPrefix(content, "---\nname: __PROJECT_NAME__\n") {
		t.Fatalf("SKILL.md should start with __PROJECT_NAME__ frontmatter, got:\n%.80s", content)
	}
	if !strings.Contains(content, "description:") || !strings.Contains(content, "metadata:") {
		t.Fatal("SKILL.md frontmatter should carry description and metadata")
	}
}

func TestGetSkillContentTeachesWorkflow(t *testing.T) {
	content := GetSkillContent()
	for _, want := range []string{"__PROJECT_NAME__ server", "__PROJECT_NAME__ get", "__PROJECT_NAME__ post", "/api/counter", "--dry-run", "http://localhost:8080"} {
		if !strings.Contains(content, want) {
			t.Fatalf("SKILL.md missing %q", want)
		}
	}
}

func TestSkillDirName(t *testing.T) {
	if SkillDirName != "__PROJECT_NAME__" {
		t.Fatalf("SkillDirName = %q, want __PROJECT_NAME__", SkillDirName)
	}
}
