## Expected

- Exit 0.
- Output mentions `update` and key flags: `--tab-id`, `--rm`, `--no-submit`.
- Prefer also `--submit`, `--create`, `--window-name`, or `--force` if present.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d

	if err != nil {
		t.Fatal(err)
	}
	out := combinedOut(resp)
	if resp.ExitCode != 0 {
		t.Fatalf("update help exit=%d out:\n%s", resp.ExitCode, out)
	}
	lower := strings.ToLower(out)
	for _, want := range []string{"update", "--tab-id", "--rm", "--no-submit"} {
		if !strings.Contains(lower, want) {
			t.Fatalf("update help missing %q; out:\n%s", want, out)
		}
	}
}
```
