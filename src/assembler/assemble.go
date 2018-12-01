package assembler

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/japanoise/numparse"
)

// Assemble assembles the sourcecode in infile and writes it to the binary file outfile, or returns an error if any stage fails
func Assemble(infile, outfile string) error {
	res, err := preProc(infile)
	if err != nil {
		return err
	}
	err = asm(res, outfile)
	if err != nil {
		return err
	}
	return nil
}

type asmState uint8

const (
	asmStateOps asmState = iota
	asmStateB64
)

func asm(sourcecode, outfile string) error {
	out, err := os.Create(outfile)
	if err != nil {
		return err
	}
	defer out.Close()
	state := asmStateOps
	var buf bytes.Buffer

	for _, line := range strings.Split(sourcecode, "\n") {
		switch state {
		case asmStateOps:
			// Trim space
			tline := strings.TrimSpace(line)

			// Strip comments and labels
			clbuild := strings.Builder{}
			if strings.Contains(tline, ":") {
				string := false
				escape := false
				for _, ru := range tline {
					if ru == ':' && !string {
						label := clbuild.String()
						tline = tline[len(label)+1:]
						// This is where we will add the label to the label list
						break
					} else if ru == ';' && !string {
						break
					} else if ru == '"' && !escape {
						string = !string
					} else if string && !escape && ru == '\\' {
						escape = true
					} else {
						if escape {
							escape = false
						}
						clbuild.WriteRune(ru)
					}
				}
				// Reset the builder so we can use it to snarf the code
				clbuild = strings.Builder{}
			}
			if strings.Contains(tline, ";") {
				string := false
				escape := false
				for _, ru := range tline {
					if ru == ';' && !string {
						break
					} else if ru == '"' && !escape {
						string = !string
					} else if string && !escape && ru == '\\' {
						escape = true
					} else {
						if escape {
							escape = false
						}
						clbuild.WriteRune(ru)
					}
				}
				tline = clbuild.String()
			}
			// Trim again, because the labeller and comment stripper may have left ws
			tline = strings.TrimSpace(tline)

			// "tokenize" and act on the op
			spl := strings.Split(tline, " ")
			if len(spl) < 1 {
				continue
			}
			op := strings.ToLower(spl[0])
			switch op {
			case "":
				// Empty line, do nothing
			case "base64":
				if !strings.Contains(line, "\"") {
					return errors.New("base64 argument must be enclosed by double quotes")
				}
				buf = bytes.Buffer{}
				splqu := strings.Split(line, "\"")
				if len(splqu) > 1 {
					buf.WriteString(splqu[1])
					if len(splqu) == 2 {
						state = asmStateB64
					} else {
						data, err := base64.StdEncoding.DecodeString(buf.String())
						if err != nil {
							return err
						}
						out.Write(data)
						out.Sync()
					}
				}
			case "db":
				if len(spl) < 1 {
					return errors.New("db requires at least one argument")
				}
				argsstr := strings.Join(spl[1:], " ")
				instring := false
				escape := false
				strarg := false
				nbuf := strings.Builder{}
				bufd := false
				for _, ru := range argsstr {
					if !escape && ru == '"' {
						instring = !instring
						strarg = true
					} else if instring && ru == '\\' && !escape {
						escape = true
					} else if instring {
						out.WriteString(string(ru))
						escape = false
					} else if ru == ',' {
						if strarg {
							nbuf = strings.Builder{}
							bufd = false
							continue
						}
						if !bufd {
							return errors.New("malformed arguments to db")
						}
						res, err := numparse.UNumParse(nbuf.String())
						if err != nil {
							return err
						}
						if res > 0xFF {
							return errors.New("argument to db larger than 0xFF")
						}
						out.Write([]byte{byte(res)})
						nbuf = strings.Builder{}
						bufd = false
					} else if !unicode.IsSpace(ru) {
						bufd = true
						nbuf.WriteRune(ru)
					}
				}
				if bufd {
					res, err := numparse.UNumParse(nbuf.String())
					if err != nil {
						return err
					}
					if res > 0xFF {
						return fmt.Errorf("argument to db larger than 0xFF: %d/0o%o/0x%X", res, res, res)
					}
					out.Write([]byte{byte(res)})
				}
				out.Sync()
			default:
				return fmt.Errorf("unknown opcode %s", op)
			}

		case asmStateB64:
			if strings.Contains(line, "\"") {
				splqu := strings.Split(line, "\"")
				if len(splqu) >= 1 {
					buf.WriteString(splqu[0])
				}
				data, err := base64.StdEncoding.DecodeString(buf.String())
				if err != nil {
					return err
				}
				out.Write(data)
				out.Sync()
				state = asmStateOps
			} else {
				buf.WriteString(line)
			}
		}
	}

	if state == asmStateB64 {
		return errors.New("unterminated base64")
	}

	return nil
}
