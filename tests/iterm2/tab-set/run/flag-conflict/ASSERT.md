## Expected

- Non-zero exit (prefer 1).
- Error message mentions conflict / mutually exclusive / both flags.

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
		t.Fatalf("expected flag conflict error; stdout=%s", resp.Stdout)
	}
	out := strings.ToLower(resp.Stderr + resp.Stdout)
	// Must be tab-set flag handling, not open-dir "unrecognized flag: --dry-run".
	if strings.Contains(out, "unrecognized flag") || strings.Contains(out, "unrecognized argument") {
		t.Fatalf("tab-set run not routed (open-dir fallback); out:\n%s", resp.Stderr+resp.Stdout)
	}
	if !strings.Contains(out, "new-window") && !strings.Contains(out, "no-new-window") &&
		!strings.Contains(out, "conflict") && !strings.Contains(out, "mutually") &&
		!strings.Contains(out, "cannot") && !strings.Contains(out, "incompatible") {
		t.Fatalf("expected conflict message; out:\n%s", resp.Stderr+resp.Stdout)
	}
}
```

