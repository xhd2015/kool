# Scenario

**Feature**: --env without KEY=VALUE form is invalid

```
kool sandbox build -o sandbox.bin --env NOTVALID
  -> non-zero; stderr mentions env or =
```

## Steps

1. Pass malformed --env (no `=`).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Output = "sandbox.bin"
	req.OutputSet = true
	req.ExtraEnv = []string{"NOTVALID"}
	return nil
}
```
