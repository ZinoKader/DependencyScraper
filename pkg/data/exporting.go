package data

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func fileExists(file string) bool {
	if _, err := os.Stat(file); err == nil {
		return true
	}
	return false
}

func WriteToFile(file string, content []byte) error {
	finalFilePath := file
	if fileExists(file) {
		// find where to put the suffix and the eventual file extension
		fileExtension := ""
		suffixPosition := len(file)
		if strings.Contains(file, ".") {
			suffixPosition = strings.LastIndex(file, ".")
			fileExtension = file[suffixPosition:]
		}

		// add a suffix if the file already exists
		suffix := 0
		for fileExists(finalFilePath) {
			finalFilePath = fmt.Sprintf("%s%d", file[0:suffixPosition], suffix)
			suffix += 1
		}

		// re-add file extension
		if len(fileExtension) != 0 {
			finalFilePath = fmt.Sprintf("%s%s", finalFilePath, fileExtension)
		}
	}

	f, err := os.Create(finalFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	w.Write(content)
	w.Flush()

	return nil
}

func AppendToFile(file string, content string) error {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(fmt.Sprintf("%s\n", content)); err != nil {
		return err
	}
	return nil
}
