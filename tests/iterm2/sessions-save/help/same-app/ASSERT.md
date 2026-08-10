## Expected

- Exit 0
- Help documents `--same-app`
- Help mentions prefer / default toward `~/Applications` (or home Applications)
  when multiple installs exist

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
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	out := resp.Stdout
	if !strings.Contains(out, "--same-app") {
		t.Fatalf("restore help missing --same-app:\n%s", out)
	}
	low := strings.ToLower(out)
	// Prefer-home default when multiple installs (wording flexible).
	preferOK := (strings.Contains(low, "~/applications") ||
		strings.Contains(low, "home") ||
		strings.Contains(out, "~/Applications")) &&
		(strings.Contains(low, "prefer") ||
			strings.Contains(low, "default") ||
			strings.Contains(low, "multiple") ||
			strings.Contains(low, "both"))
	// Also accept explicit same-app explanation that recorded app is recreated.
	sameOK := strings.Contains(low, "recorded") ||
		strings.Contains(low, "same app") ||
		strings.Contains(low, "per-window") ||
		strings.Contains(low, "each window")
	if !preferOK && !sameOK {
		t.Fatalf("restore help should document prefer-home when multi-install and/or --same-app recorded-app behavior:\n%s", out)
	}
	// Stronger prefer-home signal when both phrases appear is ideal; require at
	// least one of prefer-home wording OR clear same-app + home path mention.
	if !strings.Contains(out, "~/Applications") && !strings.Contains(low, "applications/iterm") {
		// Still require some path or home-install cue alongside --same-app.
		if !strings.Contains(low, "home") && !strings.Contains(low, "install") {
			t.Fatalf("restore help should mention home install / Applications path:\n%s", out)
		}
	}
}
```
