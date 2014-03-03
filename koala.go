// Koala is a package for scanning a directory, returning a list of
// files. It's similar to ls -R1 | xargs, except that it only lists
// the files. It is used as a building block in other systems.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	files       = []string{}
	showDirs    *bool
	showHidden  *bool
	noStripRoot *bool
	curDir      string
)

var formats = map[string]func([]string) string{
	"lisp":  lispOut,
	"space": joinOut(" "),
	"comma": joinOut(","),
	"null":  joinOut("\x00"),
}

func main() {
	var (
		rootDirs []string
		err      error
	)

	// err is checked later; failure to retrieve the current directory
	// is only fatal if no directories are passed in or we aren't stripping
	// the current working directory.
	curDir, err = os.Getwd()

	showHidden = flag.Bool("a", false, "show hidden (dot) files")
	showDirs = flag.Bool("d", false, "show directories")
	outStyle := flag.String("o", "space", "output style (use help as the style name to see a list)")
	noStripRoot = flag.Bool("s", false, "don't strip the current working directory from the output")
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
		if err != nil {
			fmt.Fprintf(os.Stderr, "No directories specified, and cannot determine current directory.")
		}
		rootDirs = []string{curDir}
	} else {
		rootDirs = flag.Args()
	}

	for _, scanDir := range rootDirs {
		preScanLen := len(files)
		scan := func(path string, info os.FileInfo, err error) error {
			if path == scanDir {
				return nil
			} else if info.IsDir() && !*showDirs {
				return nil
			} else if !*showHidden {
				if filepath.Base(path)[0] == '.' {
					if info.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}

			if !*noStripRoot {
				relPath, err := filepath.Rel(scanDir, path)
				if err != nil {
					return err
				}
				if len(relPath) >= 2 {
					if relPath[:2] != ".." {
						path = relPath
					}
				}
			}
			files = append(files, path)
			return nil
		}

		err := filepath.Walk(scanDir, scan)
		if err != nil {
			files = files[:preScanLen]
			fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}
	}

	formatted := formats[*outStyle](files)
	fmt.Println(formatted)
}

func lispOut(fs []string) string {
	lst := strings.Join(fs, " ")
	lst = "'(" + lst + ")"
	return lst
}

func joinOut(sep string) func([]string) string {
	return func(fs []string) string {
		return strings.Join(fs, sep)
	}
}
