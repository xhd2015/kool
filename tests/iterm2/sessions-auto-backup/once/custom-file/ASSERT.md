## Expected

- Exit 0
- File written under WorkingDir (temp), not DOCTEST_CASE / source tree
- Stdout Saved and mentions custom-auto.json (or full path)

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
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "Saved") {
		t.Fatalf("stdout missing Saved:\n%s", resp.Stdout)
	}
	if resp.ResolvedPath == "" {
		t.Fatal("ResolvedPath empty")
	}
	// Must not land in the source leaf (git-tracked tree).
	caseDir, _ := filepath.Abs(d.DOCTEST_CASE)
	resolved, _ := filepath.Abs(resp.ResolvedPath)
	if caseDir != "" && (resolved == caseDir || strings.HasPrefix(resolved, caseDir+string(filepath.Separator))) {
		t.Fatalf("custom file must not be written under source leaf %s; got %s", caseDir, resolved)
	}
	if req.WorkingDir != "" {
		wd, _ := filepath.Abs(req.WorkingDir)
		if !strings.HasPrefix(resolved, wd+string(filepath.Separator)) && resolved != wd {
			t.Fatalf("expected file under WorkingDir %s; got %s", wd, resolved)
		}
	}
	if _, e := os.Stat(resp.ResolvedPath); e != nil {
		t.Fatalf("custom file missing at %s: %v", resp.ResolvedPath, e)
	}
	if resp.Doc == nil || resp.Doc.Version != 1 {
		t.Fatalf("expected versioned doc at custom path; doc=%+v", resp.Doc)
	}
	// Path should appear in output (full path or basename).
	if !strings.Contains(resp.Stdout, "custom-auto.json") &&
		!strings.Contains(resp.Stdout, resp.ResolvedPath) {
		t.Fatalf("stdout should mention custom file path:\n%s", resp.Stdout)
	}
}
```
