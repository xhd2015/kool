# Scenario

**Feature**: unlisted extension blocks open

```
# code present but extension missing from --list-extensions
EnsureExtensionListed -> error (extension id + install hint)
```

## Steps
1. Create valid directory.
2. Install fake `code` that lists other extensions only.
3. Run CLI.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	dir := initValidDir(t, req.WorkingDir, "target")
	req.DirPath = dir
	installNoExtensionPrecheck(t, req)
	return nil
}
```