## Expected Output

```
. some.com/root
sub-dir some.com/root/sub
```

## Expected

- Exit code 0, no stderr.
- stdout contains exactly two lines (order may be walk order; assert the SET):
  - `. some.com/root`
  - `sub-dir some.com/root/sub`
- Each line is `<dir><space><path>` (single space, `.` for root, plain slash-relative for
  the sub-dir with no `./` prefix).

```go
import (
	"reflect"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error running test: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout: %s\nstderr: %s", resp.ExitCode, resp.Stdout, resp.Stderr)
	}

	lines := stdoutLines(resp.Stdout)
	want := []string{
		". some.com/root",
		"sub-dir some.com/root/sub",
	}
	if !sameSet(lines, want) {
		t.Fatalf("stdout lines = %v, want set %v\nfull stdout:\n%s", lines, want, resp.Stdout)
	}
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	set := make(map[string]int)
	for _, s := range a {
		set[s]++
	}
	for _, s := range b {
		set[s]--
		if set[s] < 0 {
			return false
		}
	}
	for _, v := range set {
		if v != 0 {
			return false
		}
	}
	return true
}

var _ = reflect.DeepEqual
```
