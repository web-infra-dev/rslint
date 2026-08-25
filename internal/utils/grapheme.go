package utils

import "github.com/rivo/uniseg"

// GraphemeCount returns the number of Unicode grapheme clusters (UAX #29) in s.
func GraphemeCount(s string) int {
	return uniseg.GraphemeClusterCount(s)
}
