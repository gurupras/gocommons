package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/gurupras/gocommons"
)

func readTest() {
	file := os.Args[1]

	fstruct, err := gocommons.Open(file, os.O_RDONLY, gocommons.GZ_UNKNOWN)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to open file")
		os.Exit(-1)
	}

	defer fstruct.Close()
	if reader, err = fstruct.Reader(0); err != nil {
		return
	}

	reader.Split(bufio.ScanLines)

}
