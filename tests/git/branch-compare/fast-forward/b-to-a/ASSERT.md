## Expected
- The command exits with code 0
- Stdout contains "main is newer(a +1 commit -> main)"
- Stdout contains "to fast forward, on a:"
- Stdout contains "   git merge --ff-only  main"

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
	if !strings.Contains(resp.Stdout, "main is newer(a +1 commit -> main)") {
		t.Fatalf("expected stdout to contain 'main is newer(a +1 commit -> main)', got:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "to fast forward, on a:") {
		t.Fatalf("expected stdout to contain 'to fast forward, on a:', got:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "git merge --ff-only  main") {
		t.Fatalf("expected stdout to contain 'git merge --ff-only  main', got:\n%s", resp.Stdout)
	}
}
```
