package utils

import (
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

// IsValidRegexPattern reports whether pattern can be proven to parse cleanly
// under flags, the way JavaScript's RegExp constructor would (a try/catch
// around parsing, not a match attempt).
//
// Legacy patterns stay on the regexp runtime used by existing callers. Under
// u/v, the TypeScript scanner is the grammar authority. The small adapter in
// front of it carries constructor pattern text through a RegExp literal
// without changing its UTF-16 semantics and replaces capture names with safe
// ASCII placeholders because the scanner's general identifier reader does not
// accept every RegExpIdentifierName escape spelling. Known scanner gaps fail
// closed so callers never turn an invalid pattern into a suggested fix.
func IsValidRegexPattern(pattern string, flags RegexFlags) bool {
	valid, hasDuplicateCaptureName, hasConservativeVCase := validateRegexPattern(pattern, flags)
	return valid && !hasDuplicateCaptureName && !hasConservativeVCase
}

// IsValidRegexPatternForECMAVersion additionally rejects pattern features
// that had not entered the configured ECMAScript edition yet.
func IsValidRegexPatternForECMAVersion(pattern string, flags RegexFlags, ecmaVersion int) bool {
	valid, hasDuplicateCaptureName, hasConservativeVCase := validateRegexPattern(pattern, flags)
	if !valid || hasDuplicateCaptureName || hasConservativeVCase {
		return false
	}
	return regexFeaturesAvailable(pattern, flags, ecmaVersion)
}

func validateRegexPattern(pattern string, flags RegexFlags) (
	valid bool,
	hasDuplicateCaptureName bool,
	hasConservativeVCase bool,
) {
	if !flags.UV() {
		_, err := esregexp.Compile(pattern, "")
		return err == nil, false, false
	}
	if flags.Unicode && flags.UnicodeSets {
		return false, false, false
	}
	// JavaScript strings are sequences of UTF-16 code units. The compiler keeps
	// lone surrogates in WTF-8, so join an adjacent pair before identifier-name
	// validation just as the RegExp parser's CodePoint operation does.
	pattern = ecmascript.CombineSurrogatePairs(pattern)

	normalized, hasDuplicateCaptureName, hasConservativeVCase, ok := normalizeRegexCaptureNames(pattern, flags)
	if !ok {
		return false, hasDuplicateCaptureName, hasConservativeVCase
	}
	literal, ok := regexPatternLiteral(normalized, flags)
	if !ok {
		return false, hasDuplicateCaptureName, hasConservativeVCase
	}
	return ecmascript.IsValidRegexLiteral(literal), hasDuplicateCaptureName, hasConservativeVCase
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
			if ecmaVersion < 2018 && (strings.Contains(pattern[i:end], `\p`) ||
				strings.Contains(pattern[i:end], `\P`)) {
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
			_, width := utf8.DecodeRuneInString(pattern[i:])
			if width == 0 {
				width = 1
			}
			i += width
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

type regexCaptureNameReplacement struct {
	start       int
	end         int
	placeholder string
}

// normalizeRegexCaptureNames rewrites every declaration and reference name to
// a deterministic ASCII placeholder keyed by its decoded IdentifierName. This
// preserves equality between raw and escaped spellings while allowing tsgo's
// complete RegExp parser to validate the surrounding grammar. It also records
// duplicate declarations for the fail-closed suggestion gate.
func normalizeRegexCaptureNames(pattern string, flags RegexFlags) (string, bool, bool, bool) {
	var replacements []regexCaptureNameReplacement
	var placeholders map[string]string
	var declarations map[string]bool
	hasDuplicate := false
	hasConservativeVCase := false

	addName := func(nameStart int, nameEnd int, declaration bool) bool {
		name, ok := normalizeRegexCaptureName(pattern[nameStart:nameEnd])
		if !ok {
			return false
		}
		if placeholders == nil {
			placeholders = make(map[string]string)
		}
		placeholder, ok := placeholders[name]
		if !ok {
			placeholder = "rslint" + strconv.Itoa(len(placeholders))
			placeholders[name] = placeholder
		}
		if declaration {
			if declarations == nil {
				declarations = make(map[string]bool)
			}
			if declarations[name] {
				hasDuplicate = true
			}
			declarations[name] = true
		}
		replacements = append(replacements, regexCaptureNameReplacement{
			start:       nameStart,
			end:         nameEnd,
			placeholder: placeholder,
		})
		return true
	}

	for i := 0; i < len(pattern); {
		switch pattern[i] {
		case '[':
			end := 0
			if flags.UnicodeSets {
				var content regexVClassContent
				var ok bool
				end, content, ok = analyzeRegexVClass(pattern, i)
				if !ok {
					return "", hasDuplicate, hasConservativeVCase, false
				}
				hasConservativeVCase = hasConservativeVCase || content.hasNegatedStringOperand
			} else {
				var ok bool
				end, ok = ClassEnd(pattern, i, flags)
				if !ok {
					return "", hasDuplicate, hasConservativeVCase, false
				}
			}
			i = end
		case '\\':
			if strings.HasPrefix(pattern[i:], `\k<`) {
				_, next, ok := readAngleName(pattern, i+2)
				if !ok || !addName(i+3, next-1, false) {
					return "", hasDuplicate, hasConservativeVCase, false
				}
				i = next
				continue
			}
			step, ok := SkipPatternEscape(pattern, i, flags)
			if !ok {
				return "", hasDuplicate, hasConservativeVCase, false
			}
			i += step
		case '(':
			if strings.HasPrefix(pattern[i:], "(?<") &&
				(i+3 >= len(pattern) || pattern[i+3] != '=' && pattern[i+3] != '!') {
				_, next, ok := readAngleName(pattern, i+2)
				if !ok || !addName(i+3, next-1, true) {
					return "", hasDuplicate, hasConservativeVCase, false
				}
				i = next
				continue
			}
			i++
		default:
			_, width := utf8.DecodeRuneInString(pattern[i:])
			if width == 0 {
				width = 1
			}
			i += width
		}
	}

	if len(replacements) == 0 {
		return pattern, hasDuplicate, hasConservativeVCase, true
	}
	var result strings.Builder
	result.Grow(len(pattern))
	last := 0
	for _, replacement := range replacements {
		result.WriteString(pattern[last:replacement.start])
		result.WriteString(replacement.placeholder)
		last = replacement.end
	}
	result.WriteString(pattern[last:])
	return result.String(), hasDuplicate, hasConservativeVCase, true
}

type regexVClassContent struct {
	mayContainStrings       bool
	hasNegatedStringOperand bool
}

type regexVClassFrame struct {
	content        regexVClassContent
	negated        bool
	trailingHyphen bool
}

// analyzeRegexVClass performs the narrow v-mode checks needed before carrying
// a constructor pattern through a literal. One iterative pass covers nested
// classes, raw slashes, tsgo's trailing-hyphen and doubled-^/$ gaps, and the
// conservative string-operand bit used only by the suggestion gate. It is a
// lexical adapter, not a second ClassSetExpression parser; tsgo remains the
// grammar authority for every other production.
func analyzeRegexVClass(pattern string, start int) (end int, content regexVClassContent, ok bool) {
	if start >= len(pattern) || pattern[start] != '[' {
		return start, regexVClassContent{}, false
	}

	flags := RegexFlags{UnicodeSets: true}
	frames := []regexVClassFrame{{}}
	i := start + 1
	if i < len(pattern) && pattern[i] == '^' {
		frames[0].negated = true
		i++
	}

	for i < len(pattern) {
		frame := &frames[len(frames)-1]
		switch pattern[i] {
		case '\\':
			if strings.HasPrefix(pattern[i:], `\q{`) {
				next, qOK := scanRegexVQString(pattern, i)
				if !qOK {
					return start, regexVClassContent{}, false
				}
				frame.content.mayContainStrings = true
				frame.trailingHyphen = false
				i = next
				continue
			}
			if strings.HasPrefix(pattern[i:], `\p{`) || strings.HasPrefix(pattern[i:], `\P{`) {
				frame.content.mayContainStrings = true
			}
			step, ok := SkipPatternEscape(pattern, i, flags)
			if !ok {
				return start, regexVClassContent{}, false
			}
			frame.trailingHyphen = false
			i += step
		case '[':
			frame.trailingHyphen = false
			frames = append(frames, regexVClassFrame{})
			i++
			if i < len(pattern) && pattern[i] == '^' {
				frames[len(frames)-1].negated = true
				i++
			}
		case ']':
			if frame.trailingHyphen {
				return start, regexVClassContent{}, false
			}
			frame.content.hasNegatedStringOperand = frame.content.hasNegatedStringOperand ||
				frame.negated && frame.content.mayContainStrings
			completed := frame.content
			frames = frames[:len(frames)-1]
			i++
			if len(frames) == 0 {
				return i, completed, true
			}
			parent := &frames[len(frames)-1]
			parent.content.mayContainStrings = parent.content.mayContainStrings || completed.mayContainStrings
			parent.content.hasNegatedStringOperand = parent.content.hasNegatedStringOperand || completed.hasNegatedStringOperand
			parent.trailingHyphen = false
		case '/':
			// `/` is a ClassSetSyntaxCharacter. Escaping it only in the
			// carrier would make an invalid constructor pattern valid.
			return start, regexVClassContent{}, false
		case '^', '$':
			if i+1 < len(pattern) && pattern[i+1] == pattern[i] {
				return start, regexVClassContent{}, false
			}
			frame.trailingHyphen = false
			i++
		case '-':
			frame.trailingHyphen = true
			i++
		default:
			frame.trailingHyphen = false
			_, width := utf8.DecodeRuneInString(pattern[i:])
			if width == 0 {
				width = 1
			}
			i += width
		}
	}
	return start, regexVClassContent{}, false
}

func scanRegexVQString(pattern string, start int) (next int, ok bool) {
	flags := RegexFlags{UnicodeSets: true}
	for i := start + 3; i < len(pattern); {
		switch pattern[i] {
		case '\\':
			step, ok := SkipPatternEscape(pattern, i, flags)
			if !ok {
				return 0, false
			}
			i += step
		case '}':
			return i + 1, true
		case '/':
			return 0, false
		case '^', '$':
			if i+1 < len(pattern) && pattern[i+1] == pattern[i] {
				return 0, false
			}
			i++
		default:
			_, width := utf8.DecodeRuneInString(pattern[i:])
			if width == 0 {
				width = 1
			}
			i += width
		}
	}
	return 0, false
}

// regexPatternLiteral transports a u/v constructor pattern into a literal for
// tsgo's parser. It reads JavaScript UTF-16 code units, escapes only lexical
// delimiters, and rejects cases where encoding a line terminator or surrogate
// would turn an invalid identity escape into valid syntax.
func regexPatternLiteral(pattern string, flags RegexFlags) (string, bool) {
	pattern = ecmascript.CombineSurrogatePairs(pattern)
	units := ecmascript.StringCodeUnits(pattern)
	if ecmascript.StringFromCodeUnits(units) != pattern {
		return "", false
	}

	var literal strings.Builder
	literal.Grow(len(pattern) + 4)
	literal.WriteByte('/')
	if len(units) == 0 {
		literal.WriteString("(?:)")
	}

	escaped := false
	for _, unit := range units {
		if unit == '\\' {
			literal.WriteByte('\\')
			escaped = !escaped
			continue
		}

		switch unit {
		case '/':
			if !escaped {
				literal.WriteByte('\\')
			}
			literal.WriteByte('/')
		case '\n':
			if escaped {
				return "", false
			}
			literal.WriteString(`\n`)
		case '\r':
			if escaped {
				return "", false
			}
			literal.WriteString(`\r`)
		case 0x2028, 0x2029:
			if escaped {
				return "", false
			}
			writeRegexUnicodeEscape(&literal, unit)
		default:
			if unit >= 0xD800 && unit <= 0xDFFF {
				if escaped {
					return "", false
				}
				writeRegexUnicodeEscape(&literal, unit)
			} else {
				literal.WriteRune(rune(unit))
			}
		}
		escaped = false
	}
	if escaped {
		return "", false
	}

	literal.WriteByte('/')
	if flags.UnicodeSets {
		literal.WriteByte('v')
	} else {
		literal.WriteByte('u')
	}
	return literal.String(), true
}

func writeRegexUnicodeEscape(builder *strings.Builder, unit uint16) {
	const hex = "0123456789ABCDEF"
	builder.WriteString(`\u`)
	builder.WriteByte(hex[unit>>12])
	builder.WriteByte(hex[unit>>8&0xF])
	builder.WriteByte(hex[unit>>4&0xF])
	builder.WriteByte(hex[unit&0xF])
}

// normalizeRegexCaptureName validates the RegExpIdentifierName grammar and
// returns its decoded value so escaped and literal spellings compare alike.
func normalizeRegexCaptureName(name string) (string, bool) {
	if name == "" {
		return "", false
	}
	var result strings.Builder
	first := true
	for i := 0; i < len(name); {
		var value rune
		if name[i] != '\\' {
			r, width := utf8.DecodeRuneInString(name[i:])
			if r == utf8.RuneError && width == 1 {
				return "", false
			}
			value = r
			i += width
		} else {
			decoded, width, fixed, ok := decodeRegexCaptureNameEscape(name, i)
			if !ok {
				return "", false
			}
			i += width
			if decoded >= 0xD800 && decoded <= 0xDBFF {
				if !fixed {
					return "", false
				}
				low, lowWidth, lowFixed, ok := decodeRegexCaptureNameEscape(name, i)
				if !ok || !lowFixed || low < 0xDC00 || low > 0xDFFF {
					return "", false
				}
				value = utf16.DecodeRune(rune(decoded), rune(low))
				i += lowWidth
			} else {
				if decoded >= 0xDC00 && decoded <= 0xDFFF {
					return "", false
				}
				value = rune(decoded)
			}
		}

		if first {
			if !scanner.IsIdentifierStart(value) {
				return "", false
			}
			first = false
		} else if !scanner.IsIdentifierPart(value) {
			return "", false
		}
		result.WriteRune(value)
	}
	return result.String(), !first
}

// decodeRegexCaptureNameEscape decodes one `\u` escape. fixed is true only
// for the four-digit form; the grammar permits a surrogate pair only when both
// halves use that fixed form.
func decodeRegexCaptureNameEscape(name string, start int) (value uint32, width int, fixed bool, ok bool) {
	if start+2 >= len(name) || name[start] != '\\' || name[start+1] != 'u' {
		return 0, 0, false, false
	}
	if name[start+2] != '{' {
		if start+6 > len(name) || !allHexStr(name[start+2:start+6]) {
			return 0, 0, false, false
		}
		parsed, err := strconv.ParseUint(name[start+2:start+6], 16, 16)
		if err != nil {
			return 0, 0, false, false
		}
		return uint32(parsed), 6, true, true
	}

	value = 0
	digits := 0
	for i := start + 3; i < len(name) && name[i] != '}'; i++ {
		digit, valid := regexHexValue(name[i])
		if !valid || value > (utf8.MaxRune-digit)/16 {
			return 0, 0, false, false
		}
		value = value*16 + digit
		digits++
	}
	end := start + 3 + digits
	if digits == 0 || end >= len(name) || name[end] != '}' {
		return 0, 0, false, false
	}
	return value, end - start + 1, false, true
}

func regexHexValue(value byte) (uint32, bool) {
	switch {
	case value >= '0' && value <= '9':
		return uint32(value - '0'), true
	case value >= 'a' && value <= 'f':
		return uint32(value-'a') + 10, true
	case value >= 'A' && value <= 'F':
		return uint32(value-'A') + 10, true
	default:
		return 0, false
	}
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
