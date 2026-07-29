package thesaurus

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"io"

	"github.com/blevesearch/vellum"
)

// fst_embed.go is a benchmark spike (see cmd/fstbuild): an alternative,
// FST-backed representation of the same OEWN data already embedded via
// oewn_embed.go. Not wired into Lookup() — see lookup_bench_test.go for the
// comparison this exists to support.

//go:embed synonyms_oewn.fst
var synonymsFSTRaw []byte

//go:embed synonyms_oewn.values.json.gz
var synonymsValuesCompressed []byte

//go:embed antonyms_oewn.fst
var antonymsFSTRaw []byte

//go:embed antonyms_oewn.values.json.gz
var antonymsValuesCompressed []byte

// fstIndex adapts a vellum FST (word -> index) plus its side values array
// (index -> []string) to the wordIndex interface, so it's a drop-in
// alternative to mapIndex for lookup().
type fstIndex struct {
	fst    *vellum.FST
	values [][]string
}

func (idx fstIndex) Get(word string) ([]string, bool) {
	i, found, err := idx.fst.Get([]byte(word))
	if err != nil || !found {
		return nil, false
	}
	return idx.values[i], true
}

// mustLoadGzipJSONSlice mirrors mustLoadGzipJSON (oewn_embed.go) but decodes
// into the [][]string shape fstIndex's values side-array uses.
func mustLoadGzipJSONSlice(compressed []byte) [][]string {
	gz, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		panic(err)
	}
	defer gz.Close()

	decompressed, err := io.ReadAll(gz)
	if err != nil {
		panic(err)
	}

	var data [][]string
	if err := json.Unmarshal(decompressed, &data); err != nil {
		panic(err)
	}
	return data
}

func newFSTIndex(fstBytes, valuesCompressed []byte) fstIndex {
	fst, err := vellum.Load(fstBytes)
	if err != nil {
		panic(err)
	}
	return fstIndex{
		fst:    fst,
		values: mustLoadGzipJSONSlice(valuesCompressed),
	}
}

var synonymsFSTData fstIndex
var antonymsFSTData fstIndex

func init() {
	synonymsFSTData = newFSTIndex(synonymsFSTRaw, synonymsValuesCompressed)
	antonymsFSTData = newFSTIndex(antonymsFSTRaw, antonymsValuesCompressed)
}
