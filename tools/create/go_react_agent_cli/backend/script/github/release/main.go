// usage: go run ./script/github/release [--dry-run]
//
// Proposed behavior (sketch):
//   1. Parse --dry-run / --help flags.
//   2. Resolve tag and credentials (soft-warn on dry-run; hard-fail live).
//   3. Plan artifact names with the same formula as BuildRelease.
//   4. Dry-run: print plan without building or uploading.
//   5. Live: build multi-platform assets, create/upload GitHub Release.
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
Usage: go run ./script/github/release [--dry-run]

Release __PROJECT_NAME__ to GitHub Releases.

Options:
  --dry-run    print what would be done without actually building or uploading
  -h,--help    show help message
`

func main() {
	if err := handle(); err != nil {
		fmt.Fprintf(os.Stderr, "__PROJECT_NAME__ release: %v\n", err)
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

	tag, err := release.GetTag()
	if err != nil {
		if !dryRun {
			return err
		}
		fmt.Fprintf(os.Stderr, "[dry-run] warning: %v\n", err)
		tag = "(unknown)"
	}

	creds, err := release.LoadCredentials(".upload-credentials.json")
	if err != nil {
		if !dryRun {
			return err
		}
		fmt.Fprintf(os.Stderr, "[dry-run] warning: %v\n", err)
		creds = &release.Credentials{Owner: "__OWNER__", Repo: "__REPO__"}
	}

	// Same naming formula as release.BuildRelease.
	var planned []string
	for _, spec := range release.DefaultSpecs {
		planned = append(planned, fmt.Sprintf("__PROJECT_NAME__-%s-%s-%s", tag, spec.OS, spec.Arch))
	}

	if dryRun {
		fmt.Printf("[dry-run] tag: %s\n", tag)
		for _, name := range planned {
			fmt.Printf("[dry-run] would build: %s\n", name)
		}
		fmt.Printf("[dry-run] would upload to %s/%s release (creates if 404)\n", creds.Owner, creds.Repo)
		return nil
	}

	result, err := release.BuildRelease("__PROJECT_NAME__", nil, release.DefaultSpecs)
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
