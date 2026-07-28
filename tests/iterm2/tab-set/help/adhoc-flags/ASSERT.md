## Expected

- Exit 0.
- Help text mentions `--tab`, `--save`, and `--force` (ad-hoc/save cycle docs).
- Help mentions `no_submit` as an ad-hoc/tab prop key (stage command without Enter).
- Prefer also `--window-name` and wording that --save does not run iTerm / needs --tab.

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
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d out:\n%s", resp.ExitCode, combinedOut(resp))
	}
	out := strings.ToLower(resp.Stdout + resp.Stderr)
	for _, want := range []string{"--tab", "--save", "--force"} {
		if !strings.Contains(out, want) {
			t.Fatalf("help missing %q; out:\n%s", want, resp.Stdout)
		}
	}
	if !strings.Contains(out, "no_submit") {
		t.Fatalf("help should document no_submit prop; out:\n%s", resp.Stdout)
	}
}
```
