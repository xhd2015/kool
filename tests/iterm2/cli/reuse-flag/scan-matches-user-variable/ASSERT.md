## Expected

- Exit 0; captured script scan loop reads `user.koolTargetDir` in addition to `path`
  and branches to the focus path when either equals `targetDir`.

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
		t.Fatalf("scan must read user.koolTargetDir: %q", s)
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
	if !strings.Contains(scanLoop, "targetDir") {
		t.Fatalf("scan loop must compare against targetDir: %q", scanLoop)
	}
}
```