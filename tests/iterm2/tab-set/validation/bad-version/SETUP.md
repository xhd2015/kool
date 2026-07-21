# Scenario

**Feature**: version != 1 is rejected

```
{"version":2,...} -> show badver -> Error
```

## Steps

1. Write badver.json with version 2; show badver.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markValidationTree()
	markTabSetValidationTree()
	markTabSetTree()
	markRootTree()
	writeConfigFile(t, req.ConfigDir, "badver", `{
  "version": 2,
  "tabs": [{"id": "a", "name": "a", "command": "echo a"}]
}
`)
	req.SetName = "badver"
	req.Subcommand = "show"
	return nil
}
```
