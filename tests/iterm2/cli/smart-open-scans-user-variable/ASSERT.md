## Expected

- Exit 0; default (non `-r`) script scan loop reads `user.koolTargetDir` and matches when
  either `path` or `user.koolTargetDir` equals `targetDir`, so an existing kool-opened
  session is found even when iTerm `path` has moved.

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
		t.Fatalf("exit=%d stderr=%s", resp.ExitCode, resp.Stderr)
	}
	s := resp.CapturedScript
	if s == "" {
		t.Fatal("expected captured script")
	}
	if !strings.Contains(s, `variable named "user.koolTargetDir"`) {
		t.Fatalf("smart-open scan must read user.koolTargetDir: %q", s)
	}
	scanStart := strings.Index(s, "repeat with aWindow in windows")
	if scanStart < 0 {
		t.Fatal("missing window scan loop")
	}
	scanEnd := strings.Index(s[scanStart:], "if matchingWindow is not missing value then")
	if scanEnd < 0 {
		t.Fatal("missing match branch opener after scan")
	}
	scanLoop := s[scanStart : scanStart+scanEnd]
	if !strings.Contains(scanLoop, "user.koolTargetDir") {
		t.Fatalf("scan loop must check user.koolTargetDir: %q", scanLoop)
	}
	if !strings.Contains(scanLoop, "or koolTargetDir is targetDir") {
		t.Fatalf("scan loop must match on koolTargetDir: %q", scanLoop)
	}
}
```