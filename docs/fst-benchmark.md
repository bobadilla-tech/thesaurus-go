# Spike: FST (vellum) as a lookup backend

Branch: `spike/vellum-fst` — not merged to `main`. Kept for reference only.

## Question

thesaurus-go's whole job is `word -> []string` point lookups over the OEWN
dataset (43,370 synonym entries, 6,036 antonym entries), currently a
`map[string][]string` decoded from gzip-compressed JSON at `init()`. Would
[`github.com/blevesearch/vellum`](https://github.com/blevesearch/vellum) (a Go
FST implementation) beat the map on lookup latency, init cost, or embedded
artifact size?

## Setup

- `cmd/fstbuild`: reads the existing `synonyms_oewn.json.gz` /
  `antonyms_oewn.json.gz`, sorts keys, builds a vellum FST (`word -> index`)
  plus a parallel gzip-JSON values array (`index -> []string`).
- `pkg/thesaurus/fst_embed.go`: embeds the FST + values artifacts, loads them
  via `vellum.Load`, exposes an `fstIndex.Get(word) ([]string, bool)` adapter.
- `pkg/thesaurus/lookup_bench_test.go`: benchmarks comparing the map path
  (`Lookup`) against the FST path (`lookupFST`) over a mixed real-word sample.
- Correctness verified: FST-backed lookups return identical results to the
  map for every word in the benchmark sample before trusting the numbers.

## Results

**Steady-state lookup (the dominant workload for this library):**

| | ns/op | B/op | allocs/op |
|---|---|---|---|
| Map (current) | 22.0 | 0 | 0 |
| FST (vellum) | 277.0 | 305 | 1 |

Map wins **12.6x**.

**Isolated `Get` cost**, same key, pre-built `[]byte` (rules out
`[]byte(word)` conversion as the cause):

| | ns/op | allocs/op |
|---|---|---|
| Map (`synonymsOEWNData["hemoptysis"]`) | 4.5 | 0 |
| FST (`fst.Get(key)`) | 160.5 | 1 |

**Root cause of the FST allocation**: `vellum.FST.Get` calls the unexported
`f.get(input, nil)` — the `nil` forces a fresh state allocation for the root
node on every call. The library has a pooling hook (`prealloc fstState`
param on the unexported `get`) but doesn't expose it publicly; a fork could
patch this out. It would not close the gap, though — removing the ~20-40ns
alloc still leaves FST at roughly **30-35x slower** than the map. That
remainder is structural: FST decodes the automaton byte-by-byte per key
(arc-table decode + transition per byte, each step data-dependent on the
last), where a Go map computes one hash over the whole key and does 1-2
probes. Go 1.24+'s swiss-table map implementation (this repo targets Go
1.26) makes that gap wider, not an artifact of an unoptimized comparison.

**Dataset load / `init()` cost (one-time per process):**

| | ms/op | B/op | allocs/op |
|---|---|---|---|
| Map | 31.9 | 20.6M | 323,602 |
| FST | 20.7 | 14.5M | 236,608 |

FST wins ~35% here, but this runs once per process lifetime, not per lookup.

**Embedded artifact size**, gzip applied to both representations for a fair
comparison (not raw FST vs. gzip-JSON):

| | current (gzip JSON) | FST(gzip) + values(gzip) | delta |
|---|---|---|---|
| synonyms | 659,684 B | 653,145 B | ~1% smaller |
| antonyms | 43,751 B | 52,267 B | ~19% **larger** |

Wash on the larger dataset, worse on the smaller one. Gzip already exploits
shared structure across keys *and* values in one JSON stream; splitting into
a separate FST + values blob loses that shared dictionary, and FST's own
prefix-sharing on keys doesn't make up the difference at this dataset size.

## Conclusion

**Not adopted.** FST loses on the metric that actually matters for this
library (steady-state point lookup, by 12.6x with an added allocation), wins
only on a one-time init cost, and is a wash-to-worse on embedded size. FSTs
earn their keep for out-of-core/mmap datasets too large for RAM, prefix/
fuzzy/range queries, or key counts in the millions where per-entry map
overhead dominates — none of which apply to a 43k-word in-memory lookup
table. `main` keeps the plain `map[string][]string` backend.
