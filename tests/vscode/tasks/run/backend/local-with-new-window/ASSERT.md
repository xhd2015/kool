## Expected

- Non-zero exit.
- Message ties window flag to backend (local cannot use `-n` / new-window), or
  rejects the combination.

## Errors

- Documented policy: window flags with local **error** (not ignore).

## Exit Code

- ≠ 0

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
		t.Fatalf("local + -n must error (not ignore); stdout=%s", resp.Stdout)
	}
	out := strings.ToLower(combinedOut(resp))
	ok := strings.Contains(out, "local") ||
		strings.Contains(out, "iterm") ||
		strings.Contains(out, "new-window") ||
		strings.Contains(out, "new window") ||
		strings.Contains(out, "backend") ||
		strings.Contains(out, "not supported") ||
		strings.Contains(out, "invalid")
	if !ok {
		t.Fatalf("expected local+window flag error; out:\n%s", combinedOut(resp))
	}
}
```
