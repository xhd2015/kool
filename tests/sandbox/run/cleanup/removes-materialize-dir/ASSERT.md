## Expected

- Build succeeds.
- Sealed run exit 0 (successful guest so cleanup path is exercised).
- `MaterializeEmpty` is true: no remaining children under `SandboxRootParent`.

## Side Effects

- Session materialize directory under `KOOL_SANDBOX_ROOT` is removed best-effort
  on exit (success path).

## Exit Code

- sealed run: 0

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
		t.Fatalf("build exit=%d want 0; stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if !resp.RunExecuted {
		t.Fatal("expected sealed binary run")
	}
	if resp.RunExitCode != 0 {
		// Stub "run not implemented" fails here — RED until runtime lands.
		t.Fatalf("sealed exit=%d want 0 for cleanup path; stdout=%q stderr=%q",
			resp.RunExitCode, resp.RunStdout, resp.RunStderr)
	}
	if resp.SandboxRootParent == "" {
		t.Fatal("expected SandboxRootParent")
	}
	if !resp.MaterializeEmpty {
		t.Fatalf("expected KOOL_SANDBOX_ROOT empty after successful run; parent=%q remaining=%v",
			resp.SandboxRootParent, resp.MaterializeRemaining)
	}
	if len(resp.MaterializeRemaining) != 0 {
		t.Fatalf("unexpected remaining materialize entries: %s", strings.Join(resp.MaterializeRemaining, ", "))
	}
}
```
