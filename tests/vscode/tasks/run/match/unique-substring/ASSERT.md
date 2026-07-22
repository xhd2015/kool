## Expected

- Exit 0.
- Plan mentions Serve / bin/app; not Compile as primary if exclusive.

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
		t.Fatalf("unique substring dry-run exit=%d stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout
	if !strings.Contains(out, "Serve") && !strings.Contains(out, "bin/app") {
		t.Fatalf("unique substring should resolve Serve; out:\n%s", out)
	}
}
```
