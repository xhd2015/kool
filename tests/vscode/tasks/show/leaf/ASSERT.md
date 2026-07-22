## Expected

- Exit 0.
- Output includes `Compile`, shell type (or command text `go build`), workspace mention preferred.

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
		t.Fatalf("show leaf exit=%d stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout
	if !strings.Contains(out, "Compile") {
		t.Fatalf("show missing Compile; out:\n%s", out)
	}
	lower := strings.ToLower(out)
	if !strings.Contains(out, "go build") && !strings.Contains(lower, "shell") {
		t.Fatalf("show leaf should include command or type shell; out:\n%s", out)
	}
}
```
