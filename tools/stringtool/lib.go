package stringtool

func PrefixLines(lines []string, prefix string) []string {
	newLines := make([]string, len(lines))
	for i, line := range lines {
		newLines[i] = prefix + line
	}
	return newLines
}
