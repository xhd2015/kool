# Scenario

**Feature**: input dir with files/ + env.yaml builds successfully

```
kool sandbox build -o sandbox.bin -i in
  # in/files/hello.txt, in/env.yaml TOKEN=x
  -> exit 0; binary exists size>0; stdout mentions files/env counts
```

## Steps

1. Write fixture input dir with one file and one env entry.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_, err := writeInputDir(t, req.WorkingDir, "in",
		map[string]string{"hello.txt": "hello from sandbox\n"},
		map[string]string{"TOKEN": "fixture-token"},
		"",
	)
	if err != nil {
		return err
	}
	req.Input = "in"
	req.InputSet = true
	return nil
}
```
