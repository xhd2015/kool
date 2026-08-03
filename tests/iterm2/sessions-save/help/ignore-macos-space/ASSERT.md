## Expected

- Exit 0
- Combined save + restore help mentions `--ignore-macos-space`

## Errors

- None

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
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout
	if !strings.Contains(out, "--ignore-macos-space") {
		t.Fatalf("save/restore help missing --ignore-macos-space:\n%s", out)
	}
	// Both subcommands should appear so the combined help is not just one side.
	if !strings.Contains(out, "save") || !strings.Contains(out, "restore") {
		t.Fatalf("combined help should mention save and restore:\n%s", out)
	}
}
```
