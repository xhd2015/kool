## Expected

- Exit 0
- FileJSON contains `"space": 0` (or `"space":0`)
- Stderr has a warning mentioning space (resolve fail / not a user Desktop)

## Exit Code

- 0

```go
import (
	"regexp"
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
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if resp.FileJSON == "" {
		t.Fatal("expected checkpoint written")
	}
	// Always emit space 0 on resolve failure.
	space0 := regexp.MustCompile(`"space"\s*:\s*0`)
	if !space0.MatchString(resp.FileJSON) {
		t.Fatalf("expected \"space\": 0 on resolve fail:\n%s", resp.FileJSON)
	}
	errOut := strings.ToLower(resp.Stderr)
	if !strings.Contains(errOut, "space") {
		t.Fatalf("expected stderr warning mentioning space; stderr=%q", resp.Stderr)
	}
	// Soft: warn path should look like a warning, not a hard Error: abort.
	if strings.Contains(errOut, "error:") && resp.ExitCode != 0 {
		t.Fatalf("resolve fail is soft (warn + space 0), not fatal: %q", resp.Stderr)
	}
}
```
