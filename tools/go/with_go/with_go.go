package with_go

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
)

const downloadGo = "github.com/xhd2015/xgo/script/download-go@master"

func Handle(args []string) error {
	if len(args) == 0 {
		return errors.New("example: kool with-go [GOROOT=<X> | goX.Y] ...")
	}
	var goroot string
	var err error
	arg0 := args[0]
	if arg0 == "list" {
		return List()
	}
	args = args[1:]
	if strings.HasPrefix(arg0, "GOROOT=") {
		goroot = strings.TrimSpace(strings.TrimPrefix(arg0, "GOROOT="))
		if goroot == "" {
			return errors.New("example: kool with-go GOROOT=<X> ...")
		}
	} else {
		goVersion := "go" + strings.TrimPrefix(arg0, "go")
		if goVersion == "" {
			return errors.New("example: kool with-go go1.18 ...")
		}
		goroot, err = ResolveGoroot(goVersion)
		if err != nil {
			return err
		}
	}
	return ExecGoroot(goroot, args)
}

func HandleWithGoroot(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("example: kool with-goroot <GOROOT>")
	}
	return ExecGoroot(args[0], args[1:])
}

func GetInstallDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, "installed"), nil
}

func InstallGo(goVersion string, prompt string) (goRoot string, err error) {
	installDir, err := GetInstallDir()
	if err != nil {
		return "", err
	}
	goRoot = filepath.Join(installDir, goVersion)

	_, statErr := os.Stat(goRoot)
	if !os.IsNotExist(statErr) {
		if statErr != nil {
			return "", statErr
		}
		return goRoot, nil
	}
	if prompt != "" {
		fmt.Fprint(os.Stderr, prompt)
	}
	err = cmd.Debug().Run("go", "run", downloadGo, "download", "--dir", installDir, goVersion)
	if err != nil {
		return "", err
	}

	fmt.Fprintf(os.Stderr, "downloaded: %s, try: \n", goRoot)
	fmt.Fprintf(os.Stderr, "  %s/bin/go version\n", goRoot)
	return goRoot, nil
}

func List() error {
	return cmd.Debug().Run("go", "run", downloadGo, "list")
}
