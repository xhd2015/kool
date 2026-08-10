## Expected

- Exit 0 (soft fail, not hard Error abort)
- stderr contains `warning:` about iTerm/capture/snapshot
- fail-auto.json not created

## Exit Code

- 0

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("capture fail must be soft (exit 0); exit=%d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	errOut := resp.Stderr
	if !strings.Contains(errOut, "warning:") {
		t.Fatalf("capture fail must soft-warn on stderr; stderr=%q stdout=%q", errOut, resp.Stdout)
	}
	lowErr := strings.ToLower(errOut)
	scanish := strings.Contains(lowErr, "capture") ||
		strings.Contains(lowErr, "snapshot") ||
		strings.Contains(lowErr, "iterm") ||
		strings.Contains(lowErr, "not running") ||
		strings.Contains(lowErr, "not available") ||
		strings.Contains(lowErr, "failed")
	if !scanish {
		t.Fatalf("expected soft warn about capture/iTerm fail; stderr=%q", errOut)
	}
	// Must not hard-fail with Error: as the primary cycle outcome.
	if strings.HasPrefix(strings.TrimSpace(errOut), "Error:") && !strings.Contains(errOut, "warning:") {
		t.Fatalf("capture fail should not hard Error abort; stderr=%q", errOut)
	}
	p := filepath.Join(req.WorkingDir, "fail-auto.json")
	if _, e := os.Stat(p); !os.IsNotExist(e) {
		t.Fatalf("must not write %s on capture fail", p)
	}
}
```
