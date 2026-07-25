## Expected

- Exit 0.
- Stdout does **not** contain fixture agent session id
  `019fabcdef-1234-5678-9abc-def012345678`.
- Stdout does **not** contain agent tree connectors `├──` / `└──` from
  FormatTree (process hierarchy lines alone must not introduce these glyphs
  when enrich is off).
- Snapshot still succeeds (contains `busy` or window markers).

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
	if strings.Contains(out, fixtureGrokSessionID) {
		t.Fatalf("--no-enrich must not print agent session id; stdout:\n%s", out)
	}
	for _, glyph := range []string{"├──", "└──"} {
		if strings.Contains(out, glyph) {
			t.Fatalf("--no-enrich must not print tree connector %q:\n%s", glyph, out)
		}
	}
	// Still a successful snapshot of the fixture hierarchy.
	if !strings.Contains(out, "busy") && !strings.Contains(out, "W1") && !strings.Contains(out, "Win-") {
		t.Fatalf("expected basic snapshot output even with --no-enrich:\n%s", out)
	}
}
```
