## Expected

- Sealed run exit 0.
- Guest file content is load L (not H seed, not P primary).
- `$HOME` equals `$SANDBOX_ROOT` (session under KOOL_SANDBOX_ROOT parent), not fake real home.

## Exit Code

- sealed run: 0

```go
import (
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
		t.Fatalf("build exit=%d want 0; stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if !resp.RunExecuted {
		t.Fatal("expected sealed binary run")
	}
	if resp.RunExitCode != 0 {
		t.Fatalf("sealed exit=%d want 0; stdout=%q stderr=%q", resp.RunExitCode, resp.RunStdout, resp.RunStderr)
	}
	out := resp.RunStdout
	if strings.Contains(out, "content-H-from-real-home") {
		t.Fatalf("guest must not see home seed H for shared.txt; stdout=%q", out)
	}
	if strings.Contains(out, "content-P-from-primary") {
		t.Fatalf("guest must not see primary P for shared.txt; stdout=%q", out)
	}
	if !strings.Contains(out, "content-L-from-load") {
		t.Fatalf("guest should see load L; stdout=%q", out)
	}
	// Parse HOME= and SANDBOX= lines.
	var home, sandbox string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "HOME=") {
			home = strings.TrimPrefix(line, "HOME=")
		}
		if strings.HasPrefix(line, "SANDBOX=") {
			sandbox = strings.TrimPrefix(line, "SANDBOX=")
		}
	}
	if home == "" || sandbox == "" {
		t.Fatalf("want HOME= and SANDBOX= lines; stdout=%q", out)
	}
	if home != sandbox {
		t.Fatalf("HOME %q != SANDBOX_ROOT %q", home, sandbox)
	}
	if !filepath.IsAbs(home) {
		t.Fatalf("HOME should be absolute; got %q", home)
	}
	parent := resp.SandboxRootParent
	rel, err := filepath.Rel(parent, home)
	if err != nil {
		t.Fatalf("Rel: %v", err)
	}
	if rel == "." || strings.HasPrefix(rel, "..") {
		t.Fatalf("HOME %q should be session child of %q (rel=%q)", home, parent, rel)
	}
	if resp.SealedHome != "" && home == resp.SealedHome {
		t.Fatalf("guest HOME must not remain fake real home %q", resp.SealedHome)
	}
}
```
