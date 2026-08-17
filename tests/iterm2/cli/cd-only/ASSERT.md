## Expected

- Exit 0; captured script has cd line and no grok/codex follow-ups.

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
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%s", resp.ExitCode, resp.Stderr)
	}
	if resp.CapturedScript == "" {
		t.Fatal("expected captured script")
	}
	if !strings.Contains(resp.CapturedScript, writeTextCd()) {
		t.Fatalf("missing cd: %q", resp.CapturedScript)
	}
	if strings.Contains(resp.CapturedScript, writeTextCmd("grok")) {
		t.Fatal("unexpected follow-up in cd-only run")
	}
}
```