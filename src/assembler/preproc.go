package assembler

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func preProc(filename string, astate *state) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
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
						return fmt.Errorf("In file %s, line %d: no filename provided to #include", filename, linum)
					}
					err := preProc(line[9:], astate)
					if err != nil {
						return fmt.Errorf("In file included from %s, line %d: %s",
							filename, linum, err.Error())
					}
				case "incbin":
					// Using "base64", which is less wasteful than db
					cline := sourceLine{"\tbase64 \"", filename, linum}
					if len(line) <= 8 {
						return fmt.Errorf("In file %s, line %d: no filename provided to #incbin", filename, linum)
					}
					s := (line[8:])
					file, err := os.Open(s)
					if err != nil {
						return err
					}
					defer file.Close()
					data, err := ioutil.ReadAll(file)
					if err != nil {
						return err
					}
					// Break it up so the lines don't get too long.
					str := base64.StdEncoding.EncodeToString(data)
					oldidx := 0
					idx := 64
					for len(str) > idx {
						cline.data += str[oldidx:idx]
						astate.source = append(astate.source, cline)
						cline = sourceLine{"", filename, linum}
						oldidx = idx
						idx += 64
					}
					cline.data += str[oldidx:idx]
					astate.source = append(astate.source, cline)
				default:
					return fmt.Errorf("In file %s, line %d: unknown preprocessor directive #%s", filename, linum, dr)
				}
			} else if line[0] != ';' {
				astate.source = append(astate.source, sourceLine{line, filename, linum})
			}
		}
	}

	return nil
}
