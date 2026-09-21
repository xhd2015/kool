//go:build ignore

// Package skill embeds the __PROJECT_NAME__ agent skill shown and installed
// by `__PROJECT_NAME__ skill`.
package skill

import (
	_ "embed"
)

// SkillDirName is the directory name used when the skill is installed
// (e.g. ~/.agents/skills/__PROJECT_NAME__).
const SkillDirName = "__PROJECT_NAME__"

//go:embed SKILL.md
var skillMD string

// GetSkillContent returns the embedded SKILL.md content.
func GetSkillContent() string {
	return skillMD
}
