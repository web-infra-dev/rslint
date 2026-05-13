// cspell:ignore Damerau damerau

package aria_props

import (
	"sort"
	"strings"
)

// getSuggestion mirrors `eslint-plugin-jsx-a11y/src/util/getSuggestion.js`.
//
// Upstream computes a Damerau-Levenshtein (OSA — Optimal String Alignment)
// distance between the UPPER-CASED forms of `word` and every entry in
// `dictionary`, retains entries whose distance ≤ THRESHOLD (2), sorts by
// ascending distance (stable — `Array.prototype.sort` is stable in modern
// engines and so is `sort.SliceStable`), and slices to `limit` results.
//
// `dictionaryUpper` is the upper-cased form of `dictionary`, indexed 1:1.
// Passing both as arguments keeps the upper-casing out of the hot path —
// caller pre-computes it once at init time. The two slices must be the
// same length and aligned by index.
//
// Returns a freshly-allocated slice — never `nil` — so callers can safely
// `len()` and range without nil-guards.
func getSuggestion(word string, dictionary, dictionaryUpper []string, limit int) []string {
	type candidate struct {
		word     string
		distance int
	}
	wordUpper := strings.ToUpper(word)
	candidates := make([]candidate, 0, len(dictionary))
	for i, dUpper := range dictionaryUpper {
		dist := osaDistance(wordUpper, dUpper)
		if dist <= suggestionDistanceThreshold {
			candidates = append(candidates, candidate{word: dictionary[i], distance: dist})
		}
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].distance < candidates[j].distance
	})
	if limit > 0 && len(candidates) > limit {
		candidates = candidates[:limit]
	}
	out := make([]string, len(candidates))
	for i, c := range candidates {
		out[i] = c.word
	}
	return out
}

// osaDistance computes the Optimal String Alignment distance — the variant
// of Damerau-Levenshtein implemented by the npm `damerau-levenshtein`
// package upstream uses. OSA differs from "true" Damerau-Levenshtein in
// that no substring may be edited more than once: it allows a single
// transposition of adjacent characters but never compounds a transposition
// with a later edit on the same characters. For the ARIA-suggestion
// use-case (single-char typos in short attribute names) the two algorithms
// agree.
//
// The algorithm operates over Unicode code points (rune slices), not bytes,
// so it stays correct on non-ASCII input even though upstream's dictionary
// is pure ASCII.
//
// The DP matrix is laid out in a single contiguous `[]int` and addressed
// row-by-row via per-row slices — same access pattern as a `[][]int`, but
// two heap allocations per call instead of `la+2`.
func osaDistance(a, b string) int {
	ra := []rune(a)
	rb := []rune(b)
	la := len(ra)
	lb := len(rb)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	// d[i][j] = edit distance between ra[:i] and rb[:j], backed by a
	// single flat slice.
	backing := make([]int, (la+1)*(lb+1))
	d := make([][]int, la+1)
	for i := range d {
		d[i] = backing[i*(lb+1) : (i+1)*(lb+1)]
		d[i][0] = i
	}
	for j := 0; j <= lb; j++ {
		d[0][j] = j
	}
	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			d[i][j] = min(
				d[i-1][j]+1,      // deletion
				d[i][j-1]+1,      // insertion
				d[i-1][j-1]+cost, // substitution
			)
			// Transposition arm. npm `damerau-levenshtein` uses
			// `d[i-2][j-2] + cost`, not `+ 1`. For a real adjacent
			// swap of two distinct characters, `cost` is always 1, so
			// `+ cost` and `+ 1` agree. They diverge ONLY when all
			// four involved characters are equal, in which case
			// substitution already drives the cell to 0 and the
			// transposition arm is never selected. Mirror npm
			// verbatim regardless, to keep the algorithm byte-for-byte
			// identical.
			if i > 1 && j > 1 && ra[i-1] == rb[j-2] && ra[i-2] == rb[j-1] {
				if v := d[i-2][j-2] + cost; v < d[i][j] {
					d[i][j] = v
				}
			}
		}
	}
	return d[la][lb]
}
