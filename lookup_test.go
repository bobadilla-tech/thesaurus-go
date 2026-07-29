package thesaurus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fixtures used across tests. Kept separate from the real embedded data
// (curated.json, synonyms_oewn.json.gz, antonyms_oewn.json.gz) on purpose:
// these tests exercise the merge logic itself, not the production dataset.
var (
	fixtureCurated = map[string]Entry{
		"happy": {
			Synonyms: []string{"joyful", "cheerful"},
			Antonyms: []string{"sad", "unhappy"},
		},
	}

	// "happy" also appears here, with DIFFERENT values than the curated
	// fixture above, specifically to prove curated wins over OEWN.
	fixtureSynonymsOEWN = map[string][]string{
		"happy": {"felicitous", "glad"},
		"quick": {"fast", "swift", "speedy"},
	}

	fixtureAntonymsOEWN = map[string][]string{
		"happy": {"unhappy"}, // OEWN's own (different) antonym set for happy
		"quick": {"slow"},
	}
)

// Acceptance criterion: known word (curated) returns curated data.
func TestLookup_KnownWordCurated(t *testing.T) {
	entry, ok := lookup("happy", fixtureCurated, fixtureSynonymsOEWN, fixtureAntonymsOEWN)

	require.True(t, ok, "expected ok=true for a known curated word")

	wantSynonyms := []string{"joyful", "cheerful"}
	wantAntonyms := []string{"sad", "unhappy"}
	assert.Equal(t, wantSynonyms, entry.Synonyms)
	assert.Equal(t, wantAntonyms, entry.Antonyms)
}

// Acceptance criterion: known word (OEWN only, no curated entry) returns
// OEWN data via fallback.
func TestLookup_KnownWordOEWNFallback(t *testing.T) {
	entry, ok := lookup("quick", fixtureCurated, fixtureSynonymsOEWN, fixtureAntonymsOEWN)

	require.True(t, ok, "expected ok=true for a word known only to OEWN")
	assert.Equal(t, []string{"fast", "swift", "speedy"}, entry.Synonyms)
	assert.Equal(t, []string{"slow"}, entry.Antonyms)
}

// Acceptance criterion: unknown word (absent from both sources) returns
// ok=false and a zero-value Entry.
func TestLookup_UnknownWord(t *testing.T) {
	entry, ok := lookup("zzzznotaword", fixtureCurated, fixtureSynonymsOEWN, fixtureAntonymsOEWN)

	assert.False(t, ok, "expected ok=false for an unknown word")
	assert.Equal(t, Entry{}, entry)
}

// Acceptance criterion: correct override application. "happy" exists in
// BOTH curated and OEWN fixtures with different values — curated must win
// in full (not merged field-by-field with OEWN).
func TestLookup_CuratedOverridesOEWN(t *testing.T) {
	entry, ok := lookup("happy", fixtureCurated, fixtureSynonymsOEWN, fixtureAntonymsOEWN)

	require.True(t, ok)
	assert.Equal(t, fixtureCurated["happy"], entry, "curated entry should win entirely, not merge field-by-field")
	assert.NotContains(t, entry.Synonyms, "felicitous", "override leaked an OEWN-only synonym")
	assert.NotContains(t, entry.Synonyms, "glad", "override leaked an OEWN-only synonym")
}

// A word present in neither curated nor OEWN synonyms, but present in OEWN
// antonyms only, should still return ok=true with an empty Synonyms slice.
func TestLookup_PartialOEWNData(t *testing.T) {
	synOnly := map[string][]string{}
	antOnly := map[string][]string{"lonely": {"accompanied"}}

	entry, ok := lookup("lonely", map[string]Entry{}, synOnly, antOnly)

	require.True(t, ok, "expected ok=true when only antonyms are present")
	assert.Empty(t, entry.Synonyms)
	assert.Equal(t, []string{"accompanied"}, entry.Antonyms)
}

// Sanity check against the real embedded dataset: at least one known
// curated word must resolve end-to-end through the public Lookup function,
// confirming //go:embed + init() actually loaded curated.json correctly.
func TestLookup_RealEmbeddedData(t *testing.T) {
	entry, ok := Lookup("Happy") // mixed case, exercises normalise() too

	require.True(t, ok, "expected ok=true for 'Happy' against the real embedded curated dataset")
	assert.NotEmpty(t, entry.Synonyms)
}
