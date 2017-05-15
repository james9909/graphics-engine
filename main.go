package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
)

var profile = flag.Bool("profile", false, "Profile")

func main() {
	flag.Parse()
	args := flag.Args()
	parser := NewParser()

	if *profile {
		f, err := os.Create("cpu.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	var err error
	if len(args) == 0 {
		err = parser.ParseInput()
	} else {
		err = parser.ParseFile(args[0])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Parse error:", err)
		os.Exit(1)
	}
}
