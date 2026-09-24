//go:build ignore

// Delete guard: removing a picture a page still shows breaks that page, so a
// referenced image is refused until the reference is detached. The detach walks
// the same container registry the audit does, which is the only way the two can
// never disagree about what is supported.
package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Delete removes an image from the library.
//
// A referenced image is refused unless force is set, because deleting one
// leaves the pages that show it with a broken picture. With force every
// reference is detached first, so no dangling id is left behind. The return
// value counts the documents that were rewritten.
func (s *imageStore) Delete(id string, force bool) (int, error) {
	id = strings.TrimSpace(id)
	if !isDecimalID(id) {
		return 0, &validationError{"invalid image id"}
	}
	s.mu.RLock()
	meta, err := s.loadLocked(id)
	s.mu.RUnlock()
	if err != nil {
		return 0, err
	}
	usages, err := s.Usages(id)
	if err != nil {
		return 0, err
	}
	if len(usages) > 0 {
		// A reference this build cannot rewrite must block the delete:
		// removing the picture would leave a page pointing at an id that no
		// longer exists, which is the one outcome the guard exists to prevent.
		if undetachable := undetachableUsages(usages); len(undetachable) > 0 {
			return 0, &validationError{fmt.Sprintf(
				"image %s is used by %s, which this build cannot rewrite\n  remove the reference there, then delete the image",
				id, describeUsages(undetachable))}
		}
		if !force {
			return 0, &validationError{fmt.Sprintf(
				"image %s is used by %s\n  detach it first, or pass --force to delete and detach",
				id, describeUsages(usages))}
		}
	}
	detached := 0
	if force {
		if detached, err = s.detach(id); err != nil {
			return detached, err
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.RemoveAll(s.imageDir(meta.ID)); err != nil {
		return detached, err
	}
	return detached, nil
}

// detach drops every reference to one image, driving off the container
// registry so the guard and the delete can never disagree about what is
// supported. It returns how many documents it rewrote.
func (s *imageStore) detach(id string) (int, error) {
	containers := map[string]imageContainer{}
	for _, container := range imageContainers {
		if container.detach != nil {
			containers[container.file] = container
		}
	}
	files := []string{}
	err := filepath.WalkDir(s.dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != s.dir && d.Name() == "images" {
				return filepath.SkipDir
			}
			return nil
		}
		if _, ok := containers[d.Name()]; ok {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	detached := 0
	for _, path := range files {
		changed, err := containers[filepath.Base(path)].detach(s, path, id)
		if err != nil {
			return detached, fmt.Errorf("detach from %s: %w", path, err)
		}
		if changed {
			detached++
		}
	}
	return detached, nil
}

func describeUsages(usages []ImageUsage) string {
	parts := make([]string, 0, len(usages))
	for _, usage := range usages {
		text := usage.PagePath
		if text == "" {
			text = usage.Kind
		}
		parts = append(parts, text+" · "+usage.Kind)
	}
	return strings.Join(parts, "\n  ")
}

// undetachableUsages picks out the references this build cannot rewrite.
func undetachableUsages(usages []ImageUsage) []ImageUsage {
	out := []ImageUsage{}
	for _, usage := range usages {
		if !usage.Detachable {
			out = append(out, usage)
		}
	}
	return out
}

func appendUsage(list []ImageUsage, usage ImageUsage) []ImageUsage {
	for _, existing := range list {
		if existing.Kind == usage.Kind && existing.PagePath == usage.PagePath {
			return list
		}
	}
	return append(list, usage)
}
