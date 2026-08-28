## Expected

- Exit 0; one SendText to tab 2 UUID; stdout `sent to session <tab2>`.

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
	if resp.SendCalls[0].SessionID != fixtureTab2ID || resp.SendCalls[0].Text != "from-next" {
		t.Fatalf("call=%#v", resp.SendCalls[0])
	}
	assert.Output(t, resp.Stdout, "sent to session "+fixtureTab2ID+"\n")
}
```
