package modcache

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func formatBytes(n int64) string {
	if n < 0 {
		n = 0
	}
	const (
		k = 1024
		m = 1024 * 1024
		g = 1024 * 1024 * 1024
	)
	switch {
	case n >= g:
		return formatFrac(float64(n)/float64(g)) + "G"
	case n >= m:
		return formatFrac(float64(n)/float64(m)) + "M"
	case n >= k:
		return formatFrac(float64(n)/float64(k)) + "K"
	default:
		return fmt.Sprintf("%dB", n)
	}
}

func formatFrac(v float64) string {
	s := fmt.Sprintf("%.1f", v)
	return strings.TrimSuffix(s, ".0")
}

func dirSize(root string) int64 {
	var n int64
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := os.Lstat(path)
		if err != nil {
			return nil
		}
		n += info.Size()
		return nil
	})
	return n
}

func removePath(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.IsDir() {
		_ = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			mode := os.FileMode(0666)
			if d.IsDir() {
				mode = 0777
			}
			_ = os.Chmod(p, mode)
			return nil
		})
		return os.RemoveAll(path)
	}
	_ = os.Chmod(path, 0666)
	return os.Remove(path)
}
