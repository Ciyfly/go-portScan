package util

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

func GetLines(filename string) (out []string, err error) {
	if filename == "" {
		return out, errors.New("no filename")
	}
	file, err := os.Open(filename)
	if err != nil {
		return out, err
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			out = append(out, line)
		}
	}
	return
}
