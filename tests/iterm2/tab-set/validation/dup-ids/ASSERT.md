## Expected

- Non-zero exit.
- Message mentions duplicate / id.

## Exit Code

- ≠ 0

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
		t.Fatalf("expected error for duplicate ids; stdout=%s", resp.Stdout)
	}
	out := strings.ToLower(resp.Stderr + resp.Stdout)
	if strings.Contains(out, "unrecognized argument") || strings.Contains(out, "unrecognized flag") {
		t.Fatalf("tab-set not routed (open-dir fallback); out:\n%s", resp.Stderr+resp.Stdout)
	}
	if !strings.Contains(out, "duplicate") && !strings.Contains(out, "invalid") {
		t.Fatalf("expected duplicate-id error; out:\n%s", resp.Stderr+resp.Stdout)
	}
}
```

