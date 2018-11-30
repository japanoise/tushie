package main

import (
	"fmt"
	"os"

	"github.com/japanoise/tushie/src/assembler"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "usage:\n%s infile outfile\n", os.Args[0])
		os.Exit(1)
	}
	err := assembler.Assemble(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
