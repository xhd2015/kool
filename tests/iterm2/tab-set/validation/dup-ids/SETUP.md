# Scenario

**Feature**: duplicate tab ids are rejected

```
two tabs with id "a" -> show dupid -> Error
```

## Steps

1. Write dupid.json with duplicate ids.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeConfigFile(t, req.ConfigDir, "dupid", `{
  "version": 1,
  "tabs": [
    {"id": "a", "name": "one", "command": "echo one"},
    {"id": "a", "name": "two", "command": "echo two"}
  ]
}
`)
	req.SetName = "dupid"
	req.Subcommand = "show"
	return nil
}
```
