## Expected

- Exit 0
- JSON object has session_id, app, contents
- No ANSI

```go
import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if strings.Contains(resp.Stdout, "\x1b") {
		t.Fatalf("json has ANSI: %q", resp.Stdout)
	}
	var obj map[string]string
	if e := json.Unmarshal([]byte(resp.Stdout), &obj); e != nil {
		t.Fatalf("json: %v stdout=%q", e, resp.Stdout)
	}
	if obj["session_id"] != "B95E6BAC-3104-43D2-ABAE-86FC02A669A2" {
		t.Fatalf("session_id=%q", obj["session_id"])
	}
	if obj["app"] != "~/Applications/iTerm.app" {
		t.Fatalf("app=%q", obj["app"])
	}
	if obj["contents"] != "pane" {
		t.Fatalf("contents=%q", obj["contents"])
	}
}
```
