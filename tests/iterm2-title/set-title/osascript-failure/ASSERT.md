## Expected

- Exit non-zero (1) when osascript fails after set-title is accepted.
- Fake osascript must have been invoked (script capture written) — proves routing
  reached the title path, not open-dir.
- Must not claim success with `title changed:`.

## Errors

- Osascript / AppleScript failure propagated to CLI.

## Exit Code

- 1

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit on osascript failure\nstdout=%q\nstderr=%q", resp.Stdout, resp.Stderr)
	}
	if !resp.ScriptWritten || resp.CapturedScript == "" {
		t.Fatalf("expected osascript to run (script capture); stderr=%q", resp.Stderr)
	}
	if strings.Contains(resp.Stdout, "title changed:") {
		t.Fatalf("should not report success when osascript fails: %q", resp.Stdout)
	}
}
```
