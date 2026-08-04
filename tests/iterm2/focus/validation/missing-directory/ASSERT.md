## Expected

- The fatal error has the `Error:` prefix and discovery is not called.

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
 if err != nil { t.Fatal(err) }; if resp.ExitCode == 0 || resp.DiscoverCalls != 0 { t.Fatalf("exit=%d discovery=%d stderr=%q", resp.ExitCode, resp.DiscoverCalls, resp.Stderr) }
 assert.Output(t, resp.Stderr, `---
version: 3
---
Error: missing directory argument
.*focus --help.*
`)
}
```
