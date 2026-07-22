## Expected

- Exit 0.
- Real-run output includes absolute workspace path, basename, and `tok123`.
- No unexpanded `${workspaceFolder}` or `${env:KOOL_TASKS_TEST_TOKEN}`.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("local expand-vars exit=%d stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := combinedOut(resp)
	base := filepath.Base(req.WorkingDir)
	if !strings.Contains(out, req.WorkingDir) {
		t.Fatalf("expected workspace path %q in output; out:\n%s", req.WorkingDir, out)
	}
	if !strings.Contains(out, base) {
		t.Fatalf("expected basename %q in output; out:\n%s", base, out)
	}
	if !strings.Contains(out, "tok123") {
		t.Fatalf("expected env token tok123; out:\n%s", out)
	}
	if strings.Contains(out, "${workspaceFolder}") || strings.Contains(out, "${env:KOOL_TASKS_TEST_TOKEN}") {
		t.Fatalf("unexpanded vars remain; out:\n%s", out)
	}
}
```
