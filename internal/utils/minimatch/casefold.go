package minimatch

import (
	"slices"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

// The NoCase option compares the way a JavaScript regexp with the `i` flag and
// without the `u` flag does, which is neither Go's case folding nor the one
// regexp2 reaches for when it is told to ignore case. JavaScript canonicalizes
// a character to its uppercase and holds two characters equal when they
// canonicalize alike, with the one rule that a character outside ASCII never
// canonicalizes into it. So U+03A3 Σ, U+03C3 σ and U+03C2 ς are all one
// character to it, while U+212A K and `k` are two.
//
// Rather than ask regexp2 to ignore case, a pattern is widened to name every
// character it should compare equal to and then matched exactly.

// canonicalize maps a character to the one JavaScript compares it as. Go's
// simple uppercase stands in for String.prototype.toUpperCase, which they only
// differ over where the uppercase spans more than one character — `ß` is the
// familiar one — and JavaScript keeps the original character there too.
func canonicalize(r rune) rune {
	upper := unicode.ToUpper(r)
	if r >= utf8.RuneSelf && upper < utf8.RuneSelf {
		return r
	}
	return upper
}

// caseTables groups the characters that canonicalize alike. A group of one has
// nothing to widen, so only the rest are kept: as a lookup by member, and as a
// list to walk when a character class has to be widened by range.
var caseTables = sync.OnceValues(func() (map[rune][]rune, [][]rune) {
	grouped := map[rune][]rune{}
	record := func(r rune) {
		canonical := canonicalize(r)
		if !slices.Contains(grouped[canonical], r) {
			grouped[canonical] = append(grouped[canonical], r)
		}
	}
	// Every character that canonicalizes onto another one has a case mapping,
	// so the case ranges name them all.
	for _, caseRange := range unicode.CaseRanges {
		for r := rune(caseRange.Lo); r <= rune(caseRange.Hi); r++ {
			record(r)
			record(canonicalize(r))
		}
	}

	byMember := map[rune][]rune{}
	groups := [][]rune{}
	for _, members := range grouped {
		if len(members) < 2 {
			continue
		}
		slices.Sort(members)
		groups = append(groups, members)
		for _, member := range members {
			byMember[member] = members
		}
	}
	return byMember, groups
})

// caseClass widens one literal character to the class of characters NoCase
// compares it equal to, reporting false when it stands alone.
func caseClass(r rune) (string, bool) {
	byMember, _ := caseTables()
	members, ok := byMember[r]
	if !ok {
		return "", false
	}
	var class strings.Builder
	class.WriteByte('[')
	for _, member := range members {
		class.WriteString(escapeClassRune(member))
	}
	class.WriteByte(']')
	return class.String(), true
}

// caseCloseClass widens a character class to cover every character NoCase
// compares equal to one it already covers, taking the class body this package
// wrote and returning the body to write instead.
//
// What the class already names is left alone and the rest is appended, so a
// negated class goes on negating whatever the widened class covers. That is
// how JavaScript reads one too: `[^a]` asks whether any member compares equal
// to the character at hand, and answers no to `A`.
func caseCloseClass(body string) string {
	singles, ranges := classMembers(body)
	covers := func(r rune) bool {
		return slices.Contains(singles, r) || slices.ContainsFunc(ranges, func(bounds [2]rune) bool {
			return r >= bounds[0] && r <= bounds[1]
		})
	}

	_, groups := caseTables()
	var extras strings.Builder
	for _, members := range groups {
		if !slices.ContainsFunc(members, covers) {
			continue
		}
		for _, member := range members {
			if !covers(member) {
				extras.WriteString(escapeClassRune(member))
			}
		}
	}
	if extras.Len() == 0 {
		return body
	}
	return escapeTrailingDash(body) + extras.String()
}

// escapeTrailingDash escapes the `-` that stands for itself at the end of a
// character class, so that appending to the class does not read it as opening
// a range instead.
func escapeTrailingDash(body string) string {
	if !strings.HasSuffix(body, "-") {
		return body
	}
	backslashes := 0
	for i := len(body) - 2; i >= 0 && body[i] == '\\'; i-- {
		backslashes++
	}
	if backslashes%2 == 1 {
		return body
	}
	return body[:len(body)-1] + `\-`
}

// classMembers reads back the characters a character class covers. The body is
// this package's own output, so it holds nothing but characters, backslash
// escapes, `-` ranges and the `^` that negates the whole class.
func classMembers(body string) ([]rune, [][2]rune) {
	body = strings.TrimPrefix(body, "^")

	type member struct {
		r    rune
		dash bool
	}
	members := []member{}
	for i := 0; i < len(body); {
		r, size := utf8.DecodeRuneInString(body[i:])
		if r == '\\' && i+size < len(body) {
			escaped, escapedSize := utf8.DecodeRuneInString(body[i+size:])
			members = append(members, member{r: escaped})
			i += size + escapedSize
			continue
		}
		members = append(members, member{r: r, dash: r == '-'})
		i += size
	}

	singles := []rune{}
	ranges := [][2]rune{}
	for i := 0; i < len(members); i++ {
		// A `-` that follows a range rather than separating one, as the second
		// one in `[a-c-e]` does, stands for itself.
		if i+2 < len(members) && members[i+1].dash {
			ranges = append(ranges, [2]rune{members[i].r, members[i+2].r})
			i += 2
			continue
		}
		singles = append(singles, members[i].r)
	}
	return singles, ranges
}

// escapeClassRune escapes a character that would carry a meaning of its own
// inside a character class.
func escapeClassRune(r rune) string {
	if strings.ContainsRune(`\]^-[`, r) {
		return `\` + string(r)
	}
	return string(r)
}
