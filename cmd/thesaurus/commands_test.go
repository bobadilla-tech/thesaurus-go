package main

import (
	"bytes"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	thesaurus "github.com/bobadilla-tech/thesaurus-go"
)

func TestRunLookup_KnownWord(t *testing.T) {
	var buf bytes.Buffer

	err := runLookup("happy", &buf)

	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "Synonyms:")
	assert.Contains(t, out, "Antonyms:")
}

func TestRunLookup_UnknownWord(t *testing.T) {
	var buf bytes.Buffer

	err := runLookup("zzzznotaword", &buf)

	assert.Error(t, err, "expected an error for an unknown word")
	assert.Empty(t, buf.String(), "should not print anything for an unknown word")
}

func TestRunPrefix_KnownPrefix(t *testing.T) {
	var buf bytes.Buffer

	err := runPrefix("happ", &buf)

	require.NoError(t, err)
	lines := strings.Fields(buf.String())
	assert.NotEmpty(t, lines)
	for _, w := range lines {
		assert.True(t, strings.HasPrefix(w, "happ"), "%q does not start with 'happ'", w)
	}
}

func TestRunPrefix_UnknownPrefix(t *testing.T) {
	var buf bytes.Buffer

	err := runPrefix("zzzznotaprefix", &buf)

	require.NoError(t, err, "an unmatched prefix is not an error")
	assert.Empty(t, buf.String())
}

func TestRunCount_MatchesLibrary(t *testing.T) {
	var buf bytes.Buffer

	err := runCount(&buf)

	require.NoError(t, err)
	got, convErr := strconv.Atoi(strings.TrimSpace(buf.String()))
	require.NoError(t, convErr)
	assert.Equal(t, thesaurus.Count(), got)
}
