## Expected
- The command exits with code 0
- Stdout contains "main and a has 0 files difference"
- Stdout contains "their most recent base is" with a commit hash
- Stdout contains "main has 1 unique commit"
- Stdout contains "a has 1 unique commit"
- Stdout contains "They need to be merged"

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
	if !strings.Contains(resp.Stdout, "main and a has 0 files difference") {
		t.Fatalf("expected stdout to contain 'main and a has 0 files difference', got:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "their most recent base is") {
		t.Fatalf("expected stdout to contain 'their most recent base is', got:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "main has 1 unique commit") {
		t.Fatalf("expected stdout to contain 'main has 1 unique commit', got:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "a has 1 unique commit") {
		t.Fatalf("expected stdout to contain 'a has 1 unique commit', got:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "They need to be merged") {
		t.Fatalf("expected stdout to contain 'They need to be merged', got:\n%s", resp.Stdout)
	}
}
```
