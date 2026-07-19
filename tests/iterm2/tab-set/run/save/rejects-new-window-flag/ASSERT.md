## Expected

- Non-zero exit.
- Message mentions new-window / -n and/or save incompatibility.
- No scratch.json written.

## Exit Code

- ≠ 0

```go
import (
	"os"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	out := combinedOut(resp)
	if resp.ExitCode == 0 {
		t.Fatalf("--save with -n must error; out:\n%s", out)
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "unrecognized flag") || strings.Contains(lower, "unknown flag") {
		t.Fatalf("flags not accepted; out:\n%s", out)
	}
	if !strings.Contains(lower, "new-window") && !strings.Contains(lower, "-n") &&
		!strings.Contains(lower, "save") {
		t.Fatalf("expected -n/--save conflict message; out:\n%s", out)
	}
	if _, e := os.Stat(configPath(req.ConfigDir, "scratch")); e == nil {
		t.Fatalf("must not write scratch.json on flag error")
	}
}
```
