package thesaurus

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Acceptance criterion: buildAllWords merges and dedupes across all three
// sources, and sorts the result.
func TestBuildAllWords_MergesDedupesSorts(t *testing.T) {
	curated := map[string]Entry{"zebra": {Synonyms: []string{"equine"}}}
	syn := map[string][]string{"zebra": {"equine"}, "apple": {"fruit"}}
	ant := map[string][]string{"apple": {"vegetable"}, "mango": {"nothing"}}

	got := buildAllWords(curated, syn, ant)

	want := []string{"apple", "mango", "zebra"}
	assert.Equal(t, want, got, "expected deduped, sorted union of all three sources")
}

// Acceptance criterion: real embedded dataset produces a non-trivial,
// sorted word list, and Count/AllWords agree with it.
func TestAllWords_RealEmbeddedData(t *testing.T) {
	words := AllWords()

	require.NotEmpty(t, words)
	assert.True(t, sort.StringsAreSorted(words), "AllWords should be sorted")
	assert.Equal(t, len(words), Count(), "Count should match len(AllWords())")
}

// Acceptance criterion: mutating the slice returned by AllWords must not
// affect the package's internal state (defensive copy).
func TestAllWords_ReturnsCopy(t *testing.T) {
	words := AllWords()
	require.NotEmpty(t, words)

	original := words[0]
	words[0] = "___mutated___"

	assert.Equal(t, original, AllWords()[0], "mutating the returned slice leaked into internal state")
}

// Acceptance criterion: a known prefix returns every matching word, sorted,
// and nothing that doesn't share the prefix.
func TestWordsWithPrefix_KnownPrefix(t *testing.T) {
	matches := WordsWithPrefix("happ")

	require.NotEmpty(t, matches)
	assert.True(t, sort.StringsAreSorted(matches))
	for _, w := range matches {
		assert.True(t, len(w) >= 4 && w[:4] == "happ", "%q does not start with 'happ'", w)
	}
	assert.Contains(t, matches, "happy")
}

// Acceptance criterion: prefix matching is case-insensitive and trims
// whitespace, same as Lookup's normalisation.
func TestWordsWithPrefix_Normalised(t *testing.T) {
	assert.Equal(t, WordsWithPrefix("happ"), WordsWithPrefix(" HAPP "))
}

// Acceptance criterion: a prefix no word starts with returns nil, not an
// error or panic.
func TestWordsWithPrefix_NoMatch(t *testing.T) {
	assert.Empty(t, WordsWithPrefix("zzzznotaprefix"))
}

// Acceptance criterion: Contains agrees with Lookup's ok return, for both a
// known and an unknown word.
func TestContains_AgreesWithLookup(t *testing.T) {
	_, lookupOK := Lookup("happy")
	assert.Equal(t, lookupOK, Contains("happy"))

	_, lookupOK = Lookup("zzzznotaword")
	assert.Equal(t, lookupOK, Contains("zzzznotaword"))
}
