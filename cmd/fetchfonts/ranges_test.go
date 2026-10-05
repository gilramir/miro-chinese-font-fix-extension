package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRanges(t *testing.T) {
	r, err := parseRanges("U+4e00-9fff, U+3001")
	require.NoError(t, err)
	assert.Equal(t, []span{{0x4E00, 0x9FFF}, {0x3001, 0x3001}}, r)

	for _, bad := range []string{"4E00", "U+9FFF-4E00", "U+ZZ"} {
		_, err := parseRanges(bad)
		assert.Error(t, err, bad)
	}
}

func TestIntersectKeepsHanNotKana(t *testing.T) {
	// a slice with Latin, kana and some characters
	slice := mustRanges("U+0-FF, U+3001-3002, U+3041-3096, U+4E00-4E10, U+4E11-4E20")
	assert.Equal(t, "U+3001-3002, U+4E00-4E20", formatRanges(intersect(slice, han)))
	assert.Empty(t, intersect(mustRanges("U+0-FF, U+30A0-30FF"), han))
}
