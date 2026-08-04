## Expected

- Ambiguity is fatal, reports every 0-based candidate, includes `--index`, and never focuses.

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
 if err != nil { t.Fatal(err) }; if resp.ExitCode == 0 || resp.DiscoverCalls != 1 || len(resp.Focused) != 0 { t.Fatalf("exit=%d discovery=%d focused=%v", resp.ExitCode, resp.DiscoverCalls, resp.Focused) }
 assert.Output(t, resp.Stderr, `---
version: 3
---
Error: multiple iTerm2 sessions found for: .*

  \[0\] window 1 \("credit-pricing"\) tab 2 session w0t2p0
  \[1\] window 3 \("worktrees"\) tab 1 session w2t1p0

Specify one with:
  kool iterm2 focus .* --index <0\|1>
`)
}
```
