package iterm2

import (
	"fmt"
	"strings"

	lib "github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

// BuildSessionsRestoreScript returns AppleScript that creates one new iTerm2
// window per saved window and one tab per saved tab, then sends:
//
//	cd <cwd>
//	<resume_cmd>
//
// as two separate write text lines (empty cwd skips the cd line).
func BuildSessionsRestoreScript(doc *SaveDocument) string {
	if doc == nil || len(doc.Windows) == 0 {
		return `tell application "iTerm2"
  activate
end tell`
	}

	lines := []string{
		`tell application "iTerm2"`,
		`  activate`,
	}

	for _, win := range doc.Windows {
		lines = append(lines, `  set newWindow to (create window with default profile)`)
		for i, tab := range win.Tabs {
			if i == 0 {
				lines = append(lines, `  tell current session of newWindow`)
			} else {
				lines = append(lines,
					`  tell newWindow`,
					`    set newTab to (create tab with default profile)`,
					`    tell current session of newTab`,
				)
			}

			if tab.Cwd != "" {
				escapedCwd := lib.EscapePathForAppleScript(tab.Cwd)
				lines = append(lines,
					fmt.Sprintf(`    write text ("cd " & quoted form of "%s")`, escapedCwd),
				)
			}
			if tab.ResumeCmd != "" {
				lines = append(lines,
					fmt.Sprintf(`    write text "%s"`, lib.EscapeCommandForAppleScript(tab.ResumeCmd)),
				)
			}

			if i == 0 {
				lines = append(lines, `  end tell`)
			} else {
				lines = append(lines,
					`    end tell`,
					`  end tell`,
				)
			}
		}
		if win.Name != "" {
			lines = append(lines,
				fmt.Sprintf(`  set name of newWindow to "%s"`, lib.EscapeCommandForAppleScript(win.Name)),
			)
		}
	}

	lines = append(lines, `end tell`)
	return strings.Join(lines, "\n")
}
