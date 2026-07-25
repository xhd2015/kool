## Expected

- Exit 0.
- Stdout is valid JSON unmarshallable into an object with `windows` array length 2.
- Contains fixture session id `AAAAAAAA-0000-0000-0000-000000000001`.
- No ANSI escape sequences.
- Progressive probe: `SawW1BeforeLastListTabs` must be **false** (JSON is
  buffered; CLI `W1` markers are not used — if ListTabs probe runs, stdout
  must not look like progressive CLI). Prefer: entire stdout parses as one JSON
  value (not concatenated objects).

## Errors

- None.

## Exit Code

- 0

```go
import (
	"encoding/json"
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
	if strings.Contains(out, "\x1b[") {
		t.Fatal("ANSI in JSON output")
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("stdout is not a single JSON document: %v\n%s", err, out)
	}
	wins, ok := doc["windows"].([]any)
	if !ok || len(wins) != 2 {
		t.Fatalf("windows=%v want array len 2", doc["windows"])
	}
	if !strings.Contains(out, "AAAAAAAA-0000-0000-0000-000000000001") {
		t.Fatalf("missing fixture session id:\n%s", out)
	}
	// Buffering: must not have emitted CLI W1 mid-collection.
	if resp.SawW1BeforeLastListTabs {
		t.Fatal("JSON path must not progressive-emit CLI W1 during ListTabs")
	}
}
```
