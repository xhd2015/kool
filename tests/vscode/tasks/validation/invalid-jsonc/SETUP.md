# Scenario

**Feature**: invalid tasks.json content is rejected

```
broken JSON (not even JSONC) -> list Error parse/invalid
```

## Steps

1. Write intentionally broken file; list via --dir.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
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
