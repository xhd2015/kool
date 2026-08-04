## Expected

- A file target is fatal and is never discovered or focused.

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
 if err != nil { t.Fatal(err) }; if resp.ExitCode == 0 || resp.DiscoverCalls != 0 || len(resp.Focused) != 0 { t.Fatalf("exit=%d discovery=%d focused=%v", resp.ExitCode, resp.DiscoverCalls, resp.Focused) }
 assert.Output(t, resp.Stderr, `---
version: 3
---
Error: not a directory: .*
`)
}
```
