## Expected

- Exit 0; binary exists size > 0.
- Exact `SecretProbe` string must not appear in binary bytes (plaintext leak check).
- Prefer scanning file bytes (portable); optional `strings` tool if present.

## Exit Code

- 0

```go
import (
	"bytes"
	"os"
	"os/exec"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d want 0; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if !resp.OutputExists || resp.OutputSize <= 0 {
		t.Fatalf("expected sealed binary; exists=%v size=%d", resp.OutputExists, resp.OutputSize)
	}
	if req.SecretProbe == "" {
		t.Fatal("SecretProbe must be set by Setup")
	}
	raw, readErr := os.ReadFile(resp.OutputPath)
	if readErr != nil {
		t.Fatalf("read binary: %v", readErr)
	}
	if bytes.Contains(raw, []byte(req.SecretProbe)) {
		t.Fatalf("plaintext secret %q found in sealed binary (AES seal missing?)", req.SecretProbe)
	}
	// Optional strings(1) double-check when available.
	if path, lookErr := exec.LookPath("strings"); lookErr == nil {
		out, runErr := exec.Command(path, resp.OutputPath).CombinedOutput()
		if runErr != nil {
			t.Logf("strings failed (ignored): %v", runErr)
			return
		}
		if bytes.Contains(out, []byte(req.SecretProbe)) {
			t.Fatalf("strings(1) found secret %q in binary", req.SecretProbe)
		}
	}
}
```
