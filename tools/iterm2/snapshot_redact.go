package iterm2

import "regexp"

var redactPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(--app-secret=)\S+`),
	regexp.MustCompile(`(--app-secret\s+)\S+`),
	regexp.MustCompile(`(--secret=)\S+`),
	regexp.MustCompile(`(--token=)\S+`),
	regexp.MustCompile(`(--password=)\S+`),
	regexp.MustCompile(`(?i)(Authorization:\s*Bearer\s+)\S+`),
}

func redactCommandLine(s string) string {
	out := s
	for _, re := range redactPatterns {
		out = re.ReplaceAllString(out, `${1}***`)
	}
	return out
}
