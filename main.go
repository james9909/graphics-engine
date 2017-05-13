package main

import (
	"fmt"
	"os"
)

func main() {
	parser := NewParser()
	var err error
	if len(os.Args) < 2 {
		err = parser.ParseInput()
	} else {
		err = parser.ParseFile(os.Args[1])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Parse error:", err)
		os.Exit(1)
	}
}
