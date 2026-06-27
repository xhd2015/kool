## Expected
- Non-zero exit code.
- Stderr contains usage hint for `<path>`.

## Exit Code
- 1

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error running kool: %v", err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for missing arg, got 0\nstderr: %s", resp.Stderr)
	}
	combined := resp.Stderr + resp.Stdout
	if !strings.Contains(combined, "open-git-repo") && !strings.Contains(combined, "path") {
		t.Fatalf("expected usage mentioning open-git-repo/path, got:\n%s", combined)
	}
}
```