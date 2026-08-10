## Expected

- Exit 0
- Help lightly documents multi-app behavior (any of: dual installs, both
  Application paths, `app` field, preferred restore app)

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
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	out := strings.ToLower(resp.Stdout)
	// Light contract — multi-app specific (not mere "restored" / generic wording).
	ok := strings.Contains(out, "applications/iterm") ||
		strings.Contains(out, "~/applications") ||
		(strings.Contains(out, "dual") && (strings.Contains(out, "install") || strings.Contains(out, "app"))) ||
		(strings.Contains(out, "preferred") && strings.Contains(out, "app")) ||
		(strings.Contains(out, "multi") && strings.Contains(out, "app"))
	if !ok {
		t.Fatalf("save help should mention multi-app / dual installs / app paths / preferred restore app:\n%s", resp.Stdout)
	}
}
```