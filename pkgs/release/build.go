package release

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

// Option configures BuildRelease.
type Option func(*buildConfig)

type buildConfig struct {
	packagePath string // go build package path; default "./"
}

// WithPackagePath sets the package path passed to go build (e.g. "./cmd/doctest").
// Empty path is ignored (keeps the default "./").
func WithPackagePath(path string) Option {
	return func(c *buildConfig) {
		if path != "" {
			c.packagePath = path
		}
	}
}

func applyOptions(opts []Option) *buildConfig {
	cfg := &buildConfig{packagePath: "./"}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}
	return cfg
}

func BuildRelease(binaryName string, preBuild func() error, specs []*Spec, opts ...Option) (*BuildReleaseResult, error) {
	cfg := applyOptions(opts)

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

	if preBuild != nil {
		if err := preBuild(); err != nil {
			return nil, fmt.Errorf("pre-build failed: %w", err)
		}
	}

	var files []string
	for _, spec := range specs {
		filename := fmt.Sprintf("%s-%s-%s-%s", binaryName, tag, spec.OS, spec.Arch)

		err := cmd.Debug().Env([]string{
			"GOOS=" + spec.OS,
			"GOARCH=" + spec.Arch,
		}).Run("go", "build", "-o", filename, cfg.packagePath)
		if err != nil {
			return nil, err
		}
		files = append(files, filename)
	}
	return &BuildReleaseResult{Tag: tag, Files: files}, nil
}
