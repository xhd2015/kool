# Scenario

**Feature**: --file and --env alone produce a sealed binary

```
kool sandbox build -o sandbox.bin --file local.txt=app/local.txt --env FLAG_ENV=from-flag
  -> exit 0; binary exists size>0
```

## Steps

1. Write local source file; pass --file and --env; no -i.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if _, err := writeLocalFile(t, req.WorkingDir, "local.txt", "flag-only content\n"); err != nil {
		return err
	}
	req.ExtraFiles = []string{"local.txt=app/local.txt"}
	req.ExtraEnv = []string{"FLAG_ENV=from-flag"}
	return nil
}
```
