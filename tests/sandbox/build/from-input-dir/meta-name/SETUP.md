# Scenario

**Feature**: meta.yaml name appears in build summary stdout

```
kool sandbox build -o sandbox.bin -i in
  # in/meta.yaml name: p1-demo-sandbox; files/readme.txt
  -> exit 0; stdout mentions p1-demo-sandbox
```

## Steps

1. Write meta.yaml with a distinctive name plus one file so pack is non-empty.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	_, err := writeInputDir(t, req.WorkingDir, "in",
		map[string]string{"readme.txt": "named pack\n"},
		nil,
		"name: p1-demo-sandbox\ncomment: doctest meta name\n",
	)
	if err != nil {
		return err
	}
	req.Input = "in"
	req.InputSet = true
	return nil
}
```
