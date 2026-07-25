## Expected

- Exit 0.
- Stdout contains fixture agent session id
  `019fabcdef-1234-5678-9abc-def012345678`.
- Stdout contains `grok`.
- Stdout does **not** contain tree connectors `├──` or `└──`.

## Errors

- None.

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
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout
	if !strings.Contains(out, fixtureGrokSessionID) {
		t.Fatalf("--no-tree should still show agent session id:\n%s", out)
	}
	if !strings.Contains(out, "grok") {
		t.Fatalf("stdout missing grok:\n%s", out)
	}
	for _, glyph := range []string{"├──", "└──"} {
		if strings.Contains(out, glyph) {
			t.Fatalf("--no-tree must not print connector %q:\n%s", glyph, out)
		}
	}
}
```
