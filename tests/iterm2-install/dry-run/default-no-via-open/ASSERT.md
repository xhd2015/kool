## Expected

- Exit 0.
- Stdout contains dry-run banner and version `3.6.11`.
- Stdout does **not** claim via-open mode: no `via-open`, `user-open`,
  `user open`, or `InstallViaUserOpen` substrings (case-insensitive).
- No zip written under `--download-dir`.

## Side Effects

- No zip file at planned path under DownloadDir.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assert.Output(t, resp.Stdout, `<contains>
dry-run
3.6.11
</contains>
`)
	low := strings.ToLower(resp.Stdout)
	for _, banned := range []string{
		"via-open",
		"user-open",
		"user open",
		"installviauseropen",
	} {
		if strings.Contains(low, banned) {
			t.Fatalf("default dry-run must not claim via-open mode (%q); got:\n%s", banned, resp.Stdout)
		}
	}
	assertNoZip(t, resp)
}
```
