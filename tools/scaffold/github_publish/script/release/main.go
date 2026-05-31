package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/kool/pkgs/github"
	"github.com/xhd2015/kool/pkgs/release"
	"github.com/xhd2015/less-flags"
)

const help = `
Usage: go run ./script/release [--dry-run]

Release __NAME__ to GitHub Releases.

Options:
  --dry-run    print what would be done without actually building or uploading
  -h,--help    show help message
`

func main() {
	if err := handle(); err != nil {
		fmt.Fprintf(os.Stderr, "__NAME__ release: %v\n", err)
		os.Exit(1)
	}
}

func handle() error {
	var dryRun bool
	args, err := lessflags.
		Bool("--dry-run", &dryRun).
		Help("-h,--help", help).
		Parse(os.Args[1:])
	if err != nil {
		return err
	}
	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
	}

	if dryRun {
		return handleDryRun()
	}

	creds, err := release.LoadCredentials(".upload-credentials.json")
	if err != nil {
		return err
	}

	result, err := release.BuildRelease("__NAME__", nil, release.DefaultSpecs)
	if err != nil {
		return err
	}

	client := github.NewReleaseClient(creds.Token, creds.Owner, creds.Repo)

	rel, err := client.GetOrCreateRelease(result.Tag)
	if err != nil {
		return fmt.Errorf("failed to get or create release for tag %s: %v", result.Tag, err)
	}

	for _, file := range result.Files {
		if err := client.UploadReleaseAsset(rel.ID, file); err != nil {
			return fmt.Errorf("failed to upload %s: %v", file, err)
		}
		fmt.Printf("Uploaded %s\n", file)
	}
	return nil
}

func handleDryRun() error {
	tag, tagErr := release.GetTag()
	if tagErr != nil {
		fmt.Fprintf(os.Stderr, "[dry-run] warning: %v\n", tagErr)
		tag = "(unknown)"
	}

	creds, credsErr := release.LoadCredentials(".upload-credentials.json")
	if credsErr != nil {
		fmt.Fprintf(os.Stderr, "[dry-run] warning: %v\n", credsErr)
		creds = &release.Credentials{Owner: "__OWNER__", Repo: "__REPO__"}
	}

	fmt.Printf("[dry-run] tag: %s\n", tag)
	for _, spec := range release.DefaultSpecs {
		fmt.Printf("[dry-run] would build: __NAME__-%s-%s-%s\n", tag, spec.OS, spec.Arch)
	}
	fmt.Printf("[dry-run] would upload to %s/%s release (creates if 404)\n", creds.Owner, creds.Repo)
	return nil
}
