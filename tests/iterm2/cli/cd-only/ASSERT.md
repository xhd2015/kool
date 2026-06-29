## Expected

- Exit 0; captured script has cd line and no grok/codex follow-ups.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%s", resp.ExitCode, resp.Stderr)
	}
	if resp.CapturedScript == "" {
		t.Fatal("expected captured script")
	}
	if !strings.Contains(resp.CapturedScript, `write text ("cd " & quoted form of targetDir)`) {
		t.Fatalf("missing cd: %q", resp.CapturedScript)
	}
	if strings.Contains(resp.CapturedScript, `write text "grok"`) {
		t.Fatal("unexpected follow-up in cd-only run")
	}
}
```