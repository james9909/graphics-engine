package main

import "os"

func main() {
	parser := NewParser()
	var err error
	if len(os.Args) < 2 {
		err = parser.ParseInput()
	} else {
		err = parser.ParseFile(os.Args[1])
	}
	if err != nil {
		panic(err)
	}
}
