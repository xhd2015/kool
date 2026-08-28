## Expected

- Exit 0; SendText to tab 1 UUID.

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
	if len(resp.SendCalls) != 1 || resp.SendCalls[0].SessionID != fixtureTab1ID {
		t.Fatalf("SendCalls=%#v", resp.SendCalls)
	}
	assert.Output(t, resp.Stdout, "sent to session "+fixtureTab1ID+"\n")
}
```
