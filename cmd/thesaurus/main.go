// Command thesaurus is a small example CLI over the thesaurus package,
// demonstrating Lookup, WordsWithPrefix, and Count from the terminal.
//
// Usage:
//
//	thesaurus lookup <word>
//	thesaurus prefix <prefix>
//	thesaurus count
package main

import (
	"fmt"
	"os"
)

func usage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  thesaurus lookup <word>")
	fmt.Fprintln(os.Stderr, "  thesaurus prefix <prefix>")
	fmt.Fprintln(os.Stderr, "  thesaurus count")
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var err error
	switch cmd := os.Args[1]; cmd {
	case "lookup":
		if len(os.Args) != 3 {
			usage()
			os.Exit(2)
		}
		err = runLookup(os.Args[2], os.Stdout)
	case "prefix":
		if len(os.Args) != 3 {
			usage()
			os.Exit(2)
		}
		err = runPrefix(os.Args[2], os.Stdout)
	case "count":
		if len(os.Args) != 2 {
			usage()
			os.Exit(2)
		}
		err = runCount(os.Stdout)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", cmd)
		usage()
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
