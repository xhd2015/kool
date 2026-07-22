## Expected

- Exit 0 without requiring `--dry-run` (P1 seam retired).
- Marker `KOOL_TASKS_P2_HELLO` present (auto → local for single foreground leaf).
- Offline-safe: no iTerm dependency.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("auto single-local exit=%d stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := combinedOut(resp)
	if !strings.Contains(out, "KOOL_TASKS_P2_HELLO") {
		t.Fatalf("auto single fg leaf should exec local with marker; out:\n%s", out)
	}
	// Must not still be the P1 seam that demanded --dry-run.
	lower := strings.ToLower(out)
	if strings.Contains(lower, "not available") && strings.Contains(lower, "dry-run") {
		t.Fatalf("P1 requires-dry-run seam must be retired; out:\n%s", out)
	}
}
```
