## Expected

- Exit 0; captured script targets current session and does not create tab.

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
		t.Fatalf("exit=%d stderr=%s", resp.ExitCode, resp.Stderr)
	}
	s := resp.CapturedScript
	if !strings.Contains(s, "current session of current tab of current window") {
		t.Fatalf("missing current session: %q", s)
	}
	if strings.Contains(s, "create tab with default profile") {
		t.Fatal("reuse must not create tab")
	}
	if !strings.Contains(s, `write text ("cd " & quoted form of targetDir)`) {
		t.Fatalf("missing cd: %q", s)
	}
}
```