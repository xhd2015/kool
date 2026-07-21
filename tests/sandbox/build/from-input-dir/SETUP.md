# Scenario

**Feature**: build from config directory (`-i`)

```
user -> kool sandbox build -o OUT -i DIR
  -> merge files/ + env.yaml (+ optional meta.yaml) → sealed OUT
```

## Steps

1. Leaves write fixture dirs under WorkingDir and set Input/InputSet.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Output = "sandbox.bin"
	req.OutputSet = true
	req.BuildTwice = false
	req.AfterBuildInspect = false
	return nil
}
```
