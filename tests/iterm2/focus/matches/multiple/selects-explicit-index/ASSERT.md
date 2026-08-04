## Expected

- Candidate 1, and no other candidate, is selected.

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
 if err != nil { t.Fatal(err) }; if resp.ExitCode != 0 || len(resp.Focused) != 1 || resp.Focused[0] != "w2t1p0" { t.Fatalf("exit=%d focused=%v stderr=%q", resp.ExitCode, resp.Focused, resp.Stderr) }
 assert.Output(t, resp.Stdout, `---
version: 3
---
focused: window 3, tab 1
`)
}
```
