## Expected

- Exit 0.
- Stdout is JSON (starts with `[` or `{`) containing labels Compile / Serve / Build All.
- No ANSI escape sequences in stdout.

## Exit Code

- 0

```go
import (
	"encoding/json"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("--json list exit=%d stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := strings.TrimSpace(resp.Stdout)
	if out == "" {
		t.Fatal("expected JSON on stdout")
	}
	if strings.Contains(resp.Stdout, "\x1b[") {
		t.Fatalf("--json must not include ANSI; out:\n%q", resp.Stdout)
	}
	if !(strings.HasPrefix(out, "[") || strings.HasPrefix(out, "{")) {
		t.Fatalf("expected JSON document; out:\n%s", out)
	}
	var v any
	if err := json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("stdout not valid JSON: %v\n%s", err, out)
	}
	for _, label := range []string{"Compile", "Serve", "Build All"} {
		if !strings.Contains(out, label) {
			t.Fatalf("JSON missing label %q; out:\n%s", label, out)
		}
	}
}
```
