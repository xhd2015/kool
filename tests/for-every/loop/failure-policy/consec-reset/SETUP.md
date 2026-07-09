# Scenario

**Feature**: successful run resets consecutive failure counter

```
# odd runs fail, even succeed; max-failure 2 must not trip on non-consecutive fails
kool for-every --max-failure 2 --max-runs 5 10ms sh -c '…counter…'
  -> five run-N lines (F S F S F); exit non-zero from last fail
```

## Steps

1. Counter file in WorkingDir; fail when n%2==1; max-failure 2; max-runs 5.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.MaxFailure = intPtr(2)
	req.AllowFailure = false
	req.MaxRuns = intPtr(5)
	// Counter in cwd (WorkingDir). Odd runs exit 1; even exit 0.
	// Prints run-<n> so assertions can count iterations.
	req.Command = "sh"
	req.Args = []string{"-c",
		`n=0; if [ -f .for_every_runs ]; then n=$(cat .for_every_runs); fi; n=$((n+1)); echo "$n" > .for_every_runs; echo "run-$n"; if [ $((n % 2)) -eq 1 ]; then exit 1; fi; exit 0`,
	}
	return nil
}
```
