## Expected

- Exit 0.
- Plan targets Compile (go build).

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
		t.Fatalf("exact match dry-run exit=%d stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout
	if !strings.Contains(out, "Compile") && !strings.Contains(out, "go build") {
		t.Fatalf("exact match plan missing Compile; out:\n%s", out)
	}
}
```
