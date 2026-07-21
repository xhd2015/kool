## Expected

- Non-zero exit.
- Stderr mentions url.
- StartSession not called.

## Errors

- Missing --url.

## Exit Code

- non-zero

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for missing --url; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stderr) == "" {
		t.Fatal("expected stderr validation message")
	}
	if resp.StartCalled {
		t.Fatal("StartSession must not run when --url missing")
	}
	low := strings.ToLower(resp.Stderr)
	if !strings.Contains(low, "url") {
		t.Fatalf("stderr should mention url; got %q", resp.Stderr)
	}
}
```
