package utils

import (
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/scanner"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

// RegexCapturingGroup is one capturing group found by RegexCapturingGroups:
// either a plain `(...)` group or a named `(?<name>...)` group. Start is the
// byte offset of the group's opening `(` within the pattern; End is one past
// the matching `)`.
type RegexCapturingGroup struct {
	Start int
	End   int
	Name  string // "" for an unnamed group
}

// RegexCapturingGroups walks an ECMAScript regex pattern and returns every
// capturing group it contains, in source order. It is not a full regex
// parser — it recognizes just enough syntax (groups, character classes,
// escapes, quantifiers, ES2025 modifier groups) to correctly skip past
// constructs it doesn't need while finding every `(` that opens a capturing
// group. ok is false when the pattern fails to parse, mirroring how
// regexpp/the host engine would reject it; callers should treat that the same
// as "no capturing groups" (skip the pattern entirely).
//
// A handful of checks that need tables or a second model of the pattern are
// left out, so those patterns come back ok even though the host engine rejects
// them: unknown `\p{...}` property names, a group name repeated within one
// alternative, and v-mode's rule that a class `-` outside a range be escaped.
func RegexCapturingGroups(pattern string, flags RegexFlags) (groups []RegexCapturingGroup, ok bool) {
	if flags.Unicode && flags.UnicodeSets {
		return nil, false
	}
	if flags.UV() && !regexPropertyEscapesValid(pattern, flags) {
		return nil, false
	}
	if flags.UnicodeSets && hasIncompleteVClassOperator(pattern) {
		return nil, false
	}
	classesValid := true
	if !IterateRegexCharacterClasses(pattern, flags, func(start, end int) {
		elements, _, parsed := ParseRegexCharacterClassWithEnd(pattern, start, end, flags)
		if !parsed {
			// The shared parser intentionally leaves some Annex B legacy forms
			// unsupported (for example `[\\c]` without u/v). In strict Unicode
			// modes, though, a failure denotes a malformed class and must reject
			// the entire pattern.
			if flags.UV() {
				classesValid = false
			}
			return
		}
		for _, element := range elements {
			if element.Kind == RegexCharRange && element.Value > element.Max {
				classesValid = false
				return
			}
		}
		if flags.UV() && classHasInvalidUnicodeEscape(pattern, start, end, flags.UnicodeSets) {
			classesValid = false
		}
	}) || !classesValid {
		return nil, false
	}

	pos := 0
	if !scanRegexAlternatives(pattern, flags, &pos, &groups, false) {
		return nil, false
	}
	if pos != len(pattern) {
		return nil, false
	}
	if !regexBackrefsResolve(pattern, flags, groups) {
		return nil, false
	}
	if !regexDuplicateNamedGroupsValid(pattern, flags, groups) {
		return nil, false
	}
	return groups, true
}

// hasIncompleteVClassOperator catches the v-only set operators which the
// capture scanner otherwise skips as ordinary class text. The shared full v
// validator is intentionally stricter than this scanner for nested sets, so
// keep this check to the unambiguous dangling-operator forms.
func hasIncompleteVClassOperator(pattern string) bool {
	invalid := false
	IterateRegexCharacterClasses(pattern, RegexFlags{UnicodeSets: true}, func(start, end int) {
		if invalid {
			return
		}
		body := pattern[start+1 : end-1]
		sawIntersection := false
		sawDifference := false
		for i := 0; i < len(body); i++ {
			if body[i] == '\\' {
				i++
				continue
			}
			if body[i] == '&' && i+1 < len(body) && body[i+1] == '&' {
				if i == 0 || i+2 >= len(body) || body[i-1] == '&' || body[i+2] == '&' {
					invalid = true
					return
				}
				i++
				sawIntersection = true
				continue
			}
			if body[i] == '-' {
				if i+1 < len(body) && body[i+1] == '-' {
					if i == 0 || i+2 >= len(body) || body[i-1] == '-' || body[i+2] == '-' {
						invalid = true
						return
					}
					i++
					sawDifference = true
					continue
				}
				if i == 0 || i+1 >= len(body) {
					invalid = true
					return
				}
			}
		}
		if sawIntersection && sawDifference {
			invalid = true
		}
	})
	return invalid
}

type regexAlternative struct {
	group int
	index int
}

// regexDuplicateNamedGroupsValid permits a duplicate name only when every
// possible match chooses at most one of the declarations. The ECMAScript
// grammar permits the two sides of a disjunction to reuse a name, but rejects
// sequential or nested duplicates that can both participate in one match.
func regexDuplicateNamedGroupsValid(pattern string, flags RegexFlags, groups []RegexCapturingGroup) bool {
	paths := make(map[int][]regexAlternative)
	stack := []regexAlternative{{group: 0}}
	nextGroup := 1
	for i := 0; i < len(pattern); {
		switch pattern[i] {
		case '\\':
			step, ok := skipPatternLevelEscape(pattern, i, flags)
			if !ok {
				return false
			}
			i += step
		case '[':
			end, ok := ClassEnd(pattern, i, flags)
			if !ok {
				return false
			}
			i = end
		case '(':
			paths[i] = append([]regexAlternative(nil), stack...)
			stack = append(stack, regexAlternative{group: nextGroup})
			nextGroup++
			i++
		case '|':
			stack[len(stack)-1].index++
			i++
		case ')':
			if len(stack) == 1 {
				return false
			}
			stack = stack[:len(stack)-1]
			i++
		default:
			_, width := utf8.DecodeRuneInString(pattern[i:])
			i += width
		}
	}
	seen := make(map[string][][]regexAlternative)
	for _, group := range groups {
		if group.Name == "" {
			continue
		}
		name, ok := normalizeRegexGroupName(group.Name)
		if !ok {
			return false
		}
		path := paths[group.Start]
		for _, earlier := range seen[name] {
			if !regexPathsExclusive(earlier, path) {
				return false
			}
		}
		seen[name] = append(seen[name], path)
	}
	return true
}

func regexPathsExclusive(left, right []regexAlternative) bool {
	for i := 0; i < len(left) && i < len(right); i++ {
		if left[i].group != right[i].group {
			return false
		}
		if left[i].index != right[i].index {
			return true
		}
	}
	return false
}

// regexPropertyEscapesValid delegates each complete Unicode property escape to
// the shared ECMAScript regex compiler. The capturing-group scanner otherwise
// only needs to skip over property escapes, but an unknown property name makes
// the entire constructor pattern invalid and must suppress diagnostics.
func regexPropertyEscapesValid(pattern string, flags RegexFlags) bool {
	for i := 0; i < len(pattern); {
		if pattern[i] != '\\' {
			_, width := utf8.DecodeRuneInString(pattern[i:])
			i += width
			continue
		}
		step, ok := SkipPatternEscape(pattern, i, flags)
		if !ok {
			return false
		}
		if i+1 < len(pattern) && (pattern[i+1] == 'p' || pattern[i+1] == 'P') {
			if flags.UnicodeSets && isUnicodeSetsStringProperty(pattern[i:i+step]) {
				i += step
				continue
			}
			if _, err := esregexp.Compile(pattern[i:i+step], "u"); err != nil {
				return false
			}
		}
		i += step
	}
	return true
}

// isUnicodeSetsStringProperty recognizes the properties of strings that v-mode
// adds beyond u-mode's code-point properties. The shared compiler validates
// under u, so it cannot accept these otherwise-valid v-mode escapes.
func isUnicodeSetsStringProperty(escape string) bool {
	if len(escape) < 5 || escape[0] != '\\' || escape[1] != 'p' || escape[2] != '{' || escape[len(escape)-1] != '}' {
		return false
	}
	switch escape[3 : len(escape)-1] {
	case "Basic_Emoji", "Emoji_Keycap_Sequence", "RGI_Emoji", "RGI_Emoji_Flag_Sequence",
		"RGI_Emoji_Modifier_Sequence", "RGI_Emoji_Tag_Sequence", "RGI_Emoji_ZWJ_Sequence":
		return true
	default:
		return false
	}
}

// reservedClassSetPunctuators are the characters ECMAScript reserves inside a
// v-flag class when they appear doubled (`!!`, `##`, `..`, …). `&` is in the
// set because `&&` is only legal as the intersection operator.
const reservedClassSetPunctuators = "&!#$%*+,.:;<=>?@^`~"

// classHasInvalidUnicodeEscape validates the strict escape forms ClassEnd
// deliberately treats permissively to find a closing bracket. In u/v classes,
// an identity escape is not a fallback: incomplete hex escapes, legacy octal
// escapes, malformed control escapes, and Unicode code points above U+10FFFF
// reject the whole pattern.
func classHasInvalidUnicodeEscape(pattern string, start, end int, unicodeSets bool) bool {
	for i := start + 1; i < end-1; {
		if pattern[i] != '\\' {
			_, w := utf8.DecodeRuneInString(pattern[i:])
			i += w
			continue
		}
		if i+1 >= end-1 {
			return true
		}
		// `\B` and non-zero decimal escapes are valid AtomEscapes but not
		// CharacterClassEscape forms. They are SyntaxErrors inside a u/v class.
		if pattern[i+1] == 'B' || (pattern[i+1] >= '1' && pattern[i+1] <= '9') {
			return true
		}
		// A hyphen may be escaped in a u-mode class. In v mode it is instead a
		// class-set reserved punctuator and is handled with the other escapes.
		if !unicodeSets && pattern[i+1] == '-' {
			i += 2
			continue
		}
		// v-mode permits escaping its class-set reserved punctuators, which
		// aren't AtomEscape syntax characters outside a class.
		if unicodeSets && strings.ContainsRune(reservedClassSetPunctuators+"()[]{}\\/|-", rune(pattern[i+1])) {
			i += 2
			continue
		}
		if unicodeSets && pattern[i+1] == 'q' && i+2 < end-1 && pattern[i+2] == '{' {
			step, ok := SkipPatternEscape(pattern, i, RegexFlags{UnicodeSets: true})
			if !ok {
				return true
			}
			i += step
			continue
		}
		// Named backreferences are AtomEscapes and never class elements.
		if pattern[i+1] == 'k' {
			return true
		}
		step, ok := skipUnicodePatternEscape(pattern, i)
		if !ok {
			return true
		}
		i += step
	}
	return false
}

// regexBackrefsResolve reports whether every backreference in an already
// well-formed pattern points at a group that exists. Numeric backreferences
// require u/v mode. In legacy mode, named references are required only when a
// named capture exists; otherwise `\k` is an identity escape.
func regexBackrefsResolve(pattern string, flags RegexFlags, groups []RegexCapturingGroup) bool {
	hasNamedGroup := false
	for _, group := range groups {
		if group.Name != "" {
			hasNamedGroup = true
			break
		}
	}
	for i := 0; i < len(pattern); {
		switch pattern[i] {
		case '[':
			end, ok := ClassEnd(pattern, i, flags)
			if !ok {
				return false
			}
			i = end
		case '\\':
			step, ok := skipPatternLevelEscape(pattern, i, flags)
			if !ok {
				return false
			}
			if !flags.UV() && hasNamedGroup && i+3 < len(pattern) && pattern[i+1] == 'k' && pattern[i+2] == '<' {
				if closeRel := strings.IndexByte(pattern[i+3:], '>'); closeRel > 0 {
					if !hasGroupNamed(groups, pattern[i+3:i+3+closeRel]) {
						return false
					}
					i += closeRel + 4
					continue
				}
			}
			switch next := pattern[i+1]; {
			case flags.UV() && next >= '1' && next <= '9':
				n := 0
				for _, digit := range []byte(pattern[i+1 : i+step]) {
					n = n*10 + int(digit-'0')
					if n > len(groups) {
						return false
					}
				}
			case flags.UV() && next == 'k':
				// skipPatternLevelEscape only accepts `\k` as `\k<name>`.
				if !hasGroupNamed(groups, pattern[i+3:i+step-1]) {
					return false
				}
			}
			i += step
		default:
			_, w := utf8.DecodeRuneInString(pattern[i:])
			if w == 0 {
				return false
			}
			i += w
		}
	}
	return true
}

func hasGroupNamed(groups []RegexCapturingGroup, name string) bool {
	want, ok := normalizeRegexGroupName(name)
	if !ok {
		return false
	}
	for _, group := range groups {
		got, ok := normalizeRegexGroupName(group.Name)
		if ok && got == want {
			return true
		}
	}
	return false
}

// normalizeRegexGroupName decodes the Unicode escapes allowed in a capture
// name so a backreference can be compared by IdentifierName value rather than
// by its authored spelling.
func normalizeRegexGroupName(name string) (string, bool) {
	var result strings.Builder
	for i := 0; i < len(name); {
		if name[i] != '\\' {
			r, width := utf8.DecodeRuneInString(name[i:])
			if r == utf8.RuneError && width == 1 {
				return "", false
			}
			result.WriteRune(r)
			i += width
			continue
		}

		value, width, ok := decodeRegexNameEscape(name, i)
		if !ok {
			return "", false
		}
		if value >= 0xD800 && value <= 0xDBFF {
			next, nextWidth, ok := decodeRegexNameEscape(name, i+width)
			if !ok || next < 0xDC00 || next > 0xDFFF {
				return "", false
			}
			result.WriteRune(utf16.DecodeRune(rune(value), rune(next)))
			width += nextWidth
		} else {
			if value >= 0xDC00 && value <= 0xDFFF {
				return "", false
			}
			result.WriteRune(rune(value))
		}
		i += width
	}
	return result.String(), true
}

func decodeRegexNameEscape(name string, start int) (uint32, int, bool) {
	if start+2 >= len(name) || name[start] != '\\' || name[start+1] != 'u' {
		return 0, 0, false
	}
	if name[start+2] == '{' {
		closeRel := strings.IndexByte(name[start+3:], '}')
		if closeRel < 1 {
			return 0, 0, false
		}
		digits := name[start+3 : start+3+closeRel]
		if len(digits) > 6 || !allHexStr(digits) {
			return 0, 0, false
		}
		value, err := strconv.ParseUint(digits, 16, 32)
		if err != nil || value > utf8.MaxRune {
			return 0, 0, false
		}
		return uint32(value), closeRel + 4, true
	}
	if start+6 > len(name) || !allHexStr(name[start+2:start+6]) {
		return 0, 0, false
	}
	value, err := strconv.ParseUint(name[start+2:start+6], 16, 16)
	if err != nil {
		return 0, 0, false
	}
	return uint32(value), 6, true
}

func scanRegexAlternatives(pattern string, flags RegexFlags, pos *int, groups *[]RegexCapturingGroup, expectClose bool) bool {
	for *pos < len(pattern) {
		c := pattern[*pos]
		if c == ')' {
			return expectClose
		}
		if c == '|' {
			*pos++
			continue
		}
		if !scanRegexTerm(pattern, flags, pos, groups) {
			return false
		}
	}
	return !expectClose
}

func scanRegexTerm(pattern string, flags RegexFlags, pos *int, groups *[]RegexCapturingGroup) bool {
	// Assertions can't carry a quantifier. Annex B makes an exception for
	// lookahead, which u/v mode takes back.
	quantifiable := true
	c := pattern[*pos]
	switch c {
	case '(':
		groupQuantifiable, ok := scanRegexGroup(pattern, flags, pos, groups)
		if !ok {
			return false
		}
		quantifiable = groupQuantifiable
	case '^', '$':
		quantifiable = false
		*pos++
	case '[':
		end, ok := ClassEnd(pattern, *pos, flags)
		if !ok {
			return false
		}
		// `\q` inside a class is a SyntaxError, not an identity escape: under v
		// unless it opens a `\q{...}` string literal, and under u always, since
		// string literals are v-only. ClassEnd's own escape-skipping doesn't
		// distinguish this — check it separately so a "how do you spell an
		// invalid class" pattern isn't silently accepted as valid.
		if flags.UV() && classHasInvalidQEscape(pattern, *pos, end, flags.UnicodeSets) {
			return false
		}
		*pos = end
	case '\\':
		step, ok := skipPatternLevelEscape(pattern, *pos, flags)
		if !ok {
			return false
		}
		if next := pattern[*pos+1]; step > 1 && (next == 'b' || next == 'B') {
			quantifiable = false
		}
		*pos += step
	case '*', '+', '?':
		// Standalone quantifier without an operand. Always a syntax error
		// here (Annex B's lenient legacy-quantifier fallback is not
		// modeled); the caller bails out on the whole pattern.
		return false
	case '{':
		// A `{n}`/`{n,m}` reached at term position has no operand to quantify,
		// which is a syntax error in every mode. A `{` that doesn't spell a
		// quantifier is a syntax error under u/v and a literal outside it.
		probe := *pos
		if !skipRegexQuantifier(pattern, &probe) {
			return false
		}
		if probe != *pos || flags.UV() {
			return false
		}
		*pos++
	case '}', ']':
		// A closer with no opener: a literal outside u/v, rejected under it.
		if flags.UV() {
			return false
		}
		*pos++
	default:
		_, w := utf8.DecodeRuneInString(pattern[*pos:])
		if w == 0 {
			return false
		}
		*pos += w
	}
	if quantifiable {
		return skipRegexQuantifier(pattern, pos)
	}
	probe := *pos
	if !skipRegexQuantifier(pattern, &probe) {
		return false
	}
	return probe == *pos
}

// skipPatternLevelEscape returns how many bytes the `\`-prefixed escape at
// pattern[i] consumes when it appears outside a character class.
//
// Under u/v the escape grammar is closed — every form is spelled out and
// anything else is a SyntaxError — so those patterns get their own strict
// reading. Outside u/v the escapes are the permissive Annex B set that
// SkipPatternEscape models, with one exception: `\c` opens a control escape
// only when a ControlLetter follows, and otherwise the backslash stands alone
// as a literal `\`, leaving the `c` to be scanned as an ordinary character.
func skipPatternLevelEscape(pattern string, i int, flags RegexFlags) (int, bool) {
	if i+1 >= len(pattern) {
		return 0, false
	}
	if flags.UV() {
		return skipUnicodePatternEscape(pattern, i)
	}
	if pattern[i+1] == 'c' && (i+2 >= len(pattern) || !isASCIILetter(pattern[i+2])) {
		return 1, true
	}
	return SkipPatternEscape(pattern, i, flags)
}

// skipUnicodePatternEscape is skipPatternLevelEscape's u/v-mode reading: it
// accepts only the escapes the ES grammar lists for an AtomEscape and rejects
// everything else, where Annex B would have fallen back to an identity escape.
// `\q{...}` lands in that rejected set — it's a v-mode string disjunction, and
// those exist only inside a character class.
//
// Property names inside `\p{...}` aren't validated, so an unknown one is still
// accepted here.
func skipUnicodePatternEscape(pattern string, i int) (int, bool) {
	rest := pattern[i+2:]
	switch c := pattern[i+1]; c {
	case 'b', 'B', 'd', 'D', 's', 'S', 'w', 'W', 'f', 'n', 'r', 't', 'v',
		'^', '$', '\\', '.', '*', '+', '?', '(', ')', '[', ']', '{', '}', '|', '/':
		return 2, true
	case 'c':
		if len(rest) > 0 && isASCIILetter(rest[0]) {
			return 3, true
		}
	case '0':
		// `\0` is the NUL escape only when no digit follows; `\01` is a legacy
		// octal escape, which u/v mode doesn't have.
		if len(rest) == 0 || !isRegexDigit(rest[0]) {
			return 2, true
		}
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		n := 0
		for n < len(rest) && isRegexDigit(rest[n]) {
			n++
		}
		return 2 + n, true
	case 'x':
		if len(rest) >= 2 && isHex(rest[0]) && isHex(rest[1]) {
			return 4, true
		}
	case 'u':
		if len(rest) > 0 && rest[0] == '{' {
			n := 1
			for n < len(rest) && isHex(rest[n]) {
				n++
			}
			if n > 1 && n < len(rest) && rest[n] == '}' {
				value, err := strconv.ParseUint(rest[1:n], 16, 32)
				if err == nil && value <= utf8.MaxRune {
					return 3 + n, true
				}
			}
			break
		}
		if len(rest) >= 4 && allHexStr(rest[:4]) {
			return 6, true
		}
	case 'p', 'P':
		if len(rest) > 0 && rest[0] == '{' {
			if closeRel := strings.IndexByte(rest[1:], '}'); closeRel >= 0 {
				return 4 + closeRel, true
			}
		}
	case 'k':
		if len(rest) > 0 && rest[0] == '<' {
			if closeRel := strings.IndexByte(rest[1:], '>'); closeRel > 0 {
				return 4 + closeRel, true
			}
		}
	}
	return 0, false
}

func skipRegexQuantifier(pattern string, pos *int) bool {
	if *pos >= len(pattern) {
		return true
	}
	switch pattern[*pos] {
	case '?', '*', '+':
		*pos++
	case '{':
		save := *pos
		i := *pos + 1
		nStart := i
		for i < len(pattern) && isRegexDigit(pattern[i]) {
			i++
		}
		if i == nStart {
			return true
		}
		minimum := pattern[nStart:i]
		maximum := ""
		if i < len(pattern) && pattern[i] == ',' {
			i++
			maximumStart := i
			for i < len(pattern) && isRegexDigit(pattern[i]) {
				i++
			}
			if maximumStart != i {
				maximum = pattern[maximumStart:i]
			}
		}
		if i < len(pattern) && pattern[i] == '}' {
			if maximum != "" && decimalLess(maximum, minimum) {
				return false
			}
			*pos = i + 1
		} else {
			*pos = save
			return true
		}
	default:
		return true
	}
	if *pos < len(pattern) && pattern[*pos] == '?' {
		*pos++
	}
	return true
}

func decimalLess(left, right string) bool {
	left = strings.TrimLeft(left, "0")
	right = strings.TrimLeft(right, "0")
	if left == "" {
		left = "0"
	}
	if right == "" {
		right = "0"
	}
	if len(left) != len(right) {
		return len(left) < len(right)
	}
	return left < right
}

// scanRegexGroup consumes the `(...)` starting at pattern[*pos], recording it
// when it captures. quantifiable is false for the assertion forms a quantifier
// may not follow.
func scanRegexGroup(pattern string, flags RegexFlags, pos *int, groups *[]RegexCapturingGroup) (quantifiable, ok bool) {
	start := *pos
	*pos++ // consume '('

	isCapturing := false
	quantifiable = true
	name := ""

	if *pos < len(pattern) && pattern[*pos] == '?' {
		if *pos+1 >= len(pattern) {
			return false, false
		}
		switch pattern[*pos+1] {
		case ':':
			*pos += 2
		case '=', '!':
			// Annex B allows a quantifier on lookahead; u/v mode doesn't.
			quantifiable = !flags.UV()
			*pos += 2
		case '<':
			if *pos+2 >= len(pattern) {
				return false, false
			}
			switch pattern[*pos+2] {
			case '=', '!':
				quantifiable = false
				*pos += 3
			default:
				// (?<name>...)
				*pos += 2 // consume "?<"
				nameEnd := *pos
				for nameEnd < len(pattern) && pattern[nameEnd] != '>' {
					nameEnd++
				}
				if nameEnd >= len(pattern) {
					return false, false
				}
				name = pattern[*pos:nameEnd]
				if !isRegexGroupName(name) {
					return false, false
				}
				*pos = nameEnd + 1
				isCapturing = true
			}
		default:
			// ES2025 modifier group: (?ims-ims:...) / (?i:...) / (?-i:...)
			if !skipRegexModifierGroupHeader(pattern, pos) {
				return false, false
			}
		}
	} else {
		isCapturing = true
	}

	groupIdx := -1
	if isCapturing {
		*groups = append(*groups, RegexCapturingGroup{Start: start, Name: name})
		groupIdx = len(*groups) - 1
	}

	if !scanRegexAlternatives(pattern, flags, pos, groups, true) {
		return false, false
	}
	if *pos >= len(pattern) || pattern[*pos] != ')' {
		return false, false
	}
	*pos++ // consume ')'

	if groupIdx >= 0 {
		(*groups)[groupIdx].End = *pos
	}
	return quantifiable, true
}

// isRegexGroupName reports whether name could be an IdentifierName, the grammar
// a `(?<name>...)` group name has to follow.
func isRegexGroupName(name string) bool {
	if name == "" {
		return false
	}
	normalized, ok := normalizeRegexGroupName(name)
	if !ok || normalized == "" {
		return false
	}
	for i, r := range normalized {
		if (i == 0 && !scanner.IsIdentifierStart(r)) || (i != 0 && !scanner.IsIdentifierPart(r)) {
			return false
		}
	}
	return true
}

// skipRegexModifierGroupHeader consumes an ES2025 regex-modifier group header
// `?ims-ims:` / `?ims:` / `?-ims:` starting at pattern[*pos]=='?', leaving
// *pos positioned right after the ':'. Only `i`, `m`, and `s` are valid
// modifier flags, no flag may repeat within a set, the added and removed sets
// must be disjoint, and at least one of them must be non-empty.
func skipRegexModifierGroupHeader(pattern string, pos *int) bool {
	i := *pos + 1 // skip '?'
	added, ok := scanRegexModifierFlags(pattern, &i)
	if !ok {
		return false
	}
	removed := uint(0)
	if i < len(pattern) && pattern[i] == '-' {
		i++
		if removed, ok = scanRegexModifierFlags(pattern, &i); !ok {
			return false
		}
		if added|removed == 0 || added&removed != 0 {
			return false
		}
	}
	if i >= len(pattern) || pattern[i] != ':' {
		return false
	}
	*pos = i + 1
	return true
}

// scanRegexModifierFlags consumes the run of modifier flags at pattern[*pos]
// and returns them as a bit set, rejecting a flag that appears twice.
func scanRegexModifierFlags(pattern string, pos *int) (uint, bool) {
	seen := uint(0)
	for *pos < len(pattern) && isRegexModifierFlag(pattern[*pos]) {
		bit := uint(1) << (pattern[*pos] - 'a')
		if seen&bit != 0 {
			return 0, false
		}
		seen |= bit
		*pos++
	}
	return seen, true
}

// classHasInvalidQEscape reports whether pattern[start:end] (a `[...]` class,
// as bounded by ClassEnd) contains a `\q` that isn't a valid escape there:
// under v one that doesn't open a `\q{...}` string literal, and under u any at
// all. `\\q` (an escaped backslash followed by a literal `q`) doesn't count —
// this walks escapes the same way ClassEnd/SkipPatternEscape do so it isn't
// misled by one.
func classHasInvalidQEscape(pattern string, start, end int, unicodeSets bool) bool {
	for i := start; i < end; {
		if pattern[i] != '\\' {
			_, w := utf8.DecodeRuneInString(pattern[i:])
			if w == 0 {
				i++
			} else {
				i += w
			}
			continue
		}
		if i+1 < end && pattern[i+1] == 'q' && (!unicodeSets || i+2 >= end || pattern[i+2] != '{') {
			return true
		}
		step, ok := SkipPatternEscape(pattern, i, RegexFlags{UnicodeSets: true})
		if !ok {
			return false
		}
		i += step
	}
	return false
}

func isRegexModifierFlag(c byte) bool { return c == 'i' || c == 'm' || c == 's' }

// isASCIILetter matches a regex ControlLetter, and the ASCII half of an
// IdentifierName's character set.
func isASCIILetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func isRegexDigit(c byte) bool { return c >= '0' && c <= '9' }
