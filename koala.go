package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var files = []string{}

var formats = map[string]func([]string) string{
	"lisp":  lispOut,
	"space": joinOut(" "),
	"comma": joinOut(","),
	"null":  joinOut("\x00"),
}

func main() {
	var rootDirs []string

	curDir := os.Getcwd()
	addRoot := flag.Bool("a", false, "add the root directory to the file output")
	outStyle := flag.String("o", "space", "output style (use help as the style name to see a list)")
	flag.Parse()

	if *outStyle == "help" {
		fmt.Println("Output formats:")
		for format, _ := range formats {
			fmt.Printf("\t%s\n", format)
		}
		return
	}

	if _, ok := formats[*outStyle]; !ok {
		fmt.Fprintf(os.Stderr, "%s is an invalid style.\n", *outStyle)
		os.Exit(1)
	}

	if flag.NArg() == 0 {
		rootDirs = []string{curDir}
	} else {
		rootDirs = flag.Args()
	}

	if len(rootDirs) > 1 {
		*addRoot = true
	}

	for _, scanDir := range rootDirs {
		preScanLen := len(files)
		err := filepath.Walk(scanDir, scan)
		if err != nil {
			files = files[:preScanLen]
			fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}

		if *addRoot {
			for i, path := range files[preScanLen:] {
				files[i] = filepath.Join(scanDir, path)
			}
		}
	}

	formatted := formats[*outStyle](files)
	fmt.Println(formatted)
}

func scan(path string, info os.FileInfo, err error) error {
	files = append(files, path)
	return nil
}

func lispOut(fs []string) string {
	lst := strings.Join(fs, " ")
	lst = "(" + lst + ")"
	return lst
}

func joinOut(sep string) func([]string) string {
	return func(fs string) string {
		return strings.Join(fs, sep)
	}
}
