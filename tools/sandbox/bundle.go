package sandbox

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type metaYAML struct {
	Name      string `yaml:"name"`
	Comment   string `yaml:"comment"`
	ExpiresAt string `yaml:"expires_at"`
}

// buildOpts holds merged build inputs after flag parse.
type buildOpts struct {
	Output string
	Input  string
	Files  []string // LOCAL=SANDBOX_REL
	Env    []string // KEY=VALUE
	Goos   string
	Goarch string
}

// mergePack builds a PackBlob from -i directory and flag overrides.
// Flags win on the same relative path or env key.
func mergePack(opts *buildOpts) (*PackBlob, error) {
	blob := &PackBlob{
		Version:   packBlobVersion,
		Name:      "sandbox",
		CreatedAt: time.Now().UTC(),
		Files:     nil,
		Env:       map[string]string{},
	}

	fileMap := map[string]PackFile{} // path -> file

	if opts.Input != "" {
		if err := loadInputDir(opts.Input, blob, fileMap); err != nil {
			return nil, err
		}
	}

	// Flag files override dir paths.
	for _, entry := range opts.Files {
		local, rel, err := splitFileFlag(entry)
		if err != nil {
			return nil, err
		}
		data, err := os.ReadFile(local)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("file not found: %s", local)
			}
			return nil, fmt.Errorf("read --file %s: %w", local, err)
		}
		st, err := os.Stat(local)
		if err != nil {
			return nil, err
		}
		mode := uint32(st.Mode().Perm())
		rel = filepath.ToSlash(rel)
		fileMap[rel] = PackFile{Path: rel, Mode: mode, Content: data}
	}

	// Flag env overrides dir keys.
	for _, entry := range opts.Env {
		key, val, err := splitEnvFlag(entry)
		if err != nil {
			return nil, err
		}
		blob.Env[key] = val
	}

	// Stable file order for determinism of structure (ciphertext still random).
	paths := make([]string, 0, len(fileMap))
	for p := range fileMap {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	blob.Files = make([]PackFile, 0, len(paths))
	for _, p := range paths {
		blob.Files = append(blob.Files, fileMap[p])
	}

	if len(blob.Files) == 0 && len(blob.Env) == 0 {
		return nil, fmt.Errorf("empty pack: require at least one file or env entry")
	}
	return blob, nil
}

func loadInputDir(dir string, blob *PackBlob, fileMap map[string]PackFile) error {
	st, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("input dir: %w", err)
	}
	if !st.IsDir() {
		return fmt.Errorf("input is not a directory: %s", dir)
	}

	metaPath := filepath.Join(dir, "meta.yaml")
	if data, err := os.ReadFile(metaPath); err == nil {
		var m metaYAML
		if err := yaml.Unmarshal(data, &m); err != nil {
			return fmt.Errorf("parse meta.yaml: %w", err)
		}
		if m.Name != "" {
			blob.Name = m.Name
		}
		if m.Comment != "" {
			blob.Comment = m.Comment
		}
		if m.ExpiresAt != "" {
			t, err := time.Parse(time.RFC3339, m.ExpiresAt)
			if err == nil {
				blob.ExpiresAt = &t
			}
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read meta.yaml: %w", err)
	}

	filesRoot := filepath.Join(dir, "files")
	if st, err := os.Stat(filesRoot); err == nil && st.IsDir() {
		err := filepath.WalkDir(filesRoot, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(filesRoot, path)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			fileMap[rel] = PackFile{
				Path:    rel,
				Mode:    uint32(info.Mode().Perm()),
				Content: data,
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("walk files/: %w", err)
		}
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("files/: %w", err)
	}

	envPath := filepath.Join(dir, "env.yaml")
	if data, err := os.ReadFile(envPath); err == nil {
		// Allow plain KEY: value map.
		var env map[string]string
		if err := yaml.Unmarshal(data, &env); err != nil {
			return fmt.Errorf("parse env.yaml: %w", err)
		}
		for k, v := range env {
			blob.Env[k] = v
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read env.yaml: %w", err)
	}

	return nil
}

func splitFileFlag(entry string) (local, rel string, err error) {
	i := strings.Index(entry, "=")
	if i <= 0 || i == len(entry)-1 {
		return "", "", fmt.Errorf("invalid --file (want LOCAL=SANDBOX_REL): %s", entry)
	}
	return entry[:i], entry[i+1:], nil
}

func splitEnvFlag(entry string) (key, val string, err error) {
	i := strings.Index(entry, "=")
	if i <= 0 {
		return "", "", fmt.Errorf("invalid --env (want KEY=VALUE): %s", entry)
	}
	return entry[:i], entry[i+1:], nil
}
