# Scenario

**Feature**: invalid tasks.json content is rejected

```
broken JSON (not even JSONC) -> list Error parse/invalid
```

## Steps

1. Write intentionally broken file; list via --dir.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Not valid JSON or JSONC: unclosed brace / garbage
	writeTasksJSON(t, req.WorkingDir, `{
  "version": "2.0.0",
  "tasks": [
    { "label": "Broken"
  // missing closing braces on purpose
`)
	req.Dir = req.WorkingDir
	return nil
}
```
