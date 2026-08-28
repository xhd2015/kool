## Expected

- Exit 0; SendText to resolved UUID with Focus; stdout uses user ref.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%s", resp.ExitCode, resp.Stderr)
	}
	if len(resp.SendCalls) != 1 {
		t.Fatalf("SendCalls=%#v", resp.SendCalls)
	}
	c := resp.SendCalls[0]
	if c.SessionID != fixtureTab1ID || c.Text != "echo hi" || !c.Opts.Focus {
		t.Fatalf("call=%#v", c)
	}
	assert.Output(t, resp.Stdout, "sent to session 11111111\n")
}
```
