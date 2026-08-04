## Expected

- The sole candidate is focused and stdout has the selected window/tab with a final newline.

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
 if err != nil { t.Fatal(err) }; if resp.ExitCode != 0 || resp.DiscoverCalls != 1 || len(resp.Focused) != 1 || resp.Focused[0] != "w2t1p0" { t.Fatalf("exit=%d discovery=%d focused=%v stderr=%q", resp.ExitCode, resp.DiscoverCalls, resp.Focused, resp.Stderr) }
 assert.Output(t, resp.Stdout, `---
version: 3
---
focused: window 3, tab 1
`)
}
```
