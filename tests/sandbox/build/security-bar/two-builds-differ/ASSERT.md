## Expected

- First build exit 0; both output binaries exist with size > 0.
- Binaries are **not** byte-identical (`BinariesIdentical` false).

## Exit Code

- 0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d want 0; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if !resp.OutputExists || resp.OutputSize <= 0 {
		t.Fatalf("first binary missing; exists=%v size=%d path=%q", resp.OutputExists, resp.OutputSize, resp.OutputPath)
	}
	if !resp.SecondOutputExists || resp.SecondOutputSize <= 0 {
		t.Fatalf("second binary missing; exists=%v size=%d path=%q", resp.SecondOutputExists, resp.SecondOutputSize, resp.SecondOutputPath)
	}
	if resp.BinariesIdentical {
		t.Fatal("two builds of the same input must not produce byte-identical sealed binaries")
	}
}
```
