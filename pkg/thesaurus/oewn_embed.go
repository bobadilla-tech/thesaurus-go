package thesaurus

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"io"
)

// curated.json is hand-maintained directly in this repo — see it for the
// small, manually curated word list that takes priority over OEWN.

type Entry struct {
	Synonyms []string
	Antonyms []string
}

//go:embed curated.json
var curatedRaw []byte

//go:embed synonyms_oewn.json.gz
var synonymsCompressed []byte

//go:embed antonyms_oewn.json.gz
var antonymsCompressed []byte

var curatedData map[string]Entry
var synonymsOEWNData map[string][]string
var antonymsOEWNData map[string][]string

func init() {

	if err := json.Unmarshal(curatedRaw, &curatedData); err != nil {
		panic(err)
	}

	synonymsOEWNData = mustLoadGzipJSON(synonymsCompressed)
	antonymsOEWNData = mustLoadGzipJSON(antonymsCompressed)
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
