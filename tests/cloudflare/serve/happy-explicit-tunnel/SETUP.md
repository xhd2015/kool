# Scenario

**Feature**: serve --tunnel overrides derived name

```
kool cloudflare serve --domain a.example.com --url http://127.0.0.1:9 --tunnel my-tun
  -> StartSession TunnelName=my-tun; exit 0
```

## Steps

1. Domain + URL + explicit tunnel.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.DomainSet = true
	req.Domain = "a.example.com"
	req.URLSet = true
	req.URL = "http://127.0.0.1:9"
	req.TunnelSet = true
	req.Tunnel = "my-tun"
	return nil
}
```
