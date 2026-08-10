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
//
// Uses bare tell application "iTerm2". Prefer BuildSessionsRestoreScriptWithTell
// when a path target is known.
//
// Window title setting is best-effort: each set name is wrapped in try/on error
// so a refused title does not abort remaining windows. Failures are returned
// as one line per title on stdout (empty stdout means all titles OK).
func BuildSessionsRestoreScript(doc *SaveDocument) string {
	return BuildSessionsRestoreScriptWithTell(doc, "")
}

// BuildSessionsRestoreScriptWithTell is BuildSessionsRestoreScript with an
// optional path-tell target. appTell is an absolute .app path (or expanded home
// path); empty or "iTerm2" uses bare tell application "iTerm2".
func BuildSessionsRestoreScriptWithTell(doc *SaveDocument, appTell string) string {
	tellLit := appleScriptAppLiteral(appTell)
	if doc == nil || len(doc.Windows) == 0 {
		return fmt.Sprintf(`tell application %s
  activate
end tell`, tellLit)
	}

	lines := []string{
		fmt.Sprintf(`tell application %s`, tellLit),
		`  activate`,
		`  set titleWarnings to ""`,
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
			escapedName := lib.EscapeCommandForAppleScript(win.Name)
			// Soft-fail: refused titles must not abort remaining windows/tabs.
			lines = append(lines,
				fmt.Sprintf(`  set desiredTitle to "%s"`, escapedName),
				`  try`,
				`    set name of newWindow to desiredTitle`,
				`  on error errMsg`,
				`    set titleWarnings to titleWarnings & "could not set window title \"" & desiredTitle & "\": " & errMsg & linefeed`,
				`  end try`,
			)
		}
	}

	lines = append(lines,
		`  return titleWarnings`,
		`end tell`,
	)
	return strings.Join(lines, "\n")
}
