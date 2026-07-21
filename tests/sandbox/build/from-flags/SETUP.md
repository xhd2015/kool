# Scenario

**Feature**: build from CLI flags only (no `-i`)

```
user -> kool sandbox build -o OUT --file L=R --env K=V
  -> sealed OUT without config directory
```

## Steps

1. Leaves set ExtraFiles/ExtraEnv only; Input remains unset.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Output = "sandbox.bin"
	req.OutputSet = true
	req.Input = ""
	req.InputSet = false
	req.BuildTwice = false
	req.AfterBuildInspect = false
	return nil
}
```
