## Expected

- Focus help succeeds and documents its directory argument and index option.

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil { t.Fatal(err) }; if resp.ExitCode != 0 { t.Fatalf("exit=%d stderr=%q", resp.ExitCode, resp.Stderr) }
    assert.Output(t, resp.Stdout, `---
version: 3
---
Usage: kool iterm2 focus <dir> \[--index N\]
.*--index.*
`)
}
```
