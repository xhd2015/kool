package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/kool/pkgs/github"
	"github.com/xhd2015/kool/script/lib"
	"github.com/xhd2015/xgo/support/cmd"
)

const credentialsHelp = `
To obtain a GitHub token:
  1. Go to https://github.com/settings/tokens
  2. Click "Generate new token (classic)"
  3. Give it a name (e.g. "kool-release-upload")
  4. Select the "repo" scope
  5. Click "Generate token" and copy the value`

type credentials struct {
	Token string `json:"token"`
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
}

func main() {
	err := handle()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func handle() error {
	dryRun := flag.Bool("dry-run", false, "print what would be done without actually building or uploading")
	flag.Parse()

	if *dryRun {
		return handleDryRun()
	}

	creds, err := loadCredentials()
	if err != nil {
		return err
	}

	result, err := lib.BuildRelease(lib.DefaultSpecs)
	if err != nil {
		return err
	}

	client := github.NewReleaseClient(creds.Token, creds.Owner, creds.Repo)

	release, err := client.GetOrCreateRelease(result.Tag)
	if err != nil {
		return fmt.Errorf("failed to get or create release for tag %s: %v", result.Tag, err)
	}

	for _, file := range result.Files {
		if err := client.UploadReleaseAsset(release.ID, file); err != nil {
			return fmt.Errorf("failed to upload %s: %v", file, err)
		}
		fmt.Printf("Uploaded %s\n", file)
	}
	return nil
}

func handleDryRun() error {
	checkGitClean()

	tag, tagErr := getTag()
	if tagErr != nil {
		fmt.Fprintf(os.Stderr, "[dry-run] warning: %v\n", tagErr)
		tag = "(unknown)"
	}

	creds, credsErr := loadCredentials()
	if credsErr != nil {
		fmt.Fprintf(os.Stderr, "[dry-run] warning: %v\n", credsErr)
		creds = &credentials{Owner: "?", Repo: "?"}
	}

	fmt.Printf("[dry-run] tag: %s\n", tag)
	for _, spec := range lib.DefaultSpecs {
		fmt.Printf("[dry-run] would build: kool-%s-%s-%s\n", tag, spec.OS, spec.Arch)
	}
	fmt.Printf("[dry-run] would upload to %s/%s release (creates if 404)\n", creds.Owner, creds.Repo)
	return nil
}

func checkGitClean() {
	out, err := cmd.Output("git", "status", "--porcelain")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[dry-run] warning: failed to check git status: %v\n", err)
		return
	}
	if strings.TrimSpace(string(out)) != "" {
		fmt.Fprintf(os.Stderr, "[dry-run] warning: git status is not clean, ensure everything is committed. check with 'git status'\n")
	}
}

func getTag() (string, error) {
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

func loadCredentials() (*credentials, error) {
	data, err := os.ReadFile(".upload-credentials.json")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf(".upload-credentials.json not found.\n\nCreate .upload-credentials.json with the following format:\n  {\"token\": \"ghp_...\", \"owner\": \"<github-owner>\", \"repo\": \"<repo-name>\"}\n%s", credentialsHelp)
		}
		return nil, fmt.Errorf("failed to read .upload-credentials.json: %v", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf(".upload-credentials.json is empty.\n\nCreate it with the following format:\n  {\"token\": \"ghp_...\", \"owner\": \"<github-owner>\", \"repo\": \"<repo-name>\"}\n%s", credentialsHelp)
	}

	var creds credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf(".upload-credentials.json is not valid JSON: %v\n\nExpected format:\n  {\"token\": \"ghp_...\", \"owner\": \"<github-owner>\", \"repo\": \"<repo-name>\"}\n%s", err, credentialsHelp)
	}

	if creds.Token == "" {
		return nil, fmt.Errorf(".upload-credentials.json exists but token is empty.\n\nSet your GitHub token in the file:\n  {\"token\": \"ghp_...\", \"owner\": \"<github-owner>\", \"repo\": \"<repo-name>\"}\n%s", credentialsHelp)
	}
	if creds.Owner == "" {
		return nil, fmt.Errorf(".upload-credentials.json exists but owner is empty.\n\nSet the GitHub owner in the file:\n  {\"token\": \"ghp_...\", \"owner\": \"<github-owner>\", \"repo\": \"<repo-name>\"}\n%s", credentialsHelp)
	}
	if creds.Repo == "" {
		return nil, fmt.Errorf(".upload-credentials.json exists but repo is empty.\n\nSet the repo name in the file:\n  {\"token\": \"ghp_...\", \"owner\": \"<github-owner>\", \"repo\": \"<repo-name>\"}\n%s", credentialsHelp)
	}

	return &creds, nil
}
