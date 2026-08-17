package with_go

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/withgo"
	"github.com/xhd2015/xgo/support/downloadgo"
)

func Handle(args []string, envs []string) error {
	return HandleWith(args, envs, withgo.ResolveOptions{Download: true})
}

func HandleWith(args []string, envs []string, opts withgo.ResolveOptions) error {
	if len(args) == 0 {
		return errors.New("example: kool with-go [GOROOT=<X> | goX.Y] ...")
	}
	arg0 := args[0]
	if arg0 == "list" {
		w := opts.Stdout
		if w == nil {
			w = os.Stdout
		}
		return ListWith(context.Background(), downloadgo.ListOptions{}, w)
	}
	args = args[1:]
	if strings.HasPrefix(arg0, "GOROOT=") {
		goroot := strings.TrimSpace(strings.TrimPrefix(arg0, "GOROOT="))
		if goroot == "" {
			return errors.New("example: kool with-go GOROOT=<X> ...")
		}
		return ExecGoroot(goroot, args, envs)
	}
	goVersion := "go" + strings.TrimPrefix(arg0, "go")
	if goVersion == "" {
		return errors.New("example: kool with-go go1.18 ...")
	}
	goroot, err := ResolveGorootWith(goVersion, opts)
	if err != nil {
		return err
	}
	return ExecGoroot(goroot, args, envs)
}

func HandleWithGoroot(args []string, envs []string) error {
	if len(args) == 0 {
		return fmt.Errorf("example: kool with-goroot <GOROOT>")
	}
	return ExecGoroot(args[0], args[1:], envs)
}

func GetInstallDir() (string, error) {
	return withgo.DefaultInstallDir()
}

func InstallGo(goVersion string, prompt string) (string, error) {
	return ResolveGorootWith(goVersion, withgo.ResolveOptions{
		Download: true,
		Prompt:   prompt,
		Stderr:   os.Stderr,
	})
}

func List() error {
	return ListWith(context.Background(), downloadgo.ListOptions{}, os.Stdout)
}

func ListWith(ctx context.Context, listOpts downloadgo.ListOptions, w io.Writer) error {
	versions, err := downloadgo.List(ctx, listOpts)
	if err != nil {
		return err
	}
	if w == nil {
		w = os.Stdout
	}
	for _, v := range versions {
		fmt.Fprintf(w, "go%s\n", v)
	}
	return nil
}
