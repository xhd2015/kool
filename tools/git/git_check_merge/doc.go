// Package git_check_merge provides functionality to check if git references
// are merged into the current HEAD. It can check multiple references at once
// and outputs whether each is merged or how many commits it is ahead of HEAD.
// It sets exit code 1 if any reference is not merged.
package git_check_merge
