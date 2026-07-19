## Expected

- Exit 0.
- Help text mentions `--tab`, `--save`, and `--force` (ad-hoc/save cycle docs).
- Prefer also `--window-name` and wording that --save does not run iTerm / needs --tab.

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
		t.Fatalf("exit=%d out:\n%s", resp.ExitCode, combinedOut(resp))
	}
	out := strings.ToLower(resp.Stdout + resp.Stderr)
	for _, want := range []string{"--tab", "--save", "--force"} {
		if !strings.Contains(out, want) {
			t.Fatalf("help missing %q; out:\n%s", want, resp.Stdout)
		}
	}
}
```
