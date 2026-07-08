# git tag-next IncrementTag

`IncrementTag` is the pure function behind `kool git tag-next`. Given a version tag string,
it increments the trailing numeric segment and returns the next tag (e.g. `v0.2.0` →
`v0.2.1`). No git repository or CLI subprocess is involved in these unit cases.

# DSN (Domain Specific Notion)

The **caller** supplies a version tag string. **IncrementTag** scans the tag from the
right, isolates the trailing numeric segment, increments it by one, and returns the
recomposed tag. If the trailing segment is missing, non-numeric, or cannot be incremented,
IncrementTag returns an error.

## Version

0.0.2

## Decision Tree

The top-level split is **patch category**: tags whose trailing patch is zero (the bug
surface) versus non-zero patches (regression guard). Each leaf pins one concrete input tag
and expected next tag.

```
tag-next-increment
├── zero-patch/          # trailing patch is 0 — primary bug cases
│   ├── v0-2-0/          # v0.2.0 → v0.2.1
│   └── v0-0-0/          # v0.0.0 → v0.0.1
└── regression/          # existing increment behavior must not regress
    ├── v0-0-87/         # v0.0.87 → v0.0.88
    ├── v0-2-1/          # v0.2.1 → v0.2.2
    └── v0-2-10/         # v0.2.10 → v0.2.11
```

## Test Index

| # | Leaf | Input | Expected |
|---|------|-------|----------|
| 1 | `zero-patch/v0-2-0` | `v0.2.0` | `v0.2.1` |
| 2 | `zero-patch/v0-0-0` | `v0.0.0` | `v0.0.1` |
| 3 | `regression/v0-0-87` | `v0.0.87` | `v0.0.88` |
| 4 | `regression/v0-2-1` | `v0.2.1` | `v0.2.2` |
| 5 | `regression/v0-2-10` | `v0.2.10` | `v0.2.11` |

## How to Run

```sh
doctest vet ./tests/git/tag-next-increment
doctest test ./tests/git/tag-next-increment
```

```go
import (
	"testing"

	git_tag_next "github.com/xhd2015/kool/tools/git/git_tag_next"
)

type Request struct {
	Tag string // version tag to increment (set by leaf Setup)
}

type Response struct {
	NextTag string
	Err     error
}

func Run(t *testing.T, req *Request) (*Response, error) {
	nextTag, err := git_tag_next.IncrementTag(req.Tag)
	return &Response{NextTag: nextTag, Err: err}, nil
}
```