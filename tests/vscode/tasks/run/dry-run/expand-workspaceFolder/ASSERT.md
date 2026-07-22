## Expected

- Exit 0.
- Plan contains expanded workspace absolute path (or at least basename).
- Plan contains env token `tok123`.
- Unexpanded `${workspaceFolder}` / `${env:KOOL_TASKS_TEST_TOKEN}` must not remain.

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
		t.Fatalf("expand dry-run exit=%d stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout
	base := filepath.Base(req.WorkingDir)
	if !strings.Contains(out, req.WorkingDir) && !strings.Contains(out, base) {
		t.Fatalf("plan should expand workspace path or basename %q; out:\n%s", base, out)
	}
	if !strings.Contains(out, "tok123") {
		t.Fatalf("plan should expand ${env:KOOL_TASKS_TEST_TOKEN}=tok123; out:\n%s", out)
	}
	if strings.Contains(out, "${workspaceFolder}") ||
		strings.Contains(out, "${workspaceFolderBasename}") ||
		strings.Contains(out, "${env:KOOL_TASKS_TEST_TOKEN}") {
		t.Fatalf("plan still has unexpanded vars; out:\n%s", out)
	}
}
```
