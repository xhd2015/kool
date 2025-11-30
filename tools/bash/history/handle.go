package history

import "github.com/xhd2015/kool/tools/stringtool"

func GetHistoryLines() ([]string, error) {
	// all files
	allFiles, err := GetAllHistoryFiles()
	if err != nil {
		return nil, err
	}
	var lines []string
	for _, file := range allFiles {
		fileLines, err := ReadNonEmptyLines(file)
		if err != nil {
			return nil, err
		}
		lines = append(lines, stringtool.Reverse(fileLines)...)
	}
	return stringtool.Uniq(lines), nil
}

func DelHistoryLine(line string) error {
	// all files
	allFiles, err := GetAllHistoryFiles()
	if err != nil {
		return err
	}

	for _, file := range allFiles {
		err = delLineFromFile(file, line)
		if err != nil {
			return err
		}
	}
	return nil
}
