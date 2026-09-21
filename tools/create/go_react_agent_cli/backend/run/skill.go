//go:build ignore

package run

import (
	"github.com/xhd2015/skills/skillcmd"

	"__MODULE_NAME__/skill"
)

// handleSkill runs `__PROJECT_NAME__ skill`: show, list, version, or install
// the embedded __PROJECT_NAME__ agent skill. With no action it prints the
// skill help.
func handleSkill(args []string) error {
	skill := skillcmd.SingleSkill{
		Name:        skill.SkillDirName,
		RootContent: skill.GetSkillContent(),
		Usage:       "__PROJECT_NAME__ skill --install",
	}
	if len(args) == 0 {
		return skill.Handle([]string{"--help"})
	}
	return skill.Handle(args)
}
