## Expected

- `OpenDirOptions` returns a **non-nil error** (IPC could not handle open).
- OS opener hook was **not** called.
- Stderr does **not** contain the human URI fallback hint.

## Errors

- Exec hook or stderr hint indicates default fallback path ran under `--ipc-only`.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected harness error: %v", err)
	}
	if resp.ValidateErr == "" {
		t.Fatal("expected error when ipc-only and IPC unreachable")
	}
	if resp.ExecCalled {
		t.Fatal("ipc-only failure must not invoke OS opener")
	}
	if resp.StderrHint {
		t.Fatalf("ipc-only must not print URI fallback hint: %s", resp.Stderr)
	}
}
```