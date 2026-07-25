## Expected

- No API error; Snapshot non-nil.
- `ListWindowsCalls == 1`.
- `ListTabsCalls == 2` (windows 1 and 2 each once).
- `ListTabsByIndex[1] >= 1` and `ListTabsByIndex[2] >= 1`.
- Summary: Windows=2, Sessions=2; Idle+Busy counts consistent (idle≥1, busy≥1
  with fixture TTYs).

## Errors

- None.

## Exit Code

- 0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.APIError != "" {
		t.Fatalf("CaptureSnapshot error: %s", resp.APIError)
	}
	if resp.Snapshot == nil {
		t.Fatal("Snapshot is nil")
	}
	if resp.ListWindowsCalls != 1 {
		t.Fatalf("ListWindowsCalls=%d want 1 (Capture must use phased ListWindows)", resp.ListWindowsCalls)
	}
	if resp.ListTabsCalls != 2 {
		t.Fatalf("ListTabsCalls=%d want 2 (one ListTabsAndSessions per window)", resp.ListTabsCalls)
	}
	if resp.ListTabsByIndex[1] < 1 || resp.ListTabsByIndex[2] < 1 {
		t.Fatalf("ListTabsByIndex=%v want keys 1 and 2", resp.ListTabsByIndex)
	}
	sum := resp.Snapshot.Summary
	if sum.Windows != 2 || sum.Sessions != 2 {
		t.Fatalf("summary windows=%d sessions=%d want 2/2", sum.Windows, sum.Sessions)
	}
	if sum.Idle < 1 || sum.Busy < 1 {
		t.Fatalf("summary idle=%d busy=%d want both ≥1 (fixture enrich)", sum.Idle, sum.Busy)
	}
}
```
