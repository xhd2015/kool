## Expected
- The command exits with code 0
- Stdout contains "main and main are identical"

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error running kool: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstderr: %s", resp.ExitCode, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "main and main are identical") {
		t.Fatalf("expected stdout to contain 'main and main are identical', got:\n%s", resp.Stdout)
	}
}
```
