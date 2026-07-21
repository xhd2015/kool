# Scenario

**Feature**: two builds of the same input produce different sealed binaries

```
kool sandbox build -o sandbox.bin -i in
kool sandbox build -o sandbox.bin.second -i in
  -> both exit 0; SHA-256 digests differ (fresh key/ciphertext per build)
```

## Steps

1. Fixture pack; BuildTwice=true so Run performs two builds.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	_, err := writeInputDir(t, req.WorkingDir, "in",
		map[string]string{"a.txt": "same-input\n"},
		map[string]string{"E": "1"},
		"",
	)
	if err != nil {
		return err
	}
	req.Input = "in"
	req.InputSet = true
	req.BuildTwice = true
	return nil
}
```
