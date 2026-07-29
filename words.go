package thesaurus

import (
	"sort"
	"strings"
)

// WordsWithPrefix returns every known word starting with prefix (case-
// insensitive, using the same normalisation as Lookup), in sorted order.
// Returns nil if no word matches.
func WordsWithPrefix(prefix string) []string {
	prefix = normalise(prefix)

	start := sort.SearchStrings(allWords, prefix)

	var matches []string
	for i := start; i < len(allWords) && strings.HasPrefix(allWords[i], prefix); i++ {
		matches = append(matches, allWords[i])
	}
	return matches
}

// Contains reports whether word has an entry, equivalent to the ok return
// value of Lookup without needing the Entry itself.
func Contains(word string) bool {
	_, ok := Lookup(word)
	return ok
}

// Count returns the total number of distinct words Lookup can answer for.
func Count() int {
	return len(allWords)
}

// AllWords returns every known word, sorted. The result is a copy; mutating
// it does not affect the package's internal state.
func AllWords() []string {
	return append([]string(nil), allWords...)
}
