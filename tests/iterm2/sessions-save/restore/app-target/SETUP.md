# Scenario

**Feature**: restore path targeting — default prefer-home by disk presence; opt-in `--same-app`

```
Caller
  -> seed checkpoint (optional windows[].app)
  -> inject RestoreAppDisk (both|system|home|neither)  # implementer wires hook
  -> sessions restore [--dry-run] [--same-app]
  <- restore target / recorded app / app plan lines; warn+fallback; not stamped on dry-run
```

## Context

- Canonical paths: `fixtureAppSystem` = `/Applications/iTerm.app`,
  `fixtureAppHome` = `~/Applications/iTerm.app` (never expanded in plan text).
- `Request.RestoreAppDisk` values: `RestoreDiskBoth` / `System` / `Home` /
  `Neither`. Product hook: `SetRestoreAppDiskForTest` (implementer).
- Default mode: one global **`restore target`**; **`recorded app`** only when
  it differs from target. Does not use recorded app for create.
- `--same-app`: per-window **`app  <path>`** create target.

## Steps

1. ModeRestore (shared by app-target leaves). Leaves set DryRun / SameApp /
   RestoreAppDisk / SeedRawJSON / MockRestoreAS.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeRestore
	return nil
}

func stripANSIRough(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && s[j] != 'm' {
				j++
			}
			if j < len(s) {
				i = j
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// metaPathAfter returns the path token after label on a plan meta line, or "".
// Uses exact field match so "~/Applications/…" does not count as "/Applications/…".
func metaPathAfter(line, label string) string {
	trim := stripANSIRough(strings.TrimSpace(line))
	idx := strings.Index(trim, label)
	if idx < 0 {
		return ""
	}
	rest := strings.TrimSpace(trim[idx+len(label):])
	if rest == "" {
		return ""
	}
	// Path is the first field (canonical paths have no spaces).
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

// hasRestoreTargetLine reports a dry-run meta line for the global restore target.
func hasRestoreTargetLine(out, app string) bool {
	for _, line := range strings.Split(out, "\n") {
		if metaPathAfter(line, "restore target") == app {
			return true
		}
	}
	return false
}

func hasRecordedAppLine(out, app string) bool {
	for _, line := range strings.Split(out, "\n") {
		if metaPathAfter(line, "recorded app") == app {
			return true
		}
	}
	return false
}

// hasSameAppCreateLine is true if a line looks like same-app create meta
// (`app  <path>`) and is not `recorded app` / `restore target`.
func hasSameAppCreateLine(out string) bool {
	for _, line := range strings.Split(out, "\n") {
		trim := stripANSIRough(strings.TrimSpace(line))
		if strings.HasPrefix(trim, "recorded app") || strings.Contains(trim, "restore target") {
			continue
		}
		if strings.HasPrefix(trim, "app ") || strings.HasPrefix(trim, "app\t") {
			return true
		}
	}
	return false
}

// hasSameAppCreateLineFor reports per-window create `app  <path>` with exact path.
func hasSameAppCreateLineFor(out, app string) bool {
	for _, line := range strings.Split(out, "\n") {
		trim := stripANSIRough(strings.TrimSpace(line))
		if strings.HasPrefix(trim, "recorded app") || strings.Contains(trim, "restore target") {
			continue
		}
		if !strings.HasPrefix(trim, "app ") && !strings.HasPrefix(trim, "app\t") {
			continue
		}
		if metaPathAfter(line, "app") == app {
			return true
		}
	}
	return false
}
```
