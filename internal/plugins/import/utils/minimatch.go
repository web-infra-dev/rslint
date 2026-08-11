package utils

import (
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"unicode/utf16"

	tsstringutil "github.com/microsoft/typescript-go/shim/stringutil"
	rslintutils "github.com/web-infra-dev/rslint/internal/utils"
)

// MinimatchOptions is the subset of minimatch options that changes pure
// string matching. Filesystem-walking options do not apply to import
// specifiers and are intentionally absent.
type MinimatchOptions struct {
	NoNegate           bool
	NoComment          bool
	NoCase             bool
	MatchBase          bool
	NoGlobStar         bool
	NoExt              bool
	NoBrace            bool
	Dot                bool
	Partial            bool
	FlipNegate         bool
	AllowWindowsEscape bool
}

// Minimatcher is an immutable, reusable minimatch pattern. Compile configured
// patterns once when a rule is initialized instead of rebuilding regular
// expressions for every import specifier.
type Minimatcher struct {
	options        MinimatchOptions
	emptyPattern   bool
	commentPattern bool
	negated        bool
	matchBase      bool
	patterns       []compiledMinimatchPath
}

// CompileMinimatch compiles a minimatch-compatible glob, brace, or extglob
// pattern. Invalid pattern fragments simply never match, as minimatch does.
func CompileMinimatch(pattern string, options MinimatchOptions) *Minimatcher {
	pattern = strings.TrimFunc(pattern, rslintutils.IsStrWhiteSpace)
	pattern = normalizeMinimatchPattern(pattern, runtime.GOOS == "windows", options.AllowWindowsEscape)
	matcher := &Minimatcher{options: options}
	if pattern == "" {
		matcher.emptyPattern = true
		return matcher
	}
	if !options.NoComment && strings.HasPrefix(pattern, "#") {
		matcher.commentPattern = true
		return matcher
	}
	if !options.NoNegate {
		for strings.HasPrefix(pattern, "!") {
			matcher.negated = !matcher.negated
			pattern = pattern[1:]
		}
	}
	matcher.matchBase = options.MatchBase && !strings.Contains(pattern, "/")

	patterns := []string{pattern}
	if !options.NoBrace {
		patterns = expandBracePatterns(pattern)
	}
	matcher.patterns = make([]compiledMinimatchPath, 0, len(patterns))
	for _, candidate := range patterns {
		matcher.patterns = append(matcher.patterns, compileMinimatchPath(candidate, options))
	}
	return matcher
}

// MatchMinimatch matches one name without retaining the compiled pattern.
// Call CompileMinimatch directly when the same pattern is used repeatedly.
func MatchMinimatch(name string, pattern string, options MinimatchOptions) bool {
	return CompileMinimatch(pattern, options).Match(name)
}

// Match reports whether name matches the compiled pattern.
func (matcher *Minimatcher) Match(name string) bool {
	if matcher == nil {
		return false
	}
	if matcher.emptyPattern {
		return name == ""
	}
	if matcher.commentPattern {
		return false
	}
	// minimatch's partial mode treats the filesystem traversal root as a
	// prefix of every non-empty, non-comment pattern.
	if matcher.options.Partial && name == "/" {
		return true
	}
	if matcher.matchBase {
		parts := splitMinimatchPath(name)
		for index := len(parts) - 1; index >= 0; index-- {
			if parts[index] != "" {
				name = parts[index]
				break
			}
		}
	}

	name = encodeMinimatchUTF16(name)
	if matcher.options.NoCase {
		name = strings.Map(jsRegexpCanonicalRune, name)
	}
	parts := splitMinimatchPath(name)
	matched := false
	for index := range matcher.patterns {
		if matcher.patterns[index].match(parts) {
			matched = true
			break
		}
	}
	if matcher.negated {
		if matcher.options.FlipNegate {
			return matched
		}
		return !matched
	}
	return matched
}

func normalizeMinimatchPattern(pattern string, windows, allowWindowsEscape bool) string {
	// minimatch 3 treats the host separator as `/` on Windows unless the
	// caller explicitly asks to preserve backslashes as glob escapes.
	if windows && !allowWindowsEscape {
		return strings.ReplaceAll(pattern, `\`, "/")
	}
	return pattern
}

type compiledMinimatchPath struct {
	options MinimatchOptions
	parts   []compiledMinimatchSegment
}

type compiledMinimatchSegment struct {
	pattern   string
	globstar  bool
	dot       bool
	regexp    *regexp.Regexp
	negatives []compiledNegativeExtglob
}

type compiledNegativeExtglob struct {
	group        int
	alternatives []compiledMinimatchPath
}

func compileMinimatchPath(pattern string, options MinimatchOptions) compiledMinimatchPath {
	// minimatch 3 builds a non-Unicode JavaScript RegExp, so `?` and character
	// classes consume UTF-16 code units rather than Unicode code points. Map
	// surrogate units to private-use runes before using Go's rune-based regexp
	// engine; ordinary BMP syntax (including `/`, `*`, and brackets) stays
	// unchanged and can still be parsed below.
	pattern = encodeMinimatchUTF16(pattern)
	return compileEncodedMinimatchPath(pattern, options)
}

func compileEncodedMinimatchPath(pattern string, options MinimatchOptions) compiledMinimatchPath {
	patterns := splitMinimatchPath(pattern)
	if !options.NoGlobStar {
		patterns = collapseConsecutiveGlobstars(patterns)
	}
	compiled := compiledMinimatchPath{
		options: options,
		parts:   make([]compiledMinimatchSegment, len(patterns)),
	}
	for index, part := range patterns {
		if part == "**" && !options.NoGlobStar {
			compiled.parts[index] = compiledMinimatchSegment{pattern: part, globstar: true}
			continue
		}
		compiled.parts[index] = compileMinimatchSegment(part, options)
	}
	return compiled
}

func compileMinimatchSegment(pattern string, options MinimatchOptions) compiledMinimatchSegment {
	segment := compiledMinimatchSegment{pattern: pattern, dot: options.Dot}
	builder := globRegexBuilder{options: options}
	builder.expression.WriteByte('^')
	builder.writeFragment([]rune(pattern))
	builder.expression.WriteByte('$')
	segment.regexp, _ = regexp.Compile(builder.expression.String())
	segment.negatives = make([]compiledNegativeExtglob, 0, len(builder.negatives))
	for _, negative := range builder.negatives {
		alternativeOptions := options
		alternativeOptions.NoNegate = true
		alternativeOptions.NoComment = true
		alternativeOptions.MatchBase = false
		compiled := compiledNegativeExtglob{
			group:        negative.group,
			alternatives: make([]compiledMinimatchPath, 0, len(negative.alternatives)),
		}
		for _, alternative := range negative.alternatives {
			compiled.alternatives = append(compiled.alternatives, compileEncodedMinimatchPath(alternative, alternativeOptions))
		}
		segment.negatives = append(segment.negatives, compiled)
	}
	return segment
}

func (pattern *compiledMinimatchPath) match(names []string) bool {
	width := len(names) + 1
	memo := make([]uint8, (len(pattern.parts)+1)*width)
	remember := func(key int, result bool) bool {
		if result {
			memo[key] = 2
		} else {
			memo[key] = 1
		}
		return result
	}
	var match func(int, int) bool
	match = func(patternIndex, nameIndex int) bool {
		key := patternIndex*width + nameIndex
		if memo[key] != 0 {
			return memo[key] == 2
		}

		if patternIndex == len(pattern.parts) {
			return remember(key, nameIndex == len(names) ||
				nameIndex == len(names)-1 && names[nameIndex] == "")
		}
		if pattern.options.Partial && nameIndex == len(names) {
			return remember(key, true)
		}
		part := &pattern.parts[patternIndex]
		if part.globstar {
			// A trailing `/**` still requires the slash to exist (`foo/**`
			// does not match `foo`), while a leading/bare globstar can match
			// zero segments.
			canMatchZero := patternIndex == 0 || patternIndex+1 < len(pattern.parts) ||
				nameIndex > patternIndex
			if canMatchZero && match(patternIndex+1, nameIndex) {
				return remember(key, true)
			}
			if nameIndex < len(names) && globstarMayConsume(names[nameIndex], pattern.options.Dot) && match(patternIndex, nameIndex+1) {
				return remember(key, true)
			}
			return remember(key, false)
		}
		if nameIndex >= len(names) || !part.match(names[nameIndex]) {
			return remember(key, false)
		}
		return remember(key, match(patternIndex+1, nameIndex+1))
	}
	return match(0, 0)
}

const minimatchSurrogatePlaceholderBase = rune(0xF0000)

func encodeMinimatchUTF16(value string) string {
	needsEncoding := false
	for offset := 0; offset < len(value); {
		char, size := tsstringutil.DecodeJSStringRune(value[offset:])
		offset += size
		if char >= 0xD800 && char <= 0xDFFF || char > 0xFFFF {
			needsEncoding = true
			break
		}
	}
	if !needsEncoding {
		return value
	}

	var result strings.Builder
	result.Grow(len(value))
	for offset := 0; offset < len(value); {
		char, size := tsstringutil.DecodeJSStringRune(value[offset:])
		offset += size
		switch {
		case char >= 0xD800 && char <= 0xDFFF:
			result.WriteRune(minimatchSurrogatePlaceholderBase + char - 0xD800)
		case char > 0xFFFF:
			high, low := utf16.EncodeRune(char)
			result.WriteRune(minimatchSurrogatePlaceholderBase + high - 0xD800)
			result.WriteRune(minimatchSurrogatePlaceholderBase + low - 0xD800)
		default:
			result.WriteRune(char)
		}
	}
	return result.String()
}

func collapseConsecutiveGlobstars(patterns []string) []string {
	result := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		if pattern == "**" && len(result) > 0 && result[len(result)-1] == "**" {
			continue
		}
		result = append(result, pattern)
	}
	return result
}

func globstarMayConsume(segment string, dot bool) bool {
	if segment == "." || segment == ".." {
		return false
	}
	return dot || segment == "" || !strings.HasPrefix(segment, ".")
}

func (segment *compiledMinimatchSegment) match(name string) bool {
	// minimatch prefixes every non-empty magical segment regexp with (?=.) so
	// `*` and zero-width extglobs do not match an empty path component.
	if name == "" && segment.pattern != "" {
		return false
	}
	if strings.HasPrefix(name, ".") && !patternExplicitlyStartsWithDot(segment.pattern) {
		if name == "." || name == ".." || !segment.dot {
			return false
		}
	}
	if segment.regexp == nil {
		return false
	}
	matches := segment.regexp.FindStringSubmatch(name)
	matched := matches != nil
	if matched {
		for _, negative := range segment.negatives {
			if negative.group >= len(matches) {
				continue
			}
			candidate := matches[negative.group]
			candidateParts := splitMinimatchPath(candidate)
			for index := range negative.alternatives {
				if negative.alternatives[index].match(candidateParts) {
					matched = false
					break
				}
			}
			if !matched {
				break
			}
		}
	}
	return matched
}

func patternExplicitlyStartsWithDot(pattern string) bool {
	return strings.HasPrefix(pattern, ".")
}

// splitMinimatchPath mirrors JavaScript's split(/\/+/): adjacent separators
// collapse, while a leading or trailing separator still produces an empty
// path component.
func splitMinimatchPath(path string) []string {
	result := make([]string, 0, strings.Count(path, "/")+1)
	start := 0
	for index := 0; index < len(path); {
		if path[index] != '/' {
			index++
			continue
		}
		result = append(result, path[start:index])
		for index < len(path) && path[index] == '/' {
			index++
		}
		start = index
	}
	return append(result, path[start:])
}

// These limits mirror brace-expansion 1.x, which minimatch 3 uses to keep a
// configured path group from allocating an unbounded Cartesian product.
const (
	minimatchBraceExpansionMax       = 100_000
	minimatchBraceExpansionMaxLength = 4_000_000
)

func expandBracePatterns(pattern string) []string {
	result := make([]string, 0)
	totalLength := 0
	expandBracePatternInto(pattern, &result, &totalLength)
	return result
}

func expandBracePatternInto(pattern string, result *[]string, totalLength *int) bool {
	runes := []rune(pattern)
	for index, char := range runes {
		if char != '{' || runeIsEscaped(runes, index) {
			continue
		}
		end, ok := matchingDelimiter(runes, index+1, '{', '}')
		if !ok {
			return appendBraceExpansion(pattern, result, totalLength)
		}
		alternatives := splitGlobAlternatives(runes[index+1:end], ',')
		if len(alternatives) <= 1 {
			if sequence, ok := expandBraceSequence(alternatives[0]); ok {
				alternatives = make([][]rune, len(sequence))
				for i, value := range sequence {
					alternatives[i] = []rune(value)
				}
			} else {
				continue
			}
		}
		prefix := string(runes[:index])
		suffix := string(runes[end+1:])
		for _, alternative := range alternatives {
			if !expandBracePatternInto(prefix+string(alternative)+suffix, result, totalLength) {
				return false
			}
		}
		return true
	}
	return appendBraceExpansion(pattern, result, totalLength)
}

func appendBraceExpansion(pattern string, result *[]string, totalLength *int) bool {
	length := minimatchUTF16Length(pattern)
	if len(*result) >= minimatchBraceExpansionMax ||
		*totalLength+length > minimatchBraceExpansionMaxLength {
		return false
	}
	*result = append(*result, pattern)
	*totalLength += length
	return true
}

func minimatchUTF16Length(value string) int {
	length := 0
	for offset := 0; offset < len(value); {
		char, size := tsstringutil.DecodeJSStringRune(value[offset:])
		offset += size
		length++
		if char > 0xFFFF {
			length++
		}
	}
	return length
}

// expandBraceSequence implements brace-expansion's numeric and ASCII letter
// ranges, including descending ranges, optional steps, and zero padding.
func expandBraceSequence(fragment []rune) ([]string, bool) {
	parts := strings.Split(string(fragment), "..")
	if len(parts) != 2 && len(parts) != 3 {
		return nil, false
	}
	step := uint(1)
	if len(parts) == 3 {
		parsed, ok := parseBraceInteger(parts[2])
		if !ok {
			return nil, false
		}
		if parsed != 0 {
			// Compute the magnitude in unsigned space so MinInt is representable.
			if parsed < 0 {
				step = uint(-(parsed + 1)) + 1
			} else {
				step = uint(parsed)
			}
		}
	}

	startNumber, startIsNumber := parseBraceInteger(parts[0])
	endNumber, endIsNumber := parseBraceInteger(parts[1])
	if startIsNumber && endIsNumber {
		width := max(decimalWidth(parts[0]), decimalWidth(parts[1]))
		padded := hasLeadingZero(parts[0]) || hasLeadingZero(parts[1])
		return expandIntegerSequence(startNumber, endNumber, step, width, padded), true
	}

	startRunes, endRunes := []rune(parts[0]), []rune(parts[1])
	if len(startRunes) != 1 || len(endRunes) != 1 ||
		!isASCIILetter(startRunes[0]) || !isASCIILetter(endRunes[0]) {
		return nil, false
	}
	return expandRuneSequence(startRunes[0], endRunes[0], step), true
}

func parseBraceInteger(value string) (int, bool) {
	if value == "" {
		return 0, false
	}
	digits := value
	if digits[0] == '-' {
		digits = digits[1:]
	}
	if digits == "" {
		return 0, false
	}
	for _, char := range digits {
		if char < '0' || char > '9' {
			return 0, false
		}
	}
	result, err := strconv.Atoi(value)
	return result, err == nil
}

func decimalWidth(value string) int {
	return len(strings.TrimPrefix(value, "-"))
}

func hasLeadingZero(value string) bool {
	digits := strings.TrimPrefix(value, "-")
	return len(digits) > 1 && digits[0] == '0'
}

func expandIntegerSequence(start, end int, step uint, width int, padded bool) []string {
	direction := 1
	if end < start {
		direction = -1
	}
	var result []string
	totalLength := 0
	for value := start; ; {
		text := strconv.Itoa(value)
		if padded {
			sign := ""
			digits := text
			if strings.HasPrefix(digits, "-") {
				sign, digits = "-", digits[1:]
			}
			text = sign + strings.Repeat("0", max(0, width-len(digits))) + digits
		}
		length := minimatchUTF16Length(text)
		if len(result) >= minimatchBraceExpansionMax ||
			totalLength+length > minimatchBraceExpansionMaxLength {
			break
		}
		result = append(result, text)
		totalLength += length
		if value == end {
			break
		}
		// Unsigned subtraction gives the exact distance even when the signed
		// endpoints span the whole machine-int range.
		var distance uint
		if direction > 0 {
			distance = uint(end) - uint(value)
		} else {
			distance = uint(value) - uint(end)
		}
		if step > distance {
			break
		}
		if direction > 0 {
			value = int(uint(value) + step)
		} else {
			value = int(uint(value) - step)
		}
	}
	return result
}

func expandRuneSequence(start, end rune, step uint) []string {
	direction := rune(1)
	if end < start {
		direction = -1
	}
	distance := int64(end - start)
	if distance < 0 {
		distance = -distance
	}
	if step > uint(distance) {
		return []string{string(start)}
	}
	runeStep := rune(step)
	var result []string
	totalLength := 0
	for value := start; ; value += direction * runeStep {
		text := string(value)
		length := minimatchUTF16Length(text)
		if len(result) >= minimatchBraceExpansionMax ||
			totalLength+length > minimatchBraceExpansionMaxLength {
			break
		}
		result = append(result, text)
		totalLength += length
		if value == end || direction > 0 && value+runeStep > end || direction < 0 && value-runeStep < end {
			break
		}
	}
	return result
}

func isASCIILetter(char rune) bool {
	return char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z'
}

type negativeExtglob struct {
	group        int
	alternatives []string
}

type globRegexBuilder struct {
	options    MinimatchOptions
	expression strings.Builder
	groups     int
	negatives  []negativeExtglob
}

func (builder *globRegexBuilder) openGroup() int {
	builder.groups++
	builder.expression.WriteByte('(')
	return builder.groups
}

func (builder *globRegexBuilder) writeFragment(fragment []rune) {
	for index := 0; index < len(fragment); {
		char := fragment[index]
		if !builder.options.NoExt && index+1 < len(fragment) && fragment[index+1] == '(' && strings.ContainsRune("?*+@!", char) {
			if end, ok := matchingDelimiter(fragment, index+2, '(', ')'); ok {
				alternatives := splitGlobAlternatives(fragment[index+2:end], '|')
				if char == '!' {
					group := builder.openGroup()
					builder.expression.WriteString("[^/]*")
					builder.expression.WriteByte(')')
					raw := make([]string, len(alternatives))
					for i, alternative := range alternatives {
						raw[i] = string(alternative)
					}
					builder.negatives = append(builder.negatives, negativeExtglob{group: group, alternatives: raw})
				} else {
					builder.openGroup()
					for i, alternative := range alternatives {
						if i > 0 {
							builder.expression.WriteByte('|')
						}
						builder.writeFragment(alternative)
					}
					builder.expression.WriteByte(')')
					switch char {
					case '?', '*', '+':
						builder.expression.WriteRune(char)
					}
				}
				index = end + 1
				continue
			}
		}
		if char == '{' && !builder.options.NoBrace {
			if end, ok := matchingDelimiter(fragment, index+1, '{', '}'); ok {
				alternatives := splitGlobAlternatives(fragment[index+1:end], ',')
				if len(alternatives) > 1 {
					builder.openGroup()
					for i, alternative := range alternatives {
						if i > 0 {
							builder.expression.WriteByte('|')
						}
						builder.writeFragment(alternative)
					}
					builder.expression.WriteByte(')')
					index = end + 1
					continue
				}
			}
		}

		switch char {
		case '*':
			for index < len(fragment) && fragment[index] == '*' {
				index++
			}
			builder.expression.WriteString("[^/]*")
		case '?':
			builder.expression.WriteString("[^/]")
			index++
		case '[':
			end := index + 1
			for end < len(fragment) && fragment[end] != ']' {
				end++
			}
			if end >= len(fragment) || end == index+1 {
				builder.expression.WriteString(`\[`)
				index++
				continue
			}
			bodyRunes := fragment[index+1 : end]
			body := string(bodyRunes)
			if strings.HasPrefix(body, "!") {
				body = "^" + body[1:]
			}
			if builder.options.NoCase {
				body = caseFoldCharacterClass(body)
			}
			builder.expression.WriteByte('[')
			builder.expression.WriteString(body)
			builder.expression.WriteByte(']')
			index = end + 1
		case '\\':
			if index+1 < len(fragment) {
				builder.writeLiteral(fragment[index+1])
				index += 2
			} else {
				builder.expression.WriteString(`\\`)
				index++
			}
		default:
			builder.writeLiteral(char)
			index++
		}
	}
}

func (builder *globRegexBuilder) writeLiteral(char rune) {
	if builder.options.NoCase {
		char = jsRegexpCanonicalRune(char)
	}
	builder.expression.WriteString(regexp.QuoteMeta(string(char)))
}

// jsRegexpCanonicalRune mirrors ECMAScript's non-Unicode RegExp
// Canonicalize operation. In particular, a non-ASCII rune whose uppercase
// mapping is ASCII (Kelvin sign, long s, dotless i) does not fold to ASCII.
func jsRegexpCanonicalRune(char rune) rune {
	if char >= 'a' && char <= 'z' {
		return char - ('a' - 'A')
	}
	if char < 0x80 {
		return char
	}
	upperText := tsstringutil.ToUpperJS(string(char))
	upper, size := tsstringutil.DecodeJSStringRune(upperText)
	// Non-Unicode ECMAScript regexps only adopt one-code-unit uppercase
	// mappings; full mappings such as ß -> SS leave the input unchanged.
	if size != len(upperText) || upper > 0xFFFF {
		return char
	}
	if char >= 0x80 && upper < 0x80 {
		return char
	}
	return upper
}

func caseFoldCharacterClass(body string) string {
	prefix := ""
	content := body
	if strings.HasPrefix(content, "^") {
		prefix, content = "^", content[1:]
	}
	canonical := []rune(content)
	for index, char := range canonical {
		switch char {
		case '^', '-', '\\':
			continue
		default:
			canonical[index] = jsRegexpCanonicalRune(char)
		}
	}
	canonicalText := string(canonical)
	if canonicalText == content {
		return body
	}
	// Preserve the original range as well as its canonicalized counterpart.
	// This matters for broad ranges such as [A-z], whose punctuation members
	// must not disappear when z canonicalizes to Z.
	return prefix + content + canonicalText
}

func matchingDelimiter(fragment []rune, start int, open rune, closing rune) (int, bool) {
	depth := 1
	for index := start; index < len(fragment); index++ {
		if runeIsEscaped(fragment, index) {
			continue
		}
		switch fragment[index] {
		case open:
			depth++
		case closing:
			depth--
			if depth == 0 {
				return index, true
			}
		}
	}
	return -1, false
}

func splitGlobAlternatives(fragment []rune, separator rune) [][]rune {
	var alternatives [][]rune
	start := 0
	parenDepth := 0
	braceDepth := 0
	inClass := false
	for index, char := range fragment {
		if runeIsEscaped(fragment, index) {
			continue
		}
		if char == '[' {
			inClass = true
			continue
		}
		if char == ']' && inClass {
			inClass = false
			continue
		}
		if inClass {
			continue
		}
		switch char {
		case '(':
			parenDepth++
		case ')':
			parenDepth--
		case '{':
			braceDepth++
		case '}':
			braceDepth--
		default:
			if char == separator && parenDepth == 0 && braceDepth == 0 {
				alternatives = append(alternatives, fragment[start:index])
				start = index + 1
			}
		}
	}
	return append(alternatives, fragment[start:])
}

func runeIsEscaped(fragment []rune, index int) bool {
	backslashes := 0
	for index--; index >= 0 && fragment[index] == '\\'; index-- {
		backslashes++
	}
	return backslashes%2 == 1
}
