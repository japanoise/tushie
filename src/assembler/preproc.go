package assembler

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func preProc(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(file)
	ret := strings.Builder{}
	linum := 0
	for scanner.Scan() {
		linum++
		line := scanner.Text()
		if len(line) > 1 {
			if line[0] == '#' {
				dr := strings.ToLower(strings.Split(line[1:], " ")[0])
				switch dr {
				case "include":
					if len(line) <= 9 {
						return "", fmt.Errorf("In file %s, line %d: no filename provided to #include", filename, linum)
					}
					s, err := preProc(line[9:])
					if err != nil {
						return "", err
					}
					ret.WriteString(s)
				case "incbin":
					// Using "base64", which is less wasteful than db
					ret.WriteString("\tbase64 \"")
					if len(line) <= 8 {
						return "", fmt.Errorf("In file %s, line %d: no filename provided to #incbin", filename, linum)
					}
					s := (line[8:])
					file, err := os.Open(s)
					if err != nil {
						return "", err
					}
					defer file.Close()
					data, err := ioutil.ReadAll(file)
					if err != nil {
						return "", err
					}
					// Break it up so the lines don't get too long.
					str := base64.StdEncoding.EncodeToString(data)
					oldidx := 0
					idx := 64
					for len(str) > idx {
						ret.WriteString(str[oldidx:idx])
						ret.WriteRune('\n')
						oldidx = idx
						idx += 64
					}
					ret.WriteString(str[oldidx:])
					ret.WriteString("\"\n")
				default:
					return "", fmt.Errorf("In file %s, line %d: unknown preprocessor directive #%s", filename, linum, dr)
				}
			} else if line[0] != ';' {
				ret.WriteString(line)
				ret.WriteRune('\n')
			}
		}
	}

	return ret.String(), nil
}
