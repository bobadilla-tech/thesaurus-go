package thesaurus

import "strings"

// Lookup returns the Entry for a word, combining the curated dataset and
// the OEWN-derived dataset. The curated dataset always wins when it has an
// entry for the word; OEWN only fills in words the curated list doesn't
// cover. Both synonyms and antonyms are resolved this same way, since both
// fields in the curated dataset come from the same hand-reviewed source and
// should be trusted together, not mixed field-by-field with OEWN.
//
// Returns false only if the word has neither synonyms nor antonyms in
// either source.
func Lookup(word string) (Entry, bool) {
	return lookup(normalise(word), curatedData, synonymsOEWNData, antonymsOEWNData)
}

// lookup implements the merge logic in isolation from the package-level
// embedded data, so it can be unit tested with fixture maps instead of the
// real (large, generated) OEWN dataset. word is expected to already be
// normalised by the caller.
func lookup(word string, curated map[string]Entry, synonymsOEWN, antonymsOEWN map[string][]string) (Entry, bool) {
	if entry, ok := curated[word]; ok {
		return entry, true
	}

	entry := Entry{
		Synonyms: synonymsOEWN[word],
		Antonyms: antonymsOEWN[word],
	}
	if len(entry.Synonyms) == 0 && len(entry.Antonyms) == 0 {
		return Entry{}, false
	}
	return entry, true
}

// normalise converts an input word to the same form used by the dataset
// keys: lowercase, with surrounding whitespace trimmed. This lets callers
// look up "Happy", "HAPPY", or " happy " and still hit "happy" in the map.
func normalise(word string) string {
	return strings.ToLower(strings.TrimSpace(word))
}
