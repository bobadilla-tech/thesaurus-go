# FST (vellum) as a lookup backend — spike

## Context

thesaurus-go's entire job is `word -> []string` point lookups: `Lookup(word)`
checks a small curated map (48 entries) then falls back to two OEWN-derived maps
loaded from gzip-compressed JSON via `//go:embed` — 43,370 synonym entries and
6,036 antonym entries, decoded into `map[string][]string` at `init()`
(`oewn_embed.go`).

`github.com/blevesearch/vellum` is a Go FST (finite state transducer)
implementation. FSTs compress shared key prefixes into a compact automaton and
avoid per-key heap allocations that a 43k-entry Go map incurs, so it was worth
checking whether it would beat the hashmap approach on the metrics that matter
for a lookup-only library: per-call latency/allocs, `init()` cost, and embedded
artifact size. No prior attempt at this existed in the repo.

## Approach

Build a working FST-backed prototype alongside the existing map-based path (not
wired into `Lookup()`), benchmark both with `go test -bench`, and decide based
on real numbers rather than assumption:

- A small build-time tool (`cmd/fstbuild`) reads the already-generated OEWN
  gzip-JSON maps, sorts keys (vellum requires lexicographic insertion order),
  and emits a vellum FST (`word -> index`) plus a parallel gzip-JSON values
  array (`index -> []string`), since vellum values are a single `uint64` and
  can't hold the string lists directly.
- An `fstIndex` type in the thesaurus package wraps the loaded FST + values
  array behind the same `Get(word) ([]string, bool)` shape the map path uses, so
  both are interchangeable in benchmarks.
- Benchmarks cover: steady-state lookup (the dominant real workload), isolated
  `FST.Get` cost with a pre-built key (to rule out `[]byte` conversion overhead
  as a confound), one-time dataset-load cost, and embedded artifact size —
  compressed the same way on both sides for a fair comparison, not raw FST bytes
  vs. gzip-JSON.
- Correctness gate before trusting any number: FST-backed lookups must return
  identical results to the map path across a real-word sample.
- If the FST path won clearly on the metric that matters (lookup, since that's
  what this library does millions of times more often than it inits), wire it
  into `Lookup()` behind the existing API. If not, keep the map path and
  document why, so the question doesn't get re-litigated from scratch next time
  someone eyes a new data structure for this package.

## Final notes

**Not adopted.** FST lost decisively on the metric that actually matters:

- Steady-state lookup: map 22.0 ns/op, 0 allocs vs. FST 277.0 ns/op, 1 alloc —
  map wins 12.6x.
- Isolated `FST.Get` (same pre-built key, no conversion overhead): still 160.5
  ns/op vs. map's 4.5 ns/op. Root cause traced into vellum itself: `FST.Get`
  calls the unexported `f.get(input, nil)`, and the `nil` forces a fresh state
  allocation for the root node on every call — the library has an internal
  pooling hook it doesn't expose publicly. Even a hypothetical fork that patched
  this out would only remove ~20-40ns; the rest is structural (byte-by-byte
  automaton traversal with data-dependent transitions vs. a single hash pass +
  1-2 probes on Go's map — wider on Go 1.24+'s swiss-table map implementation,
  which this repo's Go 1.26 uses).
- Dataset load (`init()`-equivalent, one-time per process): FST ~35% faster
  (20.7ms vs 31.9ms) — real, but irrelevant next to the lookup cost, since init
  runs once and lookups run constantly.
- Embedded artifact size, gzip applied fairly to both: synonyms ~1% smaller with
  FST, antonyms ~19% _larger_. Splitting keys (FST) from values (a separate gzip
  blob) loses the cross-compression gzip gets from squeezing keys+values
  together in one JSON stream.

Kept the full spike (tool, prototype, benchmarks, this writeup) on branch
`spike/vellum-fst`, not merged to `main`. `main`'s README links to it
(`## Benchmarks & Design Notes`) so the reasoning stays discoverable without
carrying the vellum dependency in the root module. Verdict: FSTs earn their keep
for out-of-core/mmap datasets too large for RAM, prefix/fuzzy/range queries, or
key counts in the millions — none of which apply to a 43k-word in-memory lookup
table where point lookup is the entire job.
