// Package main implements datasetbuild, a BUILD-TIME preprocessor kept
// separate from the thesaurus package. It is not imported by production
// code and is not part of the API binary — it runs manually (or via
// `go generate`) only when the dataset needs to be regenerated (a new OEWN
// release, a correction to the curated antonym layer).
//
// It reads a raw dataset through a registered Provider (currently only
// "oewn", Open English WordNet's GWN-LMF XML — see provider.go and oewn.go)
// and generates two gzip-compressed JSON files (word -> []string), ready to
// be embedded into the thesaurus package with no extra compression step.
//
// Usage:
//
//	go run ./cmd/datasetbuild -input english-wordnet-2025.xml -output-dir ./dataset
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	inputPath := flag.String("input", "english-wordnet-2025.xml", "path to the raw dataset file")
	outputDir := flag.String("output-dir", ".", "directory to write synonyms.json.gz and antonyms.json.gz to")
	providerName := flag.String("provider", "oewn", "data source to parse (registered in provider.go)")
	flag.Parse()

	provider, ok := providers[*providerName]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown provider %q\n", *providerName)
		os.Exit(1)
	}

	synonyms, antonyms, err := provider.Parse(*inputPath)
	if err != nil {
		panic(err)
	}

	synPath := filepath.Join(*outputDir, "synonyms.json.gz")
	antPath := filepath.Join(*outputDir, "antonyms.json.gz")

	if err := writeJSONGzip(synPath, synonyms); err != nil {
		panic(err)
	}
	if err := writeJSONGzip(antPath, antonyms); err != nil {
		panic(err)
	}

	fmt.Printf("done: %s (%d words with synonyms), %s (%d words with antonyms)\n",
		synPath, len(synonyms), antPath, len(antonyms))
	fmt.Println("both files are gzip-compressed and ready to //go:embed in the thesaurus package")
}
