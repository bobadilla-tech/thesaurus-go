// Package main implements fstbuild, a BUILD-TIME spike tool kept separate
// from pkg/thesaurus. It reads the already-generated OEWN gzip-JSON maps
// (word -> []string) and re-emits the same data as a vellum FST (word ->
// index) plus a parallel gzip-JSON values array ([][]string, indexed by that
// uint64), so the two representations can be benchmarked against each other.
//
// This does not reparse the raw WordNet XML — it reuses the maps wnparser
// already produced, since the goal is an apples-to-apples comparison of
// lookup backends over identical data, not a second data pipeline.
//
// Usage:
//
//	go run ./cmd/fstbuild -input ./pkg/thesaurus/synonyms_oewn.json.gz -output-prefix ./pkg/thesaurus/synonyms_oewn
package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/blevesearch/vellum"
)

func loadGzipJSON(path string) map[string][]string {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		panic(err)
	}
	defer gz.Close()

	raw, err := io.ReadAll(gz)
	if err != nil {
		panic(err)
	}

	var data map[string][]string
	if err := json.Unmarshal(raw, &data); err != nil {
		panic(err)
	}
	return data
}

func writeGzipJSON(path string, data any) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	defer gz.Close()

	if err := json.NewEncoder(gz).Encode(data); err != nil {
		panic(err)
	}
}

func main() {
	inputPath := flag.String("input", "", "path to source .json.gz (word -> []string)")
	outputPrefix := flag.String("output-prefix", "", "output files are <prefix>.fst and <prefix>.values.json.gz")
	flag.Parse()

	if *inputPath == "" || *outputPrefix == "" {
		fmt.Println("usage: fstbuild -input <path.json.gz> -output-prefix <prefix>")
		os.Exit(1)
	}

	data := loadGzipJSON(*inputPath)

	words := make([]string, 0, len(data))
	for word := range data {
		words = append(words, word)
	}
	sort.Strings(words) // vellum requires lexicographic insertion order

	values := make([][]string, len(words))

	var fstBuf bytes.Buffer
	builder, err := vellum.New(&fstBuf, nil)
	if err != nil {
		panic(err)
	}
	for i, word := range words {
		if err := builder.Insert([]byte(word), uint64(i)); err != nil {
			panic(err)
		}
		values[i] = data[word]
	}
	if err := builder.Close(); err != nil {
		panic(err)
	}

	fstPath := *outputPrefix + ".fst"
	if err := os.WriteFile(fstPath, fstBuf.Bytes(), 0o644); err != nil {
		panic(err)
	}

	valuesPath := *outputPrefix + ".values.json.gz"
	writeGzipJSON(valuesPath, values)

	fmt.Printf("done: %s (%d bytes), %s, %d words\n", fstPath, fstBuf.Len(), valuesPath, len(words))
}
