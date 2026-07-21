# Scenario

**Feature**: serve derives tunnel name kool-lb-<leftmost-label>

```
kool cloudflare serve --domain a.example.com --url http://127.0.0.1:9
  -> StartSession Domain=a.example.com LocalURL=http://127.0.0.1:9 TunnelName=kool-lb-a
  -> WaitSignal; Stop; exit 0; stdout contains https://a.example.com
```

## Steps

1. Domain + URL; no --tunnel.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.DomainSet = true
	req.Domain = "a.example.com"
	req.URLSet = true
	req.URL = "http://127.0.0.1:9"
	req.TunnelSet = false
	return nil
}
```
