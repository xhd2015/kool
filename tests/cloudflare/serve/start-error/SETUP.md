# Scenario

**Feature**: StartSession error surfaces and fails serve

```
kool cloudflare serve --domain a.example.com --url http://127.0.0.1:9
  + inject StartSession returns error "simulated start failure"
  -> non-zero; stderr contains failure text; Stop not required
```

## Steps

1. AllowStart with StartSessionErr set.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.DomainSet = true
	req.Domain = "a.example.com"
	req.URLSet = true
	req.URL = "http://127.0.0.1:9"
	req.StartSessionErr = "simulated start failure"
	return nil
}
```
