package create

import "strings"

// applyPlaceholders replaces __KEY__ tokens in s with values from replacements.
func applyPlaceholders(s string, replacements map[string]string) string {
	for key, value := range replacements {
		s = strings.ReplaceAll(s, "__"+key+"__", value)
	}
	return s
}

func standardPlaceholders(projectName, moduleName string) map[string]string {
	m := map[string]string{
		"PROJECT_NAME": projectName,
		"MODULE_NAME":  moduleName,
		"NAME":         projectName,
	}
	owner, repo := parseGitHubOwnerRepo(moduleName)
	if owner == "" {
		owner = "xhd2015"
	}
	if repo == "" {
		repo = projectName
	}
	m["OWNER"] = owner
	m["REPO"] = repo
	return m
}

// parseGitHubOwnerRepo extracts owner/repo from a module path like
// github.com/owner/repo or github.com/owner/repo/subpath.
func parseGitHubOwnerRepo(moduleName string) (owner, repo string) {
	const prefix = "github.com/"
	if !strings.HasPrefix(moduleName, prefix) {
		return "", ""
	}
	rest := strings.TrimPrefix(moduleName, prefix)
	parts := strings.Split(rest, "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", ""
	}
	return parts[0], parts[1]
}