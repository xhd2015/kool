# Scenario

**Feature**: nested packed files are readable via relative paths because cwd is root

```
kool sandbox build -o sandbox.bin --file nested.txt=app/data/nested.txt
KOOL_SANDBOX_ROOT=PARENT ./sandbox.bin -- sh -c 'cat app/data/nested.txt'
  -> exit 0; stdout == packed content
```

## Steps

1. Pack a nested relative path under `app/data/`.
2. Guest cats that relative path from materialize cwd.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if _, err := writeLocalFile(t, req.WorkingDir, "nested.txt", "nested-content\n"); err != nil {
		return err
	}
	req.ExtraFiles = []string{"nested.txt=app/data/nested.txt"}
	req.SealedArgs = []string{"sh", "-c", "cat app/data/nested.txt"}
	return nil
}
```
