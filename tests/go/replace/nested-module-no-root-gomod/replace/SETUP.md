# Scenario

**Feature**: library Replace() for nested module without root go.mod

## Steps

1. Build dot-pkgs-like fixture
2. Set operation to library replace

```go
func Setup(t *testing.T, req *Request) error {
	req.Operation = "replace"
	return nil
}
```