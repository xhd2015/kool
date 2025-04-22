package with_go

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xhd2015/xgo/support/cmd"
)

const downloadGo = "github.com/xhd2015/xgo/script/download-go@master"

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
