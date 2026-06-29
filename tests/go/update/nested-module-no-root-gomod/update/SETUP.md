# Scenario

**Feature**: library Update() for nested module without root go.mod

## Steps

1. Build dot-pkgs-like fixture
2. Set operation to library update

```go
func Setup(t *testing.T, req *Request) error {
	req.Operation = "update"
	return nil
}
```