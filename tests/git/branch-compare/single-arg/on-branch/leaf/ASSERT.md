## Expected

- The command exits with code 0
- Stdout matches two-arg `kool git compare-branch main feature` fast-forward output
- Stdout contains "main is newer(feature +1 commit -> main)"
- Stdout contains "to fast forward, on feature:"
- Stdout contains "git merge --ff-only  main"

## Exit Code

0

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
	if !strings.Contains(resp.Stdout, "main is newer(feature +1 commit -> main)") {
		t.Fatalf("expected stdout to contain 'main is newer(feature +1 commit -> main)', got:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "to fast forward, on feature:") {
		t.Fatalf("expected stdout to contain 'to fast forward, on feature:', got:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "git merge --ff-only  main") {
		t.Fatalf("expected stdout to contain 'git merge --ff-only  main', got:\n%s", resp.Stdout)
	}
}
```