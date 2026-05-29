package scaffold

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"sort"
	"strings"

	"github.com/xhd2015/less-gen/flags"
)

//go:embed all:github_publish
var githubPublishFS embed.FS

const help = `
Usage: kool scaffold [--list]
       kool scaffold <name>

Scaffolds:
  go-cmd-run-lib
  github-publish

Options:
  --list       list available scaffold names
  -h,--help    show help message
`

type scaffold struct {
	Name    string
	Content string
	Files   embed.FS
}

var scaffolds = []scaffold{
	{
		Name:    "go-cmd-run-lib",
		Content: goCmdRunLibScaffold,
	},
	{
		Name:  "github-publish",
		Files: githubPublishFS,
	},
}

const goCmdRunLibScaffold = `# cmd/__NAME__/main.go
package main

import (
    "fmt"
    "os"

    __NAME__ "__MODULE__/run/__NAME__"
)

func main() {
    if err := __NAME__.Run(os.Args[1:]); err != nil {
        fmt.Fprintf(os.Stderr, "__NAME__: %v\n", err)
        os.Exit(1)
    }
}

# run/__NAME__/run.go
package __NAME__

import (
    "fmt"
    "strings"

    core "__MODULE__/pkgs/__NAME__"
    "github.com/xhd2015/less-gen/flags"
)

const help = ` + "`" + `
Usage: __NAME__ [OPTIONS]

Options:
  -h,--help    show help message
` + "`" + `

func Run(args []string) error {
    config := core.Config{}
    args, err := flags.
        Help("-h,--help", help).
        Parse(args)
    if err != nil {
        return err
    }
    if len(args) > 0 {
        return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
    }
    return core.Run(config)
}

# pkgs/__NAME__/__NAME__.go
package __NAME__

type Config struct{}

func Run(config Config) error {
    // Core library logic goes here.
    return nil
}
`

func Handle(args []string) error {
	return HandleWithWriter(os.Stdout, args)
}

func HandleWithWriter(w io.Writer, args []string) error {
	var list bool
	args, err := flags.Bool("--list", &list).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if list {
		if len(args) > 0 {
			return fmt.Errorf("--list does not accept arguments")
		}
		for _, scaffold := range scaffolds {
			fmt.Fprintln(w, scaffold.Name)
		}
		return nil
	}

	if len(args) == 0 {
		fmt.Fprint(w, strings.TrimPrefix(help, "\n"))
		return nil
	}
	if len(args) > 1 {
		return fmt.Errorf("unrecognized extra arguments: %s", strings.Join(args[1:], " "))
	}

	s, ok := lookup(args[0])
	if !ok {
		return fmt.Errorf("unknown scaffold: %s", args[0])
	}

	if s.Content != "" {
		fmt.Fprint(w, s.Content)
		return nil
	}

	return writeFilesFromFS(w, s.Files)
}

func lookup(name string) (*scaffold, bool) {
	for i := range scaffolds {
		if scaffolds[i].Name == name {
			return &scaffolds[i], true
		}
	}
	return nil, false
}

func writeFilesFromFS(w io.Writer, fsys embed.FS) error {
	// find the template root directory (e.g. "github_publish")
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}
	root := entries[0].Name()

	var paths []string
	err = fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		// strip root prefix to get relative path
		relPath := strings.TrimPrefix(path, root+"/")
		paths = append(paths, relPath)
		return nil
	})
	if err != nil {
		return err
	}
	sort.Strings(paths)

	for _, relPath := range paths {
		content, err := fs.ReadFile(fsys, root+"/"+relPath)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "# %s\n%s", relPath, content)
	}
	return nil
}
