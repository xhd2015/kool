package release

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
)

const CredentialsHelp = `
To obtain a GitHub token:
  1. Go to https://github.com/settings/tokens
  2. Click "Generate new token (classic)"
  3. Give it a name (e.g. "<app>-release-upload")
  4. Select the "repo" scope
  5. Click "Generate token" and copy the value`

type Credentials struct {
	Token string `json:"token"`
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
}

func LoadCredentials(path string) (*Credentials, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%s not found.\n\nCreate %s with the following format:\n  {\"token\": \"ghp_...\", \"owner\": \"<github-owner>\", \"repo\": \"<repo-name>\"}\n%s", path, path, CredentialsHelp)
		}
		return nil, fmt.Errorf("failed to read %s: %v", path, err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("%s is empty.\n\nCreate it with the following format:\n  {\"token\": \"ghp_...\", \"owner\": \"<github-owner>\", \"repo\": \"<repo-name>\"}\n%s", path, CredentialsHelp)
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("%s is not valid JSON: %v\n\nExpected format:\n  {\"token\": \"ghp_...\", \"owner\": \"<github-owner>\", \"repo\": \"<repo-name>\"}\n%s", path, err, CredentialsHelp)
	}

	if creds.Token == "" {
		return nil, fmt.Errorf("%s exists but token is empty.\n\nSet your GitHub token in the file:\n  {\"token\": \"ghp_...\", \"owner\": \"<github-owner>\", \"repo\": \"<repo-name>\"}\n%s", path, CredentialsHelp)
	}
	if creds.Owner == "" {
		return nil, fmt.Errorf("%s exists but owner is empty.\n\nSet the GitHub owner in the file:\n  {\"token\": \"ghp_...\", \"owner\": \"<github-owner>\", \"repo\": \"<repo-name>\"}\n%s", path, CredentialsHelp)
	}
	if creds.Repo == "" {
		return nil, fmt.Errorf("%s exists but repo is empty.\n\nSet the repo name in the file:\n  {\"token\": \"ghp_...\", \"owner\": \"<github-owner>\", \"repo\": \"<repo-name>\"}\n%s", path, CredentialsHelp)
	}

	return &creds, nil
}

func GetTag() (string, error) {
	out, err := cmd.Output("git", "status", "--porcelain")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(string(out)) != "" {
		return "", fmt.Errorf("git status is not clean, ensure everything is committed. check with 'git status'")
	}

	tag, err := cmd.Output("git", "describe", "--tags", "HEAD")
	if err != nil {
		return "", err
	}
	tag = strings.TrimSpace(string(tag))
	if tag == "" {
		return "", fmt.Errorf("no tag found, ensure you are on a tagged commit")
	}
	if !strings.HasPrefix(tag, "v") {
		return "", fmt.Errorf("tag %s is not a valid version, must start with 'v'", tag)
	}
	return tag, nil
}
