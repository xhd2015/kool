# compare-branch

`kool git compare-branch a b` compares two git references and reports their relationship.

## How to Run

```sh
doctest test -v ./
```

## Test Cases

| # | Test | Description |
|---|------|-------------|
| 1 | identical/leaf | Both refs resolve to the same commit |
| 2 | fast-forward/a-to-b/leaf | a is ancestor of b — a can fast-forward to b |
| 3 | fast-forward/b-to-a/leaf | b is ancestor of a — b can fast-forward to a |
| 4 | divergent/leaf | Both have unique commits with file differences |
| 5 | divergent/no-file-diff/leaf | Divergent commits but zero file differences |
| 6 | errors/invalid-ref-a/leaf | a does not resolve to any commit |
| 7 | errors/invalid-ref-b/leaf | b does not resolve to any commit |
| 8 | with-C-flag/valid-dir/leaf | -C flag points to a valid git repo |
| 9 | with-C-flag/nonexistent/leaf | -C flag points to a non-existent directory |
| 10 | not-a-git-repo/leaf | Directory is not a git repository |
