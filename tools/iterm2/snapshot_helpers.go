package iterm2

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Helpers retained in kool for restore/live_scan/tests. Hierarchy capture itself
// lives in shell/iterm2/snapshot.

func appleScriptAppLiteral(appTell string) string {
	appTell = strings.TrimSpace(appTell)
	if appTell == "" || appTell == "iTerm2" {
		return `"iTerm2"`
	}
	// Expand ~ for AppleScript path tell.
	if strings.HasPrefix(appTell, "~/") {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			appTell = home + appTell[1:]
		}
	}
	esc := strings.ReplaceAll(appTell, `\`, `\\`)
	esc = strings.ReplaceAll(esc, `"`, `\"`)
	return `"` + esc + `"`
}

func listTabsAndSessionsAppleScript(windowIndex int, appTell string) string {
	// Use numeric indexes throughout. Nested `repeat with t in tabs of w` object
	// references are unstable for path-targeted iTerm instances and can resolve
	// as invalid indexes while windows or tabs are active.
	return fmt.Sprintf(`
tell application %s
  set out to ""
  set target to %d
  set windowCount to count of windows
  if target is less than or equal to windowCount then
    set tabCount to count of tabs of window target
    repeat with ti from 1 to tabCount
      try
        set tname to name of current session of tab ti of window target
      on error
        set tname to ""
      end try
      set out to out & "###T###" & ti & "###" & tname & linefeed
      set sessionCount to count of sessions of tab ti of window target
      repeat with si from 1 to sessionCount
        try
          set nm to name of session si of tab ti of window target
        on error
          set nm to "?"
        end try
        try
          set ttyn to tty of session si of tab ti of window target
        on error
          set ttyn to ""
        end try
        try
          set prof to profile name of session si of tab ti of window target
        on error
          set prof to ""
        end try
        try
          set proc to is processing of session si of tab ti of window target
        on error
          set proc to false
        end try
        try
          set uid to unique ID of session si of tab ti of window target
        on error
          set uid to ""
        end try
        set out to out & "###S###" & si & "###" & ttyn & "###" & proc & "###" & prof & "###" & uid & "###" & nm & linefeed
      end repeat
    end repeat
  end if
  return out
end tell
`, appleScriptAppLiteral(appTell), windowIndex)
}

func commandBase(cmd string) string {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return ""
	}
	first := strings.Fields(cmd)[0]
	base := filepath.Base(first)
	return strings.TrimPrefix(base, "-")
}

func zoneOffset(t time.Time) string {
	_, off := t.Zone()
	sign := "+"
	if off < 0 {
		sign = "-"
		off = -off
	}
	h := off / 3600
	m := (off % 3600) / 60
	return sign + twoDigit(h) + twoDigit(m)
}

func twoDigit(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}

var redactPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(--app-secret=)\S+`),
	regexp.MustCompile(`(--app-secret\s+)\S+`),
	regexp.MustCompile(`(--secret=)\S+`),
	regexp.MustCompile(`(--token=)\S+`),
	regexp.MustCompile(`(--password=)\S+`),
	regexp.MustCompile(`(?i)(Authorization:\s*Bearer\s+)\S+`),
}

func redactCommandLine(s string) string {
	out := s
	for _, re := range redactPatterns {
		out = re.ReplaceAllString(out, `${1}***`)
	}
	return out
}
