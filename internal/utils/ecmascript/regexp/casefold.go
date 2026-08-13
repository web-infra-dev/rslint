package regexp

import (
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// A JavaScript regexp carrying the `i` flag compares two characters by
// ecmascript.Canonicalize, which is neither Go's case folding nor the Unicode
// folding regexp2 reaches for when told to ignore case. It is also not one
// reading but two: `u` and `v` select simple case folding, and their absence
// selects an uppercase mapping that never crosses into ASCII. U+212A KELVIN
// SIGN against `k` comes out differently under each, so every entry point
// here carries the flag through rather than picking one.
//
// So rather than ask regexp2 to ignore case, a pattern is widened: every
// literal and every character class is rewritten to name each character it
// should compare equal to, and the widened pattern is matched exactly.

// CaseClass widens one literal character to the class of characters a `/i`
// comparison accepts for it, reporting false when the character stands alone
// and needs no widening.
func CaseClass(r rune, unicode bool) (string, bool) {
	members := ecmascript.CaseEquivalents(r, unicode)
	if len(members) == 0 {
		return "", false
	}
	var class strings.Builder
	class.WriteByte('[')
	for _, member := range members {
		class.WriteString(EscapeClassRune(member))
	}
	class.WriteByte(']')
	return class.String(), true
}

// CaseCloseClass widens a character class to cover every character a `/i`
// comparison accepts for one it already covers. It takes the body of a class —
// what sits between the brackets, leading `^` included — and returns the body
// to write instead.
//
// What the class already names is left alone and the rest is appended, so a
// negated class goes on negating whatever the widened class covers. That is
// how JavaScript reads one too: `[^a]` asks whether any member compares equal
// to the character at hand, and answers no to `A`.
func CaseCloseClass(body string, unicode bool) string {
	singles, ranges := classMembers(body, unicode)
	covers := func(r rune) bool {
		return slices.Contains(singles, r) || slices.ContainsFunc(ranges, func(bounds [2]rune) bool {
			return r >= bounds[0] && r <= bounds[1]
		})
	}

	var extras strings.Builder
	for _, members := range ecmascript.CaseEquivalenceGroups(unicode) {
		if !slices.ContainsFunc(members, covers) {
			continue
		}
		for _, member := range members {
			if !covers(member) {
				extras.WriteString(EscapeClassRune(member))
			}
		}
	}
	if extras.Len() == 0 {
		return body
	}
	return escapeTrailingDash(body) + extras.String()
}

// EscapeClassRune escapes a character that would carry a meaning of its own
// inside a character class.
func EscapeClassRune(r rune) string {
	if strings.ContainsRune(`\]^-[`, r) {
		return `\` + string(r)
	}
	return string(r)
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

// classMembers reads back the characters a character class covers, so that
// widening it can tell what it already has. It understands the shapes a class
// body takes: a leading `^`, backslash escapes, `-` ranges, and plain
// characters.
//
// An escape is decoded to the character it names, so that `[\u0041-\u005A]`
// is read as the range `A` to `Z` it is. One naming a set rather than a
// character — `\d`, `\w`, `\s` — decodes to utf8.RuneError, which has no
// case-equivalent and so widens nothing.
func classMembers(body string, unicode bool) ([]rune, [][2]rune) {
	body = strings.TrimPrefix(body, "^")

	type member struct {
		r    rune
		dash bool
	}
	members := []member{}
	for i := 0; i < len(body); {
		r, size := utf8.DecodeRuneInString(body[i:])
		if r == '\\' && i+size < len(body) {
			escaped, escapedSize, ok := decodeEscape(body, i+size, unicode)
			if !ok {
				break
			}
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
