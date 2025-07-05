package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
)

func main() {
	err := handle(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type Spec struct {
	Arch string
	OS   string
}

func handle(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, ","))
	}

	out, err := cmd.Output("git", "status", "--porcelain")
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(out)) != "" {
		return fmt.Errorf("git status is not clean, ensure everything is commited. check with 'git status'")
	}

	// "git", "describe", "--tags", "HEAD"
	tag, err := cmd.Output("git", "describe", "--tags", "HEAD")
	if err != nil {
		return err
	}
	tag = strings.TrimSpace(string(tag))
	if tag == "" {
		return fmt.Errorf("no tag found, ensure you are on a tagged commit")
	}

	if !strings.HasPrefix(tag, "v") {
		return fmt.Errorf("tag %s is not a valid version, must start with 'v'", tag)
	}

	specs := []*Spec{
		{
			Arch: "amd64",
			OS:   "darwin",
		},
		{
			Arch: "arm64",
			OS:   "darwin",
		},
		{
			Arch: "amd64",
			OS:   "linux",
		},
		{
			Arch: "arm64",
			OS:   "linux",
		},
	}
	for _, spec := range specs {
		filename := fmt.Sprintf("kool-%s-%s-%s", tag, spec.OS, spec.Arch)

		err := cmd.Debug().Env([]string{
			"GOOS=" + spec.OS,
			"GOARCH=" + spec.Arch,
		}).Run("go", "build", "-o", filename, "./")
		if err != nil {
			return err
		}
	}
	return nil
}
