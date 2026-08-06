## Expected

- Exit 0.
- Stdout is a JSON array of objects with fields `coverage` (number, percent) and
  `package` (string), sorted coverage ascending then package name:
  internal/run @ 0, cli @ 100.
- Stdout ends with `\n`.
- No markdown table headers.

## Exit Code

- 0

```go
import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

type covRow struct {
	Coverage float64 `json:"coverage"`
	Package  string  `json:"package"`
}

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d want 0; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("json stdout must end with newline; got %q", resp.Stdout)
	}
	if strings.Contains(resp.Stdout, "| Coverage |") {
		t.Fatalf("json mode must not emit markdown table; got:\n%s", resp.Stdout)
	}
	var rows []covRow
	if err := json.Unmarshal([]byte(strings.TrimSpace(resp.Stdout)), &rows); err != nil {
		t.Fatalf("stdout is not JSON array: %v\n%s", err, resp.Stdout)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d: %+v", len(rows), rows)
	}
	if rows[0].Package != "internal/run" || rows[0].Coverage != 0 {
		t.Fatalf("row0 want internal/run @ 0, got %+v", rows[0])
	}
	if rows[1].Package != "cli" || rows[1].Coverage != 100 {
		t.Fatalf("row1 want cli @ 100, got %+v", rows[1])
	}
}
```
