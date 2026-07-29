package thesaurus

import (
	"bytes"
	"compress/gzip"
	"io"
	"testing"
)

// benchWords is a realistic mixed sample: a curated hit, OEWN-only hits
// (random sample of real dataset words), and a miss — cycled through so
// benchmarks aren't measuring a single always-cached lookup.
var benchWords = []string{
	"happy", // curated hit
	"violet", "clappers", "amphibian", "hemoptysis", "galactic",
	"feldspar", "crassness", "chaplaincy", "specialism", "cadent",
	"tinsmith", "parasite", "antipathetic", "anomalousness", "carinated",
	"fagot", "fluoridation", "sandaled", "trouble", "anapest", // OEWN-only hits
	"zzzznotaword", // miss
}

func lookupFST(word string) (Entry, bool) {
	word = normalise(word)
	if entry, ok := curatedData[word]; ok {
		return entry, true
	}
	entry := Entry{}
	if syn, ok := synonymsFSTData.Get(word); ok {
		entry.Synonyms = syn
	}
	if ant, ok := antonymsFSTData.Get(word); ok {
		entry.Antonyms = ant
	}
	if len(entry.Synonyms) == 0 && len(entry.Antonyms) == 0 {
		return Entry{}, false
	}
	return entry, true
}

func BenchmarkLookup_Map(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Lookup(benchWords[i%len(benchWords)])
	}
}

func BenchmarkLookup_FST(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		lookupFST(benchWords[i%len(benchWords)])
	}
}

func BenchmarkDatasetLoad_Map(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = mustLoadGzipJSON(synonymsCompressed)
	}
}

func BenchmarkDatasetLoad_FST(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = newFSTIndex(synonymsFSTRaw, synonymsValuesCompressed)
	}
}

// BenchmarkArtifactDecompress isolates just the gzip decompression step (no
// JSON unmarshal / FST parse) for the synonyms artifact under each scheme,
// since the FST variant's values blob still needs gzip+JSON decode just like
// the map does today.
func BenchmarkArtifactDecompress_MapJSON(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		gz, _ := gzip.NewReader(bytes.NewReader(synonymsCompressed))
		_, _ = io.Copy(io.Discard, gz)
	}
}

func BenchmarkArtifactDecompress_FSTValues(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		gz, _ := gzip.NewReader(bytes.NewReader(synonymsValuesCompressed))
		_, _ = io.Copy(io.Discard, gz)
	}
}
