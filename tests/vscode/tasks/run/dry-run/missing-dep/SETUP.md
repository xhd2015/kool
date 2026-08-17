# Scenario

**Feature**: dry-run errors when dependsOn label is missing

```
Root dependsOn "No Such Task" -> run Root --dry-run -> Error
```

## Steps

1. missingDepJSONC; Query=`Root`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeTasksJSON(t, req.WorkingDir, missingDepJSONC)
	req.Dir = req.WorkingDir
	req.Query = "Root"
	return nil
}
```
