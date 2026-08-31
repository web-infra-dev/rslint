package utils

import (
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/scanner"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

// IsValidRegexPattern reports whether pattern parses cleanly under flags, the
// way JavaScript's RegExp constructor would (a try/catch around parsing, not
// a match attempt). Under the u/v flags this combines three checks:
//
//   - IterateRegexCharacterClasses catches unterminated classes, including
//     v-flag nested classes such as `[[a]` (valid `u` syntax, a SyntaxError
//     under `v`).
//   - regexp2 compile catches other structural errors (unclosed `(`, unmatched
//     quantifier under Unicode mode, bad hex escapes, etc.). esregexp has no
//     `v`-flag token, so `v` validation reuses the `u` compile — the two
//     flags share the same non-class grammar.
//   - A narrow u-flag identity-escape check for the handful of escapes
//     regexp2 accepts but ES-u-mode rejects (`\a`, `\9`, …), a scan for the
//     syntax characters regexp2 reads as literals, and a check that every
//     `\k<name>` resolves to a group the pattern declares.
//   - Under `v` alone, a scan of each character class for the
//     ClassSetExpression rules that only the `v` grammar imposes — an
//     unescaped `(`, `-`, `|`, … or a doubled reserved punctuator.
//
// If ANY check fails the pattern is treated as unparsable — matching
// JavaScript's own parse-or-reject behavior.
func IsValidRegexPattern(pattern string, flags RegexFlags) bool {
	if flags.UV() {
		if !IterateRegexCharacterClasses(pattern, flags, func(int, int) {}) {
			return false
		}
	}

	compilePattern := pattern
	compileFlags := ""
	if flags.UV() {
		compileFlags = "u"
	}
	if flags.UnicodeSets {
		// The runtime compiler implements u grammar, where escapes such as
		// `\&` are invalid even inside a class. In v grammar the reserved class
		// punctuators may be escaped, so feed the compiler an equivalent hex
		// escape and leave the original spelling to the v-specific scanner.
		compilePattern = normalizeVClassPunctuatorEscapes(pattern)
	}
	if _, err := esregexp.Compile(compilePattern, compileFlags); err != nil {
		return false
	}

	if flags.UV() {
		if hasInvalidIdentityEscapeForUFlag(pattern) {
			return false
		}
		if hasInvalidSyntaxCharForUFlag(pattern, flags) {
			return false
		}
		if hasUnresolvedNamedBackreferenceForUFlag(pattern) {
			return false
		}
	}
	if flags.UnicodeSets && hasInvalidClassContentForVFlag(pattern) {
		return false
	}
	return true
}

// IsValidRegexPatternForECMAVersion additionally rejects pattern features
// that had not entered the configured ECMAScript edition yet.
func IsValidRegexPatternForECMAVersion(pattern string, flags RegexFlags, ecmaVersion int) bool {
	return IsValidRegexPattern(pattern, flags) && regexFeaturesAvailable(pattern, flags, ecmaVersion)
}

func normalizeVClassPunctuatorEscapes(pattern string) string {
	var result strings.Builder
	result.Grow(len(pattern))
	classDepth := 0
	for i := 0; i < len(pattern); {
		if pattern[i] == '[' {
			classDepth++
			result.WriteByte(pattern[i])
			i++
			continue
		}
		if pattern[i] == ']' && classDepth > 0 {
			classDepth--
			result.WriteByte(pattern[i])
			i++
			continue
		}
		if pattern[i] == '\\' && i+1 < len(pattern) && classDepth > 0 &&
			strings.IndexByte(reservedClassSetPunctuators, pattern[i+1]) >= 0 {
			const hex = "0123456789abcdef"
			c := pattern[i+1]
			result.WriteString(`\x`)
			result.WriteByte(hex[c>>4])
			result.WriteByte(hex[c&0xf])
			i += 2
			continue
		}
		if pattern[i] == '\\' && i+1 < len(pattern) {
			result.WriteString(pattern[i : i+2])
			i += 2
			continue
		}
		result.WriteByte(pattern[i])
		i++
	}
	return result.String()
}

func regexFeaturesAvailable(pattern string, flags RegexFlags, ecmaVersion int) bool {
	for i := 0; i < len(pattern); {
		switch pattern[i] {
		case '\\':
			if i+1 < len(pattern) && ecmaVersion < 2018 &&
				(pattern[i+1] == 'p' || pattern[i+1] == 'P' || pattern[i+1] == 'k') {
				return false
			}
			step, ok := SkipPatternEscape(pattern, i, flags)
			if !ok {
				return false
			}
			i += step
		case '[':
			end, ok := ClassEnd(pattern, i, flags)
			if !ok {
				return false
			}
			if ecmaVersion < 2018 && strings.Contains(pattern[i:end], `\p`) ||
				ecmaVersion < 2018 && strings.Contains(pattern[i:end], `\P`) {
				return false
			}
			i = end
		default:
			if pattern[i] == '(' && i+2 < len(pattern) && pattern[i+1] == '?' {
				if pattern[i+2] == '<' && ecmaVersion < 2018 {
					return false
				}
				if ecmaVersion < 2025 && isInlineModifierGroup(pattern[i+2:]) {
					return false
				}
			}
			i++
		}
	}
	return true
}

func isInlineModifierGroup(suffix string) bool {
	seenFlag := false
	seenDash := false
	for i := range len(suffix) {
		switch suffix[i] {
		case 'i', 'm', 's':
			seenFlag = true
		case '-':
			if seenDash {
				return false
			}
			seenDash = true
		case ':':
			return seenFlag
		default:
			return false
		}
	}
	return false
}

// hasInvalidSyntaxCharForUFlag reports whether the pattern contains an
// unescaped `{`, `}` or `]` outside a character class. Under the u/v flag all
// three are SyntaxCharacters and legal only in their structural roles — `{`
// and `}` as a `{n}` / `{n,}` / `{n,m}` quantifier, `]` as a class terminator
// — but regexp2 accepts each as a literal (its .NET lineage). Escapes and
// character classes are skipped wholesale through the flag-aware scanners, so
// v-flag nested classes are not mistaken for a stray `]`.
func hasInvalidSyntaxCharForUFlag(pattern string, flags RegexFlags) bool {
	i := 0
	for i < len(pattern) {
		switch pattern[i] {
		case '\\':
			step, ok := SkipPatternEscape(pattern, i, flags)
			if !ok {
				// Unterminated escape — let compile/scan report it.
				return false
			}
			i += step
		case '[':
			end, ok := ClassEnd(pattern, i, flags)
			if !ok {
				// Unterminated class — let compile/scan report it.
				return false
			}
			i = end
		case '{':
			end, ok := quantifierEnd(pattern, i)
			if !ok {
				return true
			}
			i = end
		case '}', ']':
			return true
		default:
			i++
		}
	}
	return false
}

// quantifierEnd returns the byte index just past the `}` of the
// `{n}` / `{n,}` / `{n,m}` quantifier opening at pattern[start], or ok=false
// when the brace does not open one.
func quantifierEnd(pattern string, start int) (int, bool) {
	i := start + 1
	digits := 0
	for i < len(pattern) && pattern[i] >= '0' && pattern[i] <= '9' {
		i++
		digits++
	}
	if digits == 0 {
		return start, false
	}
	if i < len(pattern) && pattern[i] == ',' {
		i++
		for i < len(pattern) && pattern[i] >= '0' && pattern[i] <= '9' {
			i++
		}
	}
	if i < len(pattern) && pattern[i] == '}' {
		return i + 1, true
	}
	return start, false
}

// hasInvalidIdentityEscapeForUFlag scans for escapes that ECMAScript u/v mode
// rejects but regexp2 accepts. Conservative: on malformed input it returns
// false (i.e. defers to regexp2's verdict) rather than inventing an error.
func hasInvalidIdentityEscapeForUFlag(pattern string) bool {
	i := 0
	for i < len(pattern) {
		c := pattern[i]
		if c != '\\' || i+1 >= len(pattern) {
			i++
			continue
		}
		next := pattern[i+1]
		switch next {
		// Recognized single-letter escapes.
		case 'd', 'D', 'w', 'W', 's', 'S', 'b', 'B', 'n', 't', 'r', 'v', 'f', '0',
			'x', 'u', 'c', 'p', 'P', 'k', 'q',
			// SyntaxCharacter / `/` — legal identity escapes under u.
			'^', '$', '.', '*', '+', '?', '(', ')', '[', ']', '{', '}', '|', '\\', '/':
			i += 2
			continue
		}
		// Decimal backreference (\1..\9) is legal.
		if next >= '1' && next <= '9' {
			i += 2
			continue
		}
		// Any letter/digit identity escape not recognized above is illegal under u.
		if (next >= 'a' && next <= 'z') || (next >= 'A' && next <= 'Z') {
			return true
		}
		i += 2
	}
	return false
}

// classSetElementKind classifies one element of a v-flag class body for the
// purposes of range validation: only a ClassSetCharacter may sit on either
// side of a `-`.
type classSetElementKind int

const (
	classSetPlainChar classSetElementKind = iota
	classSetOperand
)

// reservedClassSetPunctuators are the characters ECMAScript reserves inside a
// v-flag class when they appear doubled (`!!`, `##`, `..`, …). `&` is in the
// set because `&&` is only legal as the intersection operator.
const reservedClassSetPunctuators = "&!#$%*+,.:;<=>?@^`~"

// hasInvalidClassContentForVFlag reports whether any character class in
// pattern holds something the v-flag ClassSetExpression grammar rejects — an
// unescaped ClassSetSyntaxCharacter, a doubled reserved punctuator, or a `-`
// that isn't a range between two single characters. It is deliberately
// conservative: set operators (`--`, and `&&` without operands on both sides)
// count as invalid, because a pattern this scanner cannot prove legal must
// not be offered the flag.
func hasInvalidClassContentForVFlag(pattern string) bool {
	flags := RegexFlags{UnicodeSets: true}
	invalid := false
	IterateRegexCharacterClasses(pattern, flags, func(start, end int) {
		if !invalid && hasInvalidClassBodyForVFlag(pattern, start, end, flags) {
			invalid = true
		}
	})
	return invalid
}

// hasInvalidClassBodyForVFlag checks the body of the single class spanning
// [start, end). Nested classes are skipped whole — IterateRegexCharacterClasses
// hands each of them to this function in its own right.
func hasInvalidClassBodyForVFlag(pattern string, start, end int, flags RegexFlags) bool {
	i := start + 1
	last := end - 1
	if i < last && pattern[i] == '^' {
		i++
	}

	elements := 0
	lastWasRange := false
	sawOperator := false
	for i < last {
		c := pattern[i]
		// A `-` here has no left-hand character to open a range with.
		if c == '-' {
			return true
		}
		if strings.IndexByte(reservedClassSetPunctuators, c) >= 0 && i+1 < last && pattern[i+1] == c {
			// `&&` is the intersection operator, legal between two operands —
			// and a range is not an operand. Every other doubled punctuator is
			// reserved.
			if c != '&' || elements != 1 || lastWasRange || i+2 >= last || pattern[i+2] == '&' {
				return true
			}
			i += 2
			elements = 0
			sawOperator = true
			continue
		}

		kind, next, ok := classSetElement(pattern, i, last, flags)
		if !ok {
			return true
		}
		i = next
		elements++
		lastWasRange = false

		if i < last && pattern[i] == '-' {
			// `--` is the difference operator, which this scanner doesn't
			// track; refuse rather than read it as a range.
			if kind != classSetPlainChar || sawOperator || i+1 >= last || pattern[i+1] == '-' {
				return true
			}
			i++
			kind, next, ok = classSetElement(pattern, i, last, flags)
			if !ok || kind != classSetPlainChar {
				return true
			}
			i = next
			lastWasRange = true
		}
	}
	return sawOperator && elements != 1
}

// classSetElement consumes one element of a v-flag class body at pattern[i],
// returning its kind and the index just past it. ok is false for the
// ClassSetSyntaxCharacters that must be escaped inside a v-flag class.
func classSetElement(pattern string, i, last int, flags RegexFlags) (classSetElementKind, int, bool) {
	switch c := pattern[i]; c {
	case '\\':
		step, ok := SkipPatternEscape(pattern, i, flags)
		if !ok || i+step > last {
			return classSetOperand, i, false
		}
		switch pattern[i+1] {
		case 'd', 'D', 'w', 'W', 's', 'S', 'p', 'P', 'q':
			return classSetOperand, i + step, true
		}
		return classSetPlainChar, i + step, true
	case '[':
		nestedEnd, ok := ClassEnd(pattern, i, flags)
		if !ok || nestedEnd > last {
			return classSetOperand, i, false
		}
		return classSetOperand, nestedEnd, true
	case '(', ')', '{', '}', '/', '|', ']':
		return classSetOperand, i, false
	}
	_, width := utf8.DecodeRuneInString(pattern[i:])
	if width == 0 {
		width = 1
	}
	return classSetPlainChar, i + width, true
}

// hasUnresolvedNamedBackreferenceForUFlag reports whether the pattern has a
// `\k` that no `(?<name>…)` group in the same pattern can resolve. Without the
// u/v flag such a `\k` reads as the literal characters `k<name>`; with it, an
// unresolved reference is a SyntaxError.
func hasUnresolvedNamedBackreferenceForUFlag(pattern string) bool {
	names := make(map[string]struct{})
	var refs []string
	for i := 0; i < len(pattern); {
		if pattern[i] == '[' {
			end, ok := ClassEnd(pattern, i, RegexFlags{Unicode: true})
			if !ok {
				return false
			}
			i = end
			continue
		}
		if pattern[i] == '\\' {
			if i+1 >= len(pattern) {
				return false
			}
			if pattern[i+1] == 'k' {
				name, next, ok := readAngleName(pattern, i+2)
				if !ok {
					return true
				}
				refs = append(refs, name)
				i = next
				continue
			}
			i += 2
			continue
		}
		// `(?<name>` declares a group; `(?<=` and `(?<!` are lookbehinds.
		if strings.HasPrefix(pattern[i:], "(?<") && i+3 < len(pattern) &&
			pattern[i+3] != '=' && pattern[i+3] != '!' {
			name, next, ok := readAngleName(pattern, i+2)
			normalized, valid := normalizeRegexCaptureName(name)
			if !ok || !valid {
				return true
			}
			if _, duplicate := names[normalized]; duplicate {
				return true
			}
			names[normalized] = struct{}{}
			i = next
			continue
		}
		i++
	}
	for _, ref := range refs {
		normalized, valid := normalizeRegexCaptureName(ref)
		if !valid {
			return true
		}
		if _, exists := names[normalized]; !exists {
			return true
		}
	}
	return false
}

// normalizeRegexCaptureName validates the IdentifierName grammar used by
// named captures and returns its decoded value so escaped and literal spellings
// compare alike.
func normalizeRegexCaptureName(name string) (string, bool) {
	if name == "" {
		return "", false
	}
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
		value, width, ok := decodeRegexCaptureNameEscape(name, i)
		if !ok {
			return "", false
		}
		if value >= 0xD800 && value <= 0xDBFF {
			next, nextWidth, ok := decodeRegexCaptureNameEscape(name, i+width)
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
	normalized := result.String()
	for i, r := range normalized {
		if (i == 0 && !scanner.IsIdentifierStart(r)) || (i != 0 && !scanner.IsIdentifierPart(r)) {
			return "", false
		}
	}
	return normalized, true
}

func decodeRegexCaptureNameEscape(name string, start int) (uint32, int, bool) {
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

// readAngleName reads a `<name>` starting at pattern[start], returning the
// name and the index just past `>`.
func readAngleName(pattern string, start int) (string, int, bool) {
	if start >= len(pattern) || pattern[start] != '<' {
		return "", start, false
	}
	closeRel := strings.IndexByte(pattern[start:], '>')
	if closeRel < 0 {
		return "", start, false
	}
	return pattern[start+1 : start+closeRel], start + closeRel + 1, true
}
