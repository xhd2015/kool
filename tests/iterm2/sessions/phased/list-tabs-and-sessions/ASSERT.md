## Expected

- No API error.
- At least one tab for window 1.
- First tab Index=1, Name=`Tab-A`.
- First session: ID prefix `AAAAAAAA`, TTY `/dev/ttys001`, Name `idle-sess`.

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
	if resp.APIError != "" {
		t.Fatalf("ListTabsAndSessions error: %s", resp.APIError)
	}
	if len(resp.Tabs) < 1 {
		t.Fatalf("tabs empty: %+v", resp.Tabs)
	}
	tab := resp.Tabs[0]
	if tab.Index != 1 || tab.Name != "Tab-A" {
		t.Fatalf("tab=%+v want Index=1 Name=Tab-A", tab)
	}
	if len(tab.Sessions) < 1 {
		t.Fatalf("no sessions on tab: %+v", tab)
	}
	s := tab.Sessions[0]
	if !strings.HasPrefix(s.ID, "AAAAAAAA") {
		t.Fatalf("session id=%q want AAAAAAAA…", s.ID)
	}
	if s.TTY != "/dev/ttys001" {
		t.Fatalf("tty=%q want /dev/ttys001", s.TTY)
	}
	if s.Name != "idle-sess" {
		t.Fatalf("name=%q want idle-sess", s.Name)
	}
}
```
