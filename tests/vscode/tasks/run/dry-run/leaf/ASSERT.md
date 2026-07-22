## Expected

- Exit 0.
- Plan mentions `Compile` and command `go build` (or bin/app path pieces).
- Prefer dry-run / plan wording; no process execution side effects required.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("dry-run leaf exit=%d stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout
	if !strings.Contains(out, "Compile") && !strings.Contains(out, "go build") {
		t.Fatalf("plan should mention Compile / go build; out:\n%s", out)
	}
}
```
