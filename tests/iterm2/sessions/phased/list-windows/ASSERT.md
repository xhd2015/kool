## Expected

- No API error.
- `ListWindows` returns exactly two windows.
- Window 1: Index=1, Name=`Win-A`.
- Window 2: Index=2, Name=`Win-B`.
- Tabs may be empty on headers-only returns (tabs filled by ListTabsAndSessions).

## Errors

- None.

## Exit Code

- 0 (APIError empty)

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
		t.Fatalf("ListWindows error: %s", resp.APIError)
	}
	if len(resp.Windows) != 2 {
		t.Fatalf("windows=%d want 2: %+v", len(resp.Windows), resp.Windows)
	}
	if resp.Windows[0].Index != 1 || resp.Windows[0].Name != "Win-A" {
		t.Fatalf("window[0]=%+v want Index=1 Name=Win-A", resp.Windows[0])
	}
	if resp.Windows[1].Index != 2 || resp.Windows[1].Name != "Win-B" {
		t.Fatalf("window[1]=%+v want Index=2 Name=Win-B", resp.Windows[1])
	}
}
```
