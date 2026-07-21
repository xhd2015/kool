# Scenario

**Feature**: packed file is visible at its sandbox-relative path from cwd

```
kool sandbox build -o sandbox.bin --file hello.txt=hello.txt
KOOL_SANDBOX_ROOT=PARENT ./sandbox.bin -- sh -c 'cat hello.txt'
  -> exit 0; stdout == packed content
```

## Steps

1. Write local `hello.txt`; pack via `--file`.
2. Guest: `sh -c 'cat hello.txt'`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if _, err := writeLocalFile(t, req.WorkingDir, "hello.txt", "hello from sandbox\n"); err != nil {
		return err
	}
	req.ExtraFiles = []string{"hello.txt=hello.txt"}
	req.SealedArgs = []string{"sh", "-c", "cat hello.txt"}
	return nil
}
```
