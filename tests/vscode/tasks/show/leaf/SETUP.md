# Scenario

**Feature**: show shell leaf prints command and options

```
show Compile -> type shell, command go build, cwd workspaceFolder
```

## Steps

1. Multi-task fixture; Query=`Compile`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeMultiTaskFixture(t, req.WorkingDir)
	req.Dir = req.WorkingDir
	req.Query = "Compile"
	return nil
}
```
