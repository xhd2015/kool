## Expected

- Exit 1; no captured script.

## Exit Code

- 1

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected failure, stderr=%s", resp.Stderr)
	}
	if resp.CapturedScript != "" {
		t.Fatal("osascript should not run for missing path")
	}
}
```