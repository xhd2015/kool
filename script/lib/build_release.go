package lib

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
)

type Spec struct {
	Arch string
	OS   string
}

var DefaultSpecs = []*Spec{
	{Arch: "amd64", OS: "darwin"},
	{Arch: "arm64", OS: "darwin"},
	{Arch: "amd64", OS: "linux"},
	{Arch: "arm64", OS: "linux"},
}

type BuildReleaseResult struct {
	Tag   string
	Files []string
}

func BuildRelease(specs []*Spec) (*BuildReleaseResult, error) {
	out, err := cmd.Output("git", "status", "--porcelain")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(string(out)) != "" {
		return nil, fmt.Errorf("git status is not clean, ensure everything is committed. check with 'git status'")
	}

	tag, err := cmd.Output("git", "describe", "--tags", "HEAD")
	if err != nil {
		return nil, err
	}
	tag = strings.TrimSpace(string(tag))
	if tag == "" {
		return nil, fmt.Errorf("no tag found, ensure you are on a tagged commit")
	}

	if !strings.HasPrefix(tag, "v") {
		return nil, fmt.Errorf("tag %s is not a valid version, must start with 'v'", tag)
	}

	if err := buildFrontend(); err != nil {
		return nil, fmt.Errorf("frontend build failed: %v", err)
	}

	var files []string
	for _, spec := range specs {
		filename := fmt.Sprintf("kool-%s-%s-%s", tag, spec.OS, spec.Arch)

		err := cmd.Debug().Env([]string{
			"GOOS=" + spec.OS,
			"GOARCH=" + spec.Arch,
		}).Run("go", "build", "-o", filename, "./")
		if err != nil {
			return nil, err
		}
		files = append(files, filename)
	}
	return &BuildReleaseResult{Tag: tag, Files: files}, nil
}

func buildFrontend() error {
	return cmd.Debug().Run("go", "run", "./script/build-react")
}
