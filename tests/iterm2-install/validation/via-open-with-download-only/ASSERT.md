## Expected

- Non-zero exit (prefer exit 1).
- Stderr contains `Error:` (hard error prefix).
- Message indicates incompatibility of `--via-open` with `--download-only`
  (both flag names, or clear “incompatible” / “cannot” wording with either flag).
- No zip written under `--download-dir`.

## Errors

- Incompatible flags: `--via-open` + `--download-only`.

## Exit Code

- non-zero

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for --via-open + --download-only; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stderr, "Error:") {
		t.Fatalf("stderr must contain Error:; got %q", resp.Stderr)
	}
	low := strings.ToLower(resp.Stderr)
	hasVia := strings.Contains(low, "via-open") || strings.Contains(low, "via open")
	hasDL := strings.Contains(low, "download-only") || strings.Contains(low, "download only")
	incompatible := strings.Contains(low, "incompatible") ||
		strings.Contains(low, "cannot") ||
		strings.Contains(low, "conflict") ||
		strings.Contains(low, "not allowed") ||
		strings.Contains(low, "mutually")
	if !(hasVia && hasDL) && !(incompatible && (hasVia || hasDL)) {
		t.Fatalf("stderr should explain --via-open vs --download-only conflict; got %q", resp.Stderr)
	}
	assertNoZip(t, resp)
}
```
