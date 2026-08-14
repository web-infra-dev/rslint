package sort_keys

import "strconv"

// naturalCompare replicates the `natural-compare` npm package (v1.4.0:
// https://github.com/litejs/natural-compare-lite) that upstream's sort-keys
// rule uses for its `natural` option. It is not a conventional "digits sort
// numerically" natural order: non-alphanumeric ASCII characters are remapped
// through a bespoke weight table (see charWeight below), so e.g. '_' sorts
// before 'A', which sorts before 'a'. Kept local to this rule rather than
// reusing utils.NaturalCompare, which implements a different, more
// conventional algorithm already relied on by other rules.
//
// Operates on Unicode code points; upstream compares by UTF-16 code unit, so
// non-BMP (astral) characters may order differently in rare cases.
//
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
func naturalCompare(a, b string) int {
	if a == b {
		return 0
	}
	ra := []rune(a)
	rb := []rune(b)
	posA, posB := 0, 0

	for {
		codeA0 := charWeight(ra, posA)
		codeB0 := charWeight(rb, posB)
		posA++
		posB++

		cmp := 0
		if isDigitWeightStart(codeA0) && isDigitWeightStart(codeB0) {
			iA := posA - 1
			for isDigitWeightCont(charWeight(ra, iA)) {
				iA++
			}
			iB := posB - 1
			for isDigitWeightCont(charWeight(rb, iB)) {
				iB++
			}
			numA := parseDigitRun(ra[posA-1 : iA])
			numB := parseDigitRun(rb[posB-1 : iB])
			posA, posB = iA, iB
			switch {
			case numA < numB:
				cmp = -1
			case numA > numB:
				cmp = 1
			}
		} else {
			switch {
			case codeA0 < codeB0:
				cmp = -1
			case codeA0 > codeB0:
				cmp = 1
			}
		}

		if cmp != 0 {
			return cmp
		}
		if codeB0 == 0 {
			return 0
		}
	}
}

// isDigitWeightStart reports whether a character weight can begin a
// multi-digit numeric run: the weights for '1'-'9' are 67-75. '0' (weight
// 66) is deliberately excluded, matching upstream's own quirk — a run can
// only start on a nonzero digit.
func isDigitWeightStart(weight int) bool {
	return weight > 66 && weight < 76
}

// isDigitWeightCont reports whether a character weight can continue an
// already-started numeric run: the weights for '0'-'9' are 66-75.
func isDigitWeightCont(weight int) bool {
	return weight > 65 && weight < 76
}

func parseDigitRun(digits []rune) float64 {
	n, _ := strconv.ParseFloat(string(digits), 64)
	return n
}

// charWeight remaps a single character's code point into upstream's sort
// weight, matching natural-compare's `getCode` helper. pos out of range
// yields weight 0, signaling end-of-string.
func charWeight(runes []rune, pos int) int {
	code := 0
	if pos >= 0 && pos < len(runes) {
		code = int(runes[pos])
	}
	switch {
	case code < 45 || code > 127:
		return code
	case code < 46: // '-'
		return 65
	case code < 48: // '.', '/'
		return code - 1
	case code < 58: // '0'-'9'
		return code + 18
	case code < 65: // ':'..'@'
		return code - 11
	case code < 91: // 'A'-'Z'
		return code + 11
	case code < 97: // '['..'`'
		return code - 37
	case code < 123: // 'a'-'z'
		return code + 5
	default: // '{'..DEL
		return code - 63
	}
}
