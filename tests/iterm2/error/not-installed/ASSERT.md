## Expected

- Exit 1; stderr mentions install or iTerm2.

## Exit Code

- 1

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected failure, stderr=%s", resp.Stderr)
	}
	combined := resp.Stderr + resp.Stdout
	if !strings.Contains(strings.ToLower(combined), "iterm") {
		t.Fatalf("stderr=%q", resp.Stderr)
	}
}
```