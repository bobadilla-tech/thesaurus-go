// Package main implements wnparser, a BUILD-TIME tool kept separate from
// pkg/thesaurus. It is not imported by production code and is not part of
// the API binary — it runs manually (or via `go generate`) only when the
// dataset needs to be regenerated (a new OEWN release, a correction to the
// curated antonym layer).
//
// It reads the raw Open English WordNet (GWN-LMF) XML and generates two
// gzip-compressed JSON files (word -> []string), ready to be embedded into
// pkg/thesaurus via //go:embed with no extra compression step.
//
// Usage:
//
//	go run ./cmd/wnparser -input english-wordnet-2025.xml -output-dir ./pkg/thesaurus
package main

import (
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ---- Structs mapping the GWN-LMF XML schema ----
// Reference for the real structure (LexicalResource > Lexicon > LexicalEntry/Synset):
// https://globalwordnet.github.io/schemas/

type LexicalResource struct {
	XMLName xml.Name `xml:"LexicalResource"`
	Lexicon Lexicon  `xml:"Lexicon"`
}

type Lexicon struct {
	LexicalEntries []LexicalEntry `xml:"LexicalEntry"`
	Synsets        []Synset       `xml:"Synset"`
}

type LexicalEntry struct {
	ID     string  `xml:"id,attr"`
	Lemma  Lemma   `xml:"Lemma"`
	Senses []Sense `xml:"Sense"`
}

type Lemma struct {
	WrittenForm  string `xml:"writtenForm,attr"`
	PartOfSpeech string `xml:"partOfSpeech,attr"`
}

type Sense struct {
	ID        string          `xml:"id,attr"`
	SynsetID  string          `xml:"synset,attr"`
	Relations []SenseRelation `xml:"SenseRelation"`
}

type SenseRelation struct {
	RelType string `xml:"relType,attr"`
	Target  string `xml:"target,attr"` // points to another Sense.ID
}

type Synset struct {
	ID           string `xml:"id,attr"`
	PartOfSpeech string `xml:"partOfSpeech,attr"`
	Definition   string `xml:"Definition"`
}

// isSingleWord reports whether a lemma is a single word, excluding
// multi-word expressions and hyphenated compounds (e.g. "happy-go-lucky",
// "devil-may-care"). Phrase support is out of scope for now — see spec.
func isSingleWord(lemma string) bool {
	return !strings.ContainsAny(lemma, " -")
}

// ---- In-memory indexes built from the parsed XML ----

type Indexes struct {
	wordToSynsets  map[string][]string // "happy" -> [synset_id, ...]
	synsetMembers  map[string][]string // synset_id -> ["happy", "felicitous", ...]
	senseIDToLemma map[string]string   // "oewn-unhappy__3.00.00.." -> "unhappy"
	senseAntonyms  map[string][]string // sense_id -> [target_sense_id, ...] (only relType=="antonym")
}

func buildIndexes(lex Lexicon) Indexes {
	idx := Indexes{
		wordToSynsets:  map[string][]string{},
		synsetMembers:  map[string][]string{},
		senseIDToLemma: map[string]string{},
		senseAntonyms:  map[string][]string{},
	}

	// A single pass over all LexicalEntry elements builds all four indexes.
	for _, entry := range lex.LexicalEntries {
		lemma := strings.ToLower(entry.Lemma.WrittenForm)

		for _, sense := range entry.Senses {
			idx.wordToSynsets[lemma] = append(idx.wordToSynsets[lemma], sense.SynsetID)
			idx.synsetMembers[sense.SynsetID] = append(idx.synsetMembers[sense.SynsetID], lemma)
			idx.senseIDToLemma[sense.ID] = lemma

			for _, rel := range sense.Relations {
				if rel.RelType == "antonym" {
					idx.senseAntonyms[sense.ID] = append(idx.senseAntonyms[sense.ID], rel.Target)
				}
			}
		}
	}

	return idx
}

// Synonyms: word -> its senses -> the synsets they point to ->
// other members of those synsets (excluding the word itself).
// Multi-word members are filtered out (see isSingleWord).
func (idx Indexes) Synonyms(word string) []string {
	word = strings.ToLower(word)
	seen := map[string]bool{}
	var result []string

	for _, synsetID := range idx.wordToSynsets[word] {
		for _, member := range idx.synsetMembers[synsetID] {
			if member != word && isSingleWord(member) && !seen[member] {
				seen[member] = true
				result = append(result, member)
			}
		}
	}
	return result
}

// Antonyms: word -> its own senses -> SenseRelation[antonym].target (another
// sense ID) -> resolve that sense ID back to the lemma that owns it.
// Multi-word targets are filtered out (see isSingleWord).
func (idx Indexes) Antonyms(word string) []string {
	word = strings.ToLower(word)
	seen := map[string]bool{}
	var result []string

	// Walk every Sense owned by this word (not by synset, since antonymy
	// lives at the sense level, not the synset level).
	for senseID, lemma := range idx.senseIDToLemma {
		if lemma != word {
			continue
		}
		for _, targetSenseID := range idx.senseAntonyms[senseID] {
			if targetLemma, ok := idx.senseIDToLemma[targetSenseID]; ok && isSingleWord(targetLemma) && !seen[targetLemma] {
				seen[targetLemma] = true
				result = append(result, targetLemma)
			}
		}
	}
	return result
}

// allLemmas returns every unique word present in the lexicon, so the full
// dataset can be generated (not just single-word lookups).
func allLemmas(lex Lexicon) []string {
	seen := map[string]bool{}
	var lemmas []string
	for _, entry := range lex.LexicalEntries {
		lemma := strings.ToLower(entry.Lemma.WrittenForm)
		if !seen[lemma] {
			seen[lemma] = true
			lemmas = append(lemmas, lemma)
		}
	}
	return lemmas
}

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

func main() {
	inputPath := flag.String("input", "english-wordnet-2025.xml", "path to the raw OEWN XML (GWN-LMF)")
	outputDir := flag.String("output-dir", ".", "directory to write synonyms.json and antonyms.json to")
	flag.Parse()

	data, err := os.ReadFile(*inputPath)
	if err != nil {
		panic(err)
	}

	fmt.Println("parsing XML...")
	var lr LexicalResource
	if err := xml.Unmarshal(data, &lr); err != nil {
		panic(err)
	}

	fmt.Println("building indexes...")
	idx := buildIndexes(lr.Lexicon)

	lemmas := allLemmas(lr.Lexicon)
	fmt.Printf("generating entries for %d unique words...\n", len(lemmas))

	synonyms := map[string][]string{}
	antonyms := map[string][]string{}

	for _, lemma := range lemmas {
		if !isSingleWord(lemma) {
			continue
		}
		if syns := idx.Synonyms(lemma); len(syns) > 0 {
			synonyms[lemma] = syns
		}
		if ants := idx.Antonyms(lemma); len(ants) > 0 {
			antonyms[lemma] = ants
		}
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
	fmt.Println("both files are gzip-compressed and ready to //go:embed in pkg/thesaurus")
}
