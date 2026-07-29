package main

import (
	"compress/gzip"
	"encoding/json"
	"os"
)

// writeJSONGzip encodes data as JSON and writes it gzip-compressed directly
// to path (expected to end in .json.gz). This avoids a separate manual
// compression step — the tool always produces the embed-ready artifact.
func writeJSONGzip(path string, data map[string][]string) error {
	f, err := os.Create(path)

	if err != nil {
		return err
	}
	
	defer f.Close()

	gz := gzip.NewWriter(f)
	defer gz.Close()

	enc := json.NewEncoder(gz)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}
