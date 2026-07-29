package thesaurus

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"io"
	"sort"
)

// dataset/curated.json is hand-maintained directly in this repo — see it
// for the small, manually curated word list that takes priority over OEWN.

type Entry struct {
	Synonyms []string
	Antonyms []string
}

//go:embed dataset/curated.json
var curatedRaw []byte

//go:embed dataset/synonyms_oewn.json.gz
var synonymsCompressed []byte

//go:embed dataset/antonyms_oewn.json.gz
var antonymsCompressed []byte

var curatedData map[string]Entry
var synonymsOEWNData map[string][]string
var antonymsOEWNData map[string][]string

// allWords is every word Lookup can answer for, deduplicated and sorted
// once at startup so WordsWithPrefix/Count/AllWords don't have to redo the
// merge (and don't have to reason about curated vs OEWN precedence, since
// this is just the set of known words, not their entries).
var allWords []string

func init() {
	if err := json.Unmarshal(curatedRaw, &curatedData); err != nil {
		panic(err)
	}

	synonymsOEWNData = mustLoadGzipJSON(synonymsCompressed)
	antonymsOEWNData = mustLoadGzipJSON(antonymsCompressed)

	allWords = buildAllWords(curatedData, synonymsOEWNData, antonymsOEWNData)
}

func buildAllWords(curated map[string]Entry, synonymsOEWN, antonymsOEWN map[string][]string) []string {
	seen := make(map[string]struct{}, len(curated)+len(synonymsOEWN)+len(antonymsOEWN))
	
	for word := range curated {
		seen[word] = struct{}{}
	}
	
	for word := range synonymsOEWN {
		seen[word] = struct{}{}
	}
	
	for word := range antonymsOEWN {
		seen[word] = struct{}{}
	}

	words := make([]string, 0, len(seen))
	
	for word := range seen {
		words = append(words, word)
	}
	
	sort.Strings(words)
	return words
}

// mustLoadGzipJSON decompresses a gzip-compressed JSON payload embedded via
// //go:embed and unmarshals it into a map. Panics on failure: this only runs
// once at process startup (via init), before the service accepts any
// traffic, so failing loudly here is preferable to starting with a corrupt
// or missing dataset.
func mustLoadGzipJSON(compressed []byte) map[string][]string {
	gz, err := gzip.NewReader(bytes.NewReader(compressed))

	if err != nil {
		panic(err)
	}

	defer gz.Close()

	decompressed, err := io.ReadAll(gz)
	
	if err != nil {
		panic(err)
	}

	var data map[string][]string
	
	if err := json.Unmarshal(decompressed, &data); err != nil {
		panic(err)
	}

	return data
}
