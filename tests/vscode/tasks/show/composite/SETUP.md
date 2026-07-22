# Scenario

**Feature**: show composite task prints dependsOn list

```
show "Build All" -> label, composite type, dependsOn Compile+Serve
```

## Steps

1. Multi-task fixture; Query exact `Build All`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeMultiTaskFixture(t, req.WorkingDir)
	req.Dir = req.WorkingDir
	req.Query = "Build All"
	return nil
}
```
