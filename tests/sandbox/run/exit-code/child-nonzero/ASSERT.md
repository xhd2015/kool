## Expected

- Build succeeds; sealed binary executed.
- Sealed exit code is exactly 42 (child status propagated).

## Exit Code

- sealed run: 42

```go
import "testing"

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
	if resp.RunExitCode != 42 {
		t.Fatalf("sealed exit=%d want 42; stdout=%q stderr=%q",
			resp.RunExitCode, resp.RunStdout, resp.RunStderr)
	}
}
```
