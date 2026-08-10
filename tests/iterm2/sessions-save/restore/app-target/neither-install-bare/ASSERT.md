## Expected

- Exit 0 (not hard error — Open1)
- stderr `warning:` about missing install / bare fallback (not unrelated noise)
- Would restore still printed (plans restore)
- Plan or warn signals **bare** create target (not a concrete home/system path as
  the resolved install) — e.g. `restore target` with bare `iTerm2`, or warning
  that neither install was found / falling back to bare
- not stamped

## Exit Code

- 0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("neither install must not hard-fail; exit=%d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	errOut := resp.Stderr
	errLow := strings.ToLower(errOut)
	if !strings.Contains(errLow, "warning") {
		t.Fatalf("neither install must warn; stderr=%q", errOut)
	}
	// Warning must be about install/disk/fallback — not an unrelated soft warn.
	// Avoid matching mere "iterm2" in paths (source names, file paths).
	warnAboutMissing := strings.Contains(errLow, "not found") ||
		strings.Contains(errLow, "missing") ||
		strings.Contains(errLow, "neither") ||
		strings.Contains(errLow, "no install") ||
		strings.Contains(errLow, "fallback") ||
		strings.Contains(errLow, "bare") ||
		(strings.Contains(errLow, "install") &&
			(strings.Contains(errLow, "disk") ||
				strings.Contains(errLow, "exist") ||
				strings.Contains(errLow, "available") ||
				strings.Contains(errLow, "found")))
	if !warnAboutMissing {
		t.Fatalf("neither install warning should mention missing install / fallback; stderr=%q", errOut)
	}

	out := resp.Stdout
	if !strings.Contains(out, "Would restore") {
		t.Fatalf("still plans restore (bare fallback):\n%s", out)
	}
	// Must not claim a concrete home/system path as the resolved restore target.
	if hasRestoreTargetLine(out, fixtureAppHome) {
		t.Fatalf("neither install must not restore-target home path:\n%s", out)
	}
	if hasRestoreTargetLine(out, fixtureAppSystem) {
		t.Fatalf("neither install must not restore-target system path:\n%s", out)
	}
	// Prefer explicit bare cue in plan (restore target iTerm2 / bare) once landed.
	outLow := strings.ToLower(out)
	bareInPlan := strings.Contains(outLow, "restore target") &&
		(strings.Contains(outLow, "bare") ||
			strings.Contains(out, `"iTerm2"`) ||
			strings.Contains(outLow, "application \"iterm2\"") ||
			// line like: restore target  iTerm2
			strings.Contains(outLow, "restore target") && strings.Contains(out, "iTerm2"))
	if !bareInPlan && !warnAboutMissing {
		t.Fatalf("expected bare restore target in plan or missing-install warn:\nstdout=%s\nstderr=%s",
			out, errOut)
	}
	// When warn is specific enough, bareInPlan is optional; still require that
	// we did not pick a concrete path (checked above). If warn is specific,
	// that alone satisfies Open1 + R8 for dry-run.
	_ = bareInPlan

	if resp.Doc == nil || resp.Doc.IsConsumed() {
		t.Fatal("dry-run must not stamp restored_at")
	}
}
```
