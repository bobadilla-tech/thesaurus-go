package main

import (
	"fmt"
	"io"

	thesaurus "github.com/bobadilla-tech/thesaurus-go"
)

// runLookup prints the synonyms and antonyms for word, or returns an error
// if word has no entry.
func runLookup(word string, out io.Writer) error {
	entry, ok := thesaurus.Lookup(word)
	if !ok {
		return fmt.Errorf("%q: not found", word)
	}

	fmt.Fprintln(out, "Synonyms:", entry.Synonyms)
	fmt.Fprintln(out, "Antonyms:", entry.Antonyms)
	return nil
}

// runPrefix prints every known word starting with prefix, one per line.
func runPrefix(prefix string, out io.Writer) error {
	matches := thesaurus.WordsWithPrefix(prefix)
	for _, word := range matches {
		fmt.Fprintln(out, word)
	}
	return nil
}

// runCount prints the total number of words the dataset can look up.
func runCount(out io.Writer) error {
	fmt.Fprintln(out, thesaurus.Count())
	return nil
}
