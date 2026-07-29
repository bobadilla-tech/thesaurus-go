package main

// Provider produces synonym/antonym maps (word -> []string) from a raw
// dataset file at path. Adding a new data source means implementing this
// interface and registering it in providers — main() itself never needs to
// change.
type Provider interface {
	Parse(path string) (synonyms, antonyms map[string][]string, err error)
}

// providers is the strategy registry, keyed by the name passed via -provider.
var providers = map[string]Provider{
	"oewn": oewnProvider{},
}
