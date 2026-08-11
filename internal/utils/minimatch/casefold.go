package minimatch

import (
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// The NoCase option compares the way a JavaScript regexp with the `i` flag and
// without the `u` flag does, which ecmascript.Canonicalize spells out. Rather than ask
// regexp2 to ignore case — its own rules are Unicode's, and disagree in both
// directions — a pattern is widened to name every character it should compare
// equal to, and then matched exactly.

// caseClass widens one literal character to the class of characters NoCase
// compares it equal to, reporting false when it stands alone.
func caseClass(r rune) (string, bool) {
	members := ecmascript.CaseEquivalents(r)
	if len(members) == 0 {
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

	var extras strings.Builder
	for _, members := range ecmascript.CaseEquivalenceGroups() {
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
