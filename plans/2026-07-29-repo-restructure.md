# Root package, dataset/ directory, new lookup API, provider strategy pattern

## Context

Several unrelated rough edges had accumulated in the repo:

- The package lived at `pkg/thesaurus`, an extra directory level for a
  single-package library whose module is already
  `github.com/bobadilla-tech/thesaurus-go`. The README's own usage example
  already imported it as `thesaurus "github.com/bobadilla-tech/thesaurus-go"` —
  as if the package were at the module root — which was actually wrong given the
  real `/pkg/thesaurus` layout at the time.
- The README undersold the dataset ("~30,000 words" against a real union of
  45,695 lookupable words across curated + OEWN) and used emoji bullets.
- The API was lookup-only (`Lookup(word)`); there was no way to enumerate or
  search the word list, something any consumer building autocomplete or stats on
  top would want.
- `cmd/wnparser/main.go` was a single 239-line file hardcoding OEWN XML parsing
  directly into `main()` — no seam for a second data source without rewriting
  `main()`.

This was a deliberate breaking change to the import path. Confirmed with the
user before starting: acceptable now, the only known consumer (Requiems API)
gets updated separately. A `v1.0.0` git tag already existed on the repo; agreed
to tag a new major version after merge to signal the break through normal
semver, without going as far as an explicit `/v2` module path suffix (small
library, one real consumer, not worth the ceremony).

## Approach

Four pieces, done together since they touch overlapping files:

1. **Move the package to repo root.** `git mv` (not delete+recreate, to keep
   `git log --follow` working) everything out of `pkg/thesaurus/` to the root:
   `lookup.go`, `lookup_test.go`, `oewn_embed.go`, and the data files. No code
   changes needed — `package thesaurus` stays the same, only the path changes,
   making the import path `github.com/bobadilla-tech/thesaurus-go` directly.

2. **New API surface** (`words.go`): build a sorted, deduplicated list of every
   word either dataset can answer for, once, at `init()` time (merge the
   curated/synonym/antonym key sets, dedupe, sort — factored into a standalone
   `buildAllWords` function so it's unit-testable with fixtures the same way
   `lookup()` already was). Add `WordsWithPrefix(prefix)`
   (`sort.SearchStrings` + prefix scan over the sorted list), `Contains(word)`
   (delegates to `Lookup`), `Count()`, and `AllWords()` (defensive copy).

3. **README overhaul**: strip emoji, correct the word count to the real number,
   document the four new functions, fix the regeneration example's output path.

4. **`cmd/wnparser` provider strategy pattern**: split the single file into
   `provider.go` (a `Provider` interface —
   `Parse(path) (synonyms, antonyms map[string][]string, err error)` — plus a
   name-keyed registry), `oewn.go` (all the GWN-LMF XML structs and extraction
   logic, wrapped in an `oewnProvider` implementing `Provider`), `writer.go`
   (the gzip-JSON writer, unchanged), and a `main.go` reduced to flag parsing +
   registry lookup + orchestration. Added a `-provider` flag (default `oewn`) so
   a future second data source is a matter of implementing `Provider` and
   registering it, not touching `main()`.

Followed up in the same effort, once the user noticed the data files sitting
directly at repo root looked wrong next to the Go sources: moved `curated.json`,
`synonyms_oewn.json.gz`, `antonyms_oewn.json.gz` into a new `dataset/`
directory. `go:embed` paths in `oewn_embed.go` updated to `dataset/...` (embed
patterns can reference subdirectories of the source file's own directory, so no
other code needed to move), and the `cmd/wnparser` output-dir example updated to
`./dataset` to match.

## Final notes

Shipped as planned, no deviations. Verified with `go build ./...`,
`go vet ./...`, `go test ./...` (all packages, `-count=1` to bypass cache), plus
a throwaway spot-check program confirming `Count()` (45,695), `Contains`, and
`WordsWithPrefix` against the real embedded data — matched the earlier
`python3`-computed union count exactly.

One thing found along the way and deliberately **not** fixed here (out of this
plan's scope, flagged to the user instead): `cmd/wnparser` wrote
`synonyms.json.gz`/`antonyms.json.gz`, but the package embeds
`synonyms_oewn.json.gz`/`antonyms_oewn.json.gz` — a pre-existing filename
mismatch predating this work, meaning a fresh generator run needs a manual
rename before the output can be embedded. Still unresolved as of this writeup.

`cmd/wnparser` itself was renamed to `cmd/datasetbuild` in the very next
session, once it became clear the old name no longer fit a provider-based tool —
see `2026-07-29-cli-tool.md`.
