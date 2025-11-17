package fs

import "os"

func IsSymbolicLink(file string) (bool, error) {
	info, err := os.Lstat(file)
	if err != nil {
		return false, err
	}
	return info.Mode()&os.ModeSymlink != 0, nil
}
