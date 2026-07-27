# Thesaurus Go

A lightweight Go package for English thesaurus lookups using an embedded,
Open English WordNet (OEWN) dataset.

[![Go Reference](https://pkg.go.dev/badge/github.com/bobadilla-tech/thesaurus-go.svg)](https://pkg.go.dev/github.com/bobadilla-tech/thesaurus-go)
[![codecov](https://codecov.io/gh/bobadilla-tech/thesaurus-go/graph/badge.svg)](https://codecov.io/gh/bobadilla-tech/thesaurus-go)

## Features

- 🚀 **Zero runtime dependencies** — pure Go, standard library only

- 📦 **Embedded dataset** — Open English WordNet (OEWN) compiled into the binary
  via `go:embed`; no external files required at runtime

- 🗜️ **Compressed embedded data** — synonym and antonym indexes are stored as
  gzip-compressed JSON to reduce binary size

- 📚 **~30,000 English words** — far broader coverage than small curated
  thesaurus lists

- ✨ **Simple API** — one function: `Lookup(word string)`

- 🔤 **Case-insensitive lookups** — input is automatically normalized by
  trimming whitespace and converting to lowercase

- 🧪 **Well tested** — lookup behavior, merge logic, normalization, and unknown
  words are all covered

## Installation

```bash
go get github.com/bobadilla-tech/thesaurus-go
```

## Usage

```go
package main

import (
	"fmt"
	"log"

	thesaurus "github.com/bobadilla-tech/thesaurus-go"
)

func main() {
	entry, ok := thesaurus.Lookup("happy")
	if !ok {
		log.Fatal("word not found")
	}

	fmt.Println("Synonyms:", entry.Synonyms)
	fmt.Println("Antonyms:", entry.Antonyms)
}
```

Output:

```
Synonyms: [glad joyful felicitous cheerful ...]
Antonyms: [unhappy sad]
```

## API

```go
func Lookup(word string) (Entry, bool)
```

Returns:

```go
type Entry struct {
	Synonyms []string
	Antonyms []string
}
```

If the word does not exist in the dataset:

```go
entry, ok := thesaurus.Lookup("xyznotaword")

// ok == false
```

## How It Works

1. **Normalize** — input words are trimmed and converted to lowercase.
2. **Curated lookup** — a small hand-maintained dataset is checked first.
   Curated entries always take precedence over OEWN.
3. **OEWN fallback** — if the word is not present in the curated dataset,
   synonyms and antonyms are resolved from the embedded Open English WordNet
   indexes.
4. **Embedded data** — the generated JSON indexes are gzip-compressed and
   embedded into the binary using `go:embed`.

## Regenerating the Dataset

The repository includes a build-time parser located in `cmd/wnparser`.

It parses the official Open English WordNet XML (GWN-LMF format) and generates
the compressed JSON files embedded by the package.

Example:

```bash
go run ./cmd/wnparser \
  -input english-wordnet-2025.xml \
  -output-dir ./pkg/thesaurus
```

This command generates:

- `synonyms_oewn.json.gz`
- `antonyms_oewn.json.gz`

These files are consumed automatically through `go:embed`.

## Testing

Run the tests:

```bash
go test -v ./...
```

## Used in Production

This package powers the
[Thesaurus](https://requiems.xyz/en/apis/thesaurus)
endpoint on
[Requiems API](https://requiems.xyz),
an all-in-one backend API for SaaS products (auth, fraud detection,
payments intelligence, global data, data integrity).

- Full API docs: https://requiems.xyz/en/apis
- Systems overview: https://requiems.xyz/en/systems

Need more language tooling? Requiems API's **Text & Language** system also
provides dictionary lookups, spell checking, language detection, sentiment
analysis, text similarity, and more through a single API.

## License

This project is licensed under the MIT License.

## Credits

- Word data derived from **Open English WordNet (OEWN)**.
- Open English WordNet is developed by the Global WordNet Association and
  distributed under the CC BY 4.0 License.