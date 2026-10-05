package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// span is a range of code points, both ends in.
type span struct{ lo, hi rune }

// parseRanges reads a CSS unicode-range: "U+4E00-9FFF, U+3001".
func parseRanges(s string) ([]span, error) {
	var out []span
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		hex, ok := strings.CutPrefix(strings.ToUpper(part), "U+")
		if !ok {
			return nil, fmt.Errorf("not a range: %q", part)
		}
		lo, hi, isSpan := strings.Cut(hex, "-")
		if !isSpan {
			hi = lo
		}
		a, err := strconv.ParseUint(lo, 16, 32)
		if err != nil {
			return nil, fmt.Errorf("not a range: %q", part)
		}
		b, err := strconv.ParseUint(hi, 16, 32)
		if err != nil || b < a {
			return nil, fmt.Errorf("not a range: %q", part)
		}
		out = append(out, span{rune(a), rune(b)})
	}
	return out, nil
}

func mustRanges(s string) []span {
	r, err := parseRanges(s)
	if err != nil {
		panic(err)
	}
	return r
}

// intersect gives the code points in both, sorted and merged.
func intersect(a, b []span) []span {
	var out []span
	for _, x := range a {
		for _, y := range b {
			lo, hi := max(x.lo, y.lo), min(x.hi, y.hi)
			if lo <= hi {
				out = append(out, span{lo, hi})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].lo < out[j].lo })
	var merged []span
	for _, s := range out {
		if n := len(merged); n > 0 && s.lo <= merged[n-1].hi+1 {
			merged[n-1].hi = max(merged[n-1].hi, s.hi)
			continue
		}
		merged = append(merged, s)
	}
	return merged
}

func formatRanges(spans []span) string {
	parts := make([]string, len(spans))
	for i, s := range spans {
		if s.lo == s.hi {
			parts[i] = fmt.Sprintf("U+%X", s.lo)
		} else {
			parts[i] = fmt.Sprintf("U+%X-%X", s.lo, s.hi)
		}
	}
	return strings.Join(parts, ", ")
}
