# Scenario

**Feature**: unique secret in packed file must not appear as plaintext in binary

```
kool sandbox build -o sandbox.bin -i in
  # in/files/secret.txt contains SECRET_PROBE_kool_sandbox_p1_9f3a2c1b7e
  -> exit 0; `strings` on binary must not contain that exact string
```

## Steps

1. Pack a highly unique secret string into a file under -i.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	const secret = "SECRET_PROBE_kool_sandbox_p1_9f3a2c1b7e"
	req.SecretProbe = secret
	_, err := writeInputDir(t, req.WorkingDir, "in",
		map[string]string{"secret.txt": secret + "\n"},
		map[string]string{"OK": "1"},
		"",
	)
	if err != nil {
		return err
	}
	req.Input = "in"
	req.InputSet = true
	return nil
}
```
