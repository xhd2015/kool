# Scenario

**Feature**: flag --file and --env override same path/key from -i

```
# dir has files/shared.txt (dir content) and env SHARED=from-dir
# flags: --file flag.txt=shared.txt --env SHARED=from-flag
kool sandbox build -o sandbox.bin -i in --file flag.txt=shared.txt --env SHARED=from-flag
  -> build exit 0; inspect shows path shared.txt and env key SHARED
  -> inspect must not print secret env values (from-flag / from-dir)
```

## Steps

1. Write dir with conflicting path and env key.
2. Write flag local file with different content.
3. AfterBuildInspect inherited true.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	_, err := writeInputDir(t, req.WorkingDir, "in",
		map[string]string{"shared.txt": "content-from-dir\n"},
		map[string]string{"SHARED": "from-dir"},
		"",
	)
	if err != nil {
		return err
	}
	if _, err := writeLocalFile(t, req.WorkingDir, "flag.txt", "content-from-flag\n"); err != nil {
		return err
	}
	req.Input = "in"
	req.InputSet = true
	req.ExtraFiles = []string{"flag.txt=shared.txt"}
	req.ExtraEnv = []string{"SHARED=from-flag"}
	return nil
}
```
