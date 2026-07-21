# Scenario

**Feature**: packed environment variables are visible to the guest process

```
kool sandbox build -o sandbox.bin --env FOO=bar
KOOL_SANDBOX_ROOT=PARENT ./sandbox.bin -- sh -c 'printf %s "$FOO"'
  -> exit 0; stdout == bar
```

## Steps

1. Pack `--env FOO=bar` (no files required).
2. Guest prints `$FOO`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ExtraEnv = []string{"FOO=bar"}
	req.SealedArgs = []string{"sh", "-c", `printf %s "$FOO"`}
	return nil
}
```
