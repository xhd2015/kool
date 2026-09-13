// usage: go run ./script/install
package main

import (
	"fmt"
	"os"

	localinstall "github.com/xhd2015/dot-pkgs/go-pkgs/gotool/localbin/install"
	"github.com/xhd2015/xgo/support/cmd"
)

func main() {
	if err := handle(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func handle() error {
	fmt.Println("==> Building")
	if err := cmd.Debug().Run("go", "run", "./script/build"); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Println("==> Installing")
	_, err = localinstall.Install(localinstall.Options{
		Dir:     root,
		Package: "./cmd/__PROJECT_NAME__",
		BinName: "__PROJECT_NAME__",
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	})
	if err != nil {
		return err
	}
	fmt.Println("install complete")
	return nil
}
