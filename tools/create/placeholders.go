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
	return map[string]string{
		"PROJECT_NAME": projectName,
		"MODULE_NAME":  moduleName,
	}
}