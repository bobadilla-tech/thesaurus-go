# Thesaurus Go

[![Go Reference](https://pkg.go.dev/badge/github.com/bobadilla-tech/thesaurus-go.svg)](https://pkg.go.dev/github.com/bobadilla-tech/thesaurus-go)
[![codecov](https://codecov.io/gh/bobadilla-tech/thesaurus-go/graph/badge.svg)](https://codecov.io/gh/bobadilla-tech/thesaurus-go)

A lightweight Go package for English thesaurus lookups using an embedded, Open
English WordNet (OEWN) dataset.

## Features

- **Zero runtime dependencies**: pure Go, standard library only

- **Embedded dataset**: Open English WordNet (OEWN) compiled into the binary via
  `go:embed`; no external files required at runtime

- **Compressed embedded data**: synonym and antonym indexes are stored as
  gzip-compressed JSON to reduce binary size

- **45,000+ English words**: far broader coverage than small curated thesaurus
  lists

- **Simple API**: `Lookup`, `WordsWithPrefix`, `Contains`, `Count`, and
  `AllWords`

- **Case-insensitive lookups**: input is automatically normalized by trimming
  whitespace and converting to lowercase

- **Well tested**: lookup behavior, merge logic, normalization, and unknown
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

```go
func WordsWithPrefix(prefix string) []string
```

Every known word starting with `prefix` (case-insensitive), sorted:

```go
thesaurus.WordsWithPrefix("happ") // ["happen", "happening", "happily", "happy", ...]
```

```go
func Contains(word string) bool
```

Reports whether `word` has an entry, without needing the full `Entry`:

```go
thesaurus.Contains("happy") // true
```

```go
func Count() int
```

Total number of distinct words the package can look up (`45,695` as of the
current dataset).

```go
func AllWords() []string
```

Every known word, sorted. Returns a copy — safe to mutate.

## How It Works

1. **Normalize**: input words are trimmed and converted to lowercase.
2. **Curated lookup**: a small hand-maintained dataset is checked first. Curated
   entries always take precedence over OEWN.
3. \*_OEWN fallback_: if the word is not present in the curated dataset,
   synonyms and antonyms are resolved from the embedded Open English WordNet
   indexes.
4. **Embedded data**: the generated JSON indexes are gzip-compressed and
   embedded into the binary using `go:embed`.

## Regenerating the Dataset

The repository includes a build-time parser located in `cmd/wnparser`. Data
sources are pluggable through a `Provider` interface (see
`cmd/wnparser/provider.go`) — `oewn` is the only one implemented today, but
adding another means implementing `Provider` and registering it, with no changes
to `main()`.

The `oewn` provider parses the official Open English WordNet XML (GWN-LMF
format) and generates the compressed JSON files embedded by the package.

Example:

```bash
go run ./cmd/wnparser \
  -input english-wordnet-2025.xml \
  -output-dir ./dataset \
  -provider oewn
```

`-provider` defaults to `oewn`. This command generates:

- `synonyms_oewn.json.gz`
- `antonyms_oewn.json.gz`

Both live in `dataset/`, alongside the hand-maintained `curated.json`, and are
consumed automatically through `go:embed`.

## Benchmarks & Design Notes

- **FST (vellum) spike** — evaluated `github.com/blevesearch/vellum` as an
  alternative lookup backend. Rejected: 12.6x slower on steady-state point
  lookup (map: 22ns/op, 0 allocs vs FST: 277ns/op, 1 alloc), wash-to-worse on
  embedded artifact size. Full writeup and prototype on the
  [`spike/vellum-fst`](https://github.com/bobadilla-tech/thesaurus-go/tree/spike/vellum-fst)
  branch, not merged — see
  [`docs/fst-benchmark.md`](https://github.com/bobadilla-tech/thesaurus-go/blob/spike/vellum-fst/docs/fst-benchmark.md)
  on that branch.

## Testing

Run the tests:

```bash
go test -v ./...
```

## Used in Production

This package powers the [Thesaurus](https://requiems.xyz/en/apis/thesaurus)
endpoint on [Requiems API](https://requiems.xyz), an all-in-one backend API for
SaaS products (auth, fraud detection, payments intelligence, global data, data
integrity).

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
