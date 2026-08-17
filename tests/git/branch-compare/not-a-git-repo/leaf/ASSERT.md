## Errors
- The command exits with non-zero exit code
- Stderr or stdout indicates this is not a git repository

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatalf("unexpected error running kool: %v", err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit code for non-git directory, got 0\nstdout: %s", resp.Stdout)
	}
	if !strings.Contains(resp.Stderr, "not a git") && !strings.Contains(resp.Stderr, "fatal:") {
		t.Fatalf("expected stderr to indicate not a git repository, got:\n%s", resp.Stderr)
	}
}
```
