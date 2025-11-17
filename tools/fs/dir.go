package fs

import (
	"fmt"
	"os"
)

func ValidateIsDir(dir string) error {
	dirStat, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !dirStat.IsDir() {
		return fmt.Errorf("not a directory: %s", dir)
	}
	return nil
}

func ValidateNotExists(name string) error {
	_, err := os.Stat(name)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	return fmt.Errorf("%s already exists", name)
}
