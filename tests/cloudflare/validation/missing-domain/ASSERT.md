## Expected

- Non-zero exit.
- Stderr mentions domain.
- StartSession not called.

## Errors

- Missing --domain.

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
		t.Fatalf("expected non-zero for missing --domain; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stderr) == "" {
		t.Fatal("expected stderr validation message")
	}
	if resp.StartCalled {
		t.Fatal("StartSession must not run when --domain missing")
	}
	low := strings.ToLower(resp.Stderr)
	if !strings.Contains(low, "domain") {
		t.Fatalf("stderr should mention domain; got %q", resp.Stderr)
	}
}
```
