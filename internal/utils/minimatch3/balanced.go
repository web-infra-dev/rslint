package minimatch3

import "strings"

// balancedMatch is a port of the npm package balanced-match. It finds the
// outermost balanced pair of open/close markers in str.
type balancedMatch struct {
	pre  string
	body string
	post string
}

// findBalanced reports the outermost balanced opener/closer pair in str,
// matching nested pairs the way brace expansion needs.
func findBalanced(opener string, closer string, str string) (balancedMatch, bool) {
	start, end, ok := balancedRange(opener, closer, str)
	if !ok {
		return balancedMatch{}, false
	}

	match := balancedMatch{pre: str[:start]}
	if start+len(opener) <= end {
		match.body = str[start+len(opener) : end]
	}
	if end+len(closer) <= len(str) {
		match.post = str[end+len(closer):]
	}
	return match, true
}

func indexFrom(str string, substr string, offset int) int {
	if offset < 0 {
		offset = 0
	}
	if offset > len(str) {
		return -1
	}
	index := strings.Index(str[offset:], substr)
	if index < 0 {
		return -1
	}
	return index + offset
}

func balancedRange(opener string, closer string, str string) (int, int, bool) {
	ai := indexFrom(str, opener, 0)
	bi := indexFrom(str, closer, ai+1)
	if ai < 0 || bi <= 0 {
		return 0, 0, false
	}

	begs := []int{}
	left := len(str)
	right := 0
	found := false

	for i := ai; i >= 0 && !found; {
		switch {
		case i == ai:
			begs = append(begs, i)
			ai = indexFrom(str, opener, i+1)
		case len(begs) == 1:
			left, right = begs[0], bi
			begs = begs[:0]
			found = true
		default:
			beg := begs[len(begs)-1]
			begs = begs[:len(begs)-1]
			if beg < left {
				left = beg
				right = bi
			}
			bi = indexFrom(str, closer, i+1)
		}

		if ai < bi && ai >= 0 {
			i = ai
		} else {
			i = bi
		}
	}

	if found || len(begs) != 0 {
		return left, right, true
	}
	return 0, 0, false
}
