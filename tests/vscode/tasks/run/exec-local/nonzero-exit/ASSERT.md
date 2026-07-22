## Expected

- Non-zero exit (failed shell child `false`).
- Must not claim success (exit 0) for a failing task.

## Errors

- Propagate child failure via process exit code.

## Exit Code

- ≠ 0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("failing local task must exit ≠ 0; stdout=%s stderr=%s", resp.Stdout, resp.Stderr)
	}
}
```
