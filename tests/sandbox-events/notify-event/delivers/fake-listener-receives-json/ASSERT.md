## Expected

- Exit 0.
- At least one payload delivered to the mock listener.
- Parsed JSON has `v` == 1, `type` == `devbox.updated`, `path` equal to
  request EventPath, and non-empty `ts` parseable as RFC3339 / RFC3339Nano.

## Exit Code

- 0

```go
import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("notify-event exit=%d want 0; stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if resp.DeliveredCount < 1 || len(resp.DeliveredRaw) < 1 {
		t.Fatalf("expected delivered JSON payload; count=%d raw=%v stdout=%q stderr=%q",
			resp.DeliveredCount, resp.DeliveredRaw, resp.Stdout, resp.Stderr)
	}
	m := resp.FirstDelivered
	if m == nil {
		t.Fatalf("delivered payload not valid JSON: %q", resp.DeliveredRaw[0])
	}
	if fmt.Sprint(m["v"]) != "1" {
		t.Fatalf("JSON v want 1; got %#v full=%v", m["v"], m)
	}
	typ, _ := m["type"].(string)
	if typ != "devbox.updated" {
		t.Fatalf("JSON type=%q want devbox.updated; full=%v", typ, m)
	}
	path, _ := m["path"].(string)
	if path != req.EventPath {
		t.Fatalf("JSON path=%q want %q; full=%v", path, req.EventPath, m)
	}
	ts, _ := m["ts"].(string)
	if strings.TrimSpace(ts) == "" {
		t.Fatalf("JSON ts empty; full=%v", m)
	}
	if _, e1 := time.Parse(time.RFC3339, ts); e1 != nil {
		if _, e2 := time.Parse(time.RFC3339Nano, ts); e2 != nil {
			t.Fatalf("JSON ts not RFC3339: %q (%v)", ts, e1)
		}
	}
}
```
