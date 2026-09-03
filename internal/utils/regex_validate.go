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
	return isValidRegexPattern(pattern, flags, 2025)
}

// IsValidRegexPatternForECMAVersion additionally rejects pattern features
// that had not entered the configured ECMAScript edition yet.
func IsValidRegexPatternForECMAVersion(pattern string, flags RegexFlags, ecmaVersion int) bool {
	if !isValidRegexPattern(pattern, flags, ecmaVersion) {
		return false
	}
	return regexFeaturesAvailable(pattern, flags, ecmaVersion)
}

type regexPatternValidation struct {
	valid                               bool
	hasDuplicateCaptureName             bool
	hasNonExclusiveDuplicateCaptureName bool
}

func isValidRegexPattern(pattern string, flags RegexFlags, ecmaVersion int) bool {
	validation := validateRegexPattern(pattern, flags)
	if !validation.valid || validation.hasNonExclusiveDuplicateCaptureName {
		return false
	}
	return ecmaVersion >= 2025 || !validation.hasDuplicateCaptureName
}

func validateRegexPattern(pattern string, flags RegexFlags) regexPatternValidation {
	if flags.Unicode && flags.UnicodeSets {
		return regexPatternValidation{}
	}
	if !flags.UV() && !hasRegexNamedCapture(pattern, flags) {
		_, err := esregexp.Compile(pattern, "")
		return regexPatternValidation{valid: err == nil}
	}
	// JavaScript strings are sequences of UTF-16 code units. The compiler keeps
	// lone surrogates in WTF-8, so join an adjacent pair before identifier-name
	// validation just as the RegExp parser's CodePoint operation does.
	pattern = ecmascript.CombineSurrogatePairs(pattern)

	normalized, captureNames, _, ok := normalizeRegexCaptureNames(pattern, flags, false)
	if !ok {
		return regexPatternValidation{
			hasDuplicateCaptureName:             captureNames.hasDuplicate,
			hasNonExclusiveDuplicateCaptureName: captureNames.hasNonExclusiveDuplicate,
		}
	}
	valid := false
	if !flags.UV() {
		_, err := esregexp.Compile(normalized, "")
		valid = err == nil
	} else {
		literal, literalOK := regexPatternLiteral(normalized, flags)
		valid = literalOK && ecmascript.IsValidRegexLiteral(literal)
	}
	return regexPatternValidation{
		valid:                               valid,
		hasDuplicateCaptureName:             captureNames.hasDuplicate,
		hasNonExclusiveDuplicateCaptureName: captureNames.hasNonExclusiveDuplicate,
	}
}

func regexFeaturesAvailable(pattern string, flags RegexFlags, ecmaVersion int) bool {
	if flags.Unicode && ecmaVersion < 2015 {
		return false
	}
	if flags.UnicodeSets && ecmaVersion < 2024 {
		return false
	}
	for i := 0; i < len(pattern); {
		switch pattern[i] {
		case '\\':
			if flags.UV() && i+1 < len(pattern) && ecmaVersion < 2018 &&
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
			if flags.UV() && ecmaVersion < 2018 && regexClassUsesUnicodeProperty(pattern, i, end, flags) {
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

func regexClassUsesUnicodeProperty(pattern string, start int, end int, flags RegexFlags) bool {
	for i := start + 1; i < end-1; {
		if pattern[i] == '\\' {
			if i+1 < end && (pattern[i+1] == 'p' || pattern[i+1] == 'P') {
				return true
			}
			step, ok := SkipPatternEscape(pattern, i, flags)
			if !ok {
				return false
			}
			i += step
			continue
		}
		_, width := utf8.DecodeRuneInString(pattern[i:])
		if width == 0 {
			width = 1
		}
		i += width
	}
	return false
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

// hasRegexNamedCapture is a cheap legacy-mode guard. Patterns without a named
// declaration stay on the existing regexp runtime, preserving Annex B escape
// handling without paying for capture-name normalization.
func hasRegexNamedCapture(pattern string, flags RegexFlags) bool {
	for i := 0; i < len(pattern); {
		switch pattern[i] {
		case '[':
			end, ok := ClassEnd(pattern, i, flags)
			if !ok {
				return false
			}
			i = end
		case '\\':
			step, ok := SkipPatternEscape(pattern, i, flags)
			if !ok {
				return false
			}
			i += step
		case '(':
			if strings.HasPrefix(pattern[i:], "(?<") &&
				(i+3 >= len(pattern) || pattern[i+3] != '=' && pattern[i+3] != '!') {
				return true
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
	return false
}

type regexCaptureNameAnalysis struct {
	hasDuplicate             bool
	hasNonExclusiveDuplicate bool
	nonExclusivePosition     int
}

type regexBranch struct {
	parent int
	base   int
	depth  int
}

// regexBranchTracker mirrors regexpp's ES2025 BranchID model. Every node is
// one Alternative; siblings share a base. Two declarations are mutually
// exclusive exactly when their ancestor paths choose different siblings of
// the same Disjunction.
type regexBranchTracker struct {
	branches []regexBranch
	current  int
}

func newRegexBranchTracker(pattern string) regexBranchTracker {
	const (
		minimumCapacity = 8
		maximumCapacity = 256
	)
	capacity := minimumCapacity
	if len(pattern) >= 64 {
		estimate := 2
		for i := 0; i < len(pattern) && estimate < maximumCapacity; i++ {
			if pattern[i] == '(' || pattern[i] == '|' {
				estimate++
			}
		}
		if estimate > capacity {
			capacity = estimate
		}
	}
	branches := make([]regexBranch, 1, capacity)
	branches[0] = regexBranch{parent: -1, base: 0}
	tracker := regexBranchTracker{branches: branches}
	tracker.enterDisjunction()
	return tracker
}

func (t *regexBranchTracker) enterDisjunction() {
	id := len(t.branches)
	depth := t.branches[t.current].depth + 1
	t.branches = append(t.branches, regexBranch{parent: t.current, base: id, depth: depth})
	t.current = id
}

func (t *regexBranchTracker) enterAlternative() {
	branch := t.branches[t.current]
	id := len(t.branches)
	t.branches = append(t.branches, branch)
	t.current = id
}

func (t *regexBranchTracker) leaveDisjunction() {
	if t.branches[t.current].depth > 1 {
		t.current = t.branches[t.current].parent
	}
}

func (t *regexBranchTracker) separatedFrom(left int, right int) bool {
	for t.branches[left].depth > t.branches[right].depth {
		left = t.branches[left].parent
	}
	for t.branches[right].depth > t.branches[left].depth {
		right = t.branches[right].parent
	}
	for left >= 0 && right >= 0 {
		leftBranch := t.branches[left]
		rightBranch := t.branches[right]
		if left != right && leftBranch.base == rightBranch.base {
			return true
		}
		left = leftBranch.parent
		right = rightBranch.parent
	}
	return false
}

// normalizeRegexCaptureNames rewrites every declaration and reference name to
// a deterministic ASCII placeholder keyed by its decoded IdentifierName. This
// preserves equality between raw and escaped spellings while allowing the
// grammar authority to validate the surrounding pattern. Duplicate
// declarations get distinct placeholders because ES2025 permits them when
// their Alternatives are mutually exclusive; references use the first
// declaration's placeholder.
func normalizeRegexCaptureNames(pattern string, flags RegexFlags, trackOffsets bool) (string, regexCaptureNameAnalysis, []int, bool) {
	var replacements []regexCaptureNameReplacement
	var placeholders map[string]string
	var lastDeclarationBranch map[string]int
	analysis := regexCaptureNameAnalysis{}
	placeholderCount := 0
	// A duplicate declaration requires at least two declaration-looking
	// prefixes. Avoid allocating the persistent branch tree for the common
	// zero/one-capture case; false positives such as lookbehinds only cause the
	// conservative tracking path to run.
	trackBranches := strings.Count(pattern, "(?<") > 1
	var branches regexBranchTracker
	if trackBranches {
		branches = newRegexBranchTracker(pattern)
	}

	newPlaceholder := func() string {
		placeholder := "rslint" + strconv.Itoa(placeholderCount)
		placeholderCount++
		return placeholder
	}

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
			placeholder = newPlaceholder()
			placeholders[name] = placeholder
		}
		if declaration {
			if lastDeclarationBranch == nil {
				lastDeclarationBranch = make(map[string]int)
			}
			previous, exists := lastDeclarationBranch[name]
			if exists {
				analysis.hasDuplicate = true
				placeholder = newPlaceholder()
				// Declarations arrive in source/DFS order. For three ordered
				// declarations p < q < r, separation of p/q and q/r implies
				// separation of p/r: well-nested Disjunction intervals cannot
				// return to p's earlier Alternative. Checking adjacent declarations
				// is therefore equivalent to regexpp's all-pairs test.
				if !analysis.hasNonExclusiveDuplicate &&
					(!trackBranches || !branches.separatedFrom(previous, branches.current)) {
					analysis.hasNonExclusiveDuplicate = true
					analysis.nonExclusivePosition = nameStart
				}
			}
			lastDeclarationBranch[name] = branches.current
		}
		replacements = append(replacements, regexCaptureNameReplacement{
			start:       nameStart,
			end:         nameEnd,
			placeholder: placeholder,
		})
		return true
	}

scanPattern:
	for i := 0; i < len(pattern); {
		switch pattern[i] {
		case '[':
			end := 0
			if flags.UnicodeSets {
				var ok bool
				end, _, ok = analyzeRegexVClass(pattern, i)
				if !ok {
					if trackOffsets {
						break scanPattern
					}
					return "", analysis, nil, false
				}
			} else {
				var ok bool
				end, ok = ClassEnd(pattern, i, flags)
				if !ok {
					if trackOffsets {
						break scanPattern
					}
					return "", analysis, nil, false
				}
			}
			i = end
		case '\\':
			if strings.HasPrefix(pattern[i:], `\k<`) {
				_, next, ok := readAngleName(pattern, i+2)
				if !ok || !addName(i+3, next-1, false) {
					if trackOffsets {
						break scanPattern
					}
					return "", analysis, nil, false
				}
				i = next
				continue
			}
			step, ok := SkipPatternEscape(pattern, i, flags)
			if !ok {
				if trackOffsets {
					break scanPattern
				}
				return "", analysis, nil, false
			}
			i += step
		case '(':
			if strings.HasPrefix(pattern[i:], "(?<") &&
				(i+3 >= len(pattern) || pattern[i+3] != '=' && pattern[i+3] != '!') {
				_, next, ok := readAngleName(pattern, i+2)
				if !ok || !addName(i+3, next-1, true) {
					if trackOffsets {
						break scanPattern
					}
					return "", analysis, nil, false
				}
				i = next
			} else {
				i++
			}
			if trackBranches {
				branches.enterDisjunction()
			}
		case '|':
			if trackBranches {
				branches.enterAlternative()
			}
			i++
		case ')':
			if trackBranches {
				branches.leaveDisjunction()
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
		return pattern, analysis, nil, true
	}
	var result strings.Builder
	result.Grow(len(pattern))
	var offsets []int
	if trackOffsets {
		offsets = make([]int, 0, len(pattern)+1)
	}
	writeOriginal := func(start int, end int) {
		result.WriteString(pattern[start:end])
		if trackOffsets {
			for i := start; i < end; i++ {
				offsets = append(offsets, i)
			}
		}
	}
	writeReplacement := func(replacement regexCaptureNameReplacement) {
		result.WriteString(replacement.placeholder)
		if trackOffsets {
			for range len(replacement.placeholder) {
				offsets = append(offsets, replacement.start)
			}
		}
	}
	last := 0
	for _, replacement := range replacements {
		writeOriginal(last, replacement.start)
		writeReplacement(replacement)
		last = replacement.end
	}
	writeOriginal(last, len(pattern))
	if trackOffsets {
		offsets = append(offsets, len(pattern))
	}
	return result.String(), analysis, offsets, true
}

type regexVClassFrame struct {
	operator       regexVClassOperator
	operandCount   int
	mayHaveStrings bool
	negated        bool
	trailingHyphen bool
}

type regexVClassOperator uint8

const (
	regexVClassUnion regexVClassOperator = iota
	regexVClassIntersection
	regexVClassSubtraction
)

func (f *regexVClassFrame) addOperand(mayHaveStrings bool) {
	if f.operandCount == 0 {
		f.mayHaveStrings = mayHaveStrings
	} else {
		switch f.operator {
		case regexVClassUnion:
			// Preserve later string operands even when the union starts with a
			// range. regexpp 4.12.2 drops that bit and can offer a v suggestion
			// which JavaScript then rejects.
			f.mayHaveStrings = f.mayHaveStrings || mayHaveStrings
		case regexVClassIntersection:
			f.mayHaveStrings = f.mayHaveStrings && mayHaveStrings
		case regexVClassSubtraction:
			// ClassSubtraction inherits only its left operand's value.
		}
	}
	f.operandCount++
	f.trailingHyphen = false
}

// analyzeRegexVClass performs the narrow v-mode checks needed before carrying
// a constructor pattern through a literal. One iterative pass covers nested
// classes, raw slashes, tsgo's trailing-hyphen and doubled-^/$ gaps. It also
// mirrors the ClassSetExpression MayContainStrings reduction because the
// pinned tsgo scanner does not propagate later union operands. tsgo remains
// the grammar authority for every other production.
func analyzeRegexVClass(pattern string, start int) (end int, errorPosition int, ok bool) {
	if start >= len(pattern) || pattern[start] != '[' {
		return start, start, false
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
				next, mayHaveStrings, qErrorPosition, qOK := scanRegexVQString(pattern, i)
				if !qOK {
					return start, qErrorPosition, false
				}
				frame.addOperand(mayHaveStrings)
				i = next
				continue
			}
			mayHaveStrings := regexVPropertyMayContainStrings(pattern, i)
			step, ok := SkipPatternEscape(pattern, i, flags)
			if !ok {
				return start, i, false
			}
			frame.addOperand(mayHaveStrings)
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
				return start, i, false
			}
			if frame.negated && frame.mayHaveStrings {
				return start, i, false
			}
			completedMayHaveStrings := !frame.negated && frame.mayHaveStrings
			frames = frames[:len(frames)-1]
			i++
			if len(frames) == 0 {
				return i, 0, true
			}
			parent := &frames[len(frames)-1]
			parent.addOperand(completedMayHaveStrings)
		case '/':
			// `/` is a ClassSetSyntaxCharacter. Escaping it only in the
			// carrier would make an invalid constructor pattern valid.
			return start, i, false
		case '^', '$':
			if i+1 < len(pattern) && pattern[i+1] == pattern[i] {
				return start, i, false
			}
			frame.addOperand(false)
			i++
		case '&':
			if i+1 < len(pattern) && pattern[i+1] == '&' {
				frame.operator = regexVClassIntersection
				frame.trailingHyphen = false
				i += 2
				continue
			}
			frame.addOperand(false)
			i++
		case '-':
			if i+1 < len(pattern) && pattern[i+1] == '-' {
				frame.operator = regexVClassSubtraction
				frame.trailingHyphen = false
				i += 2
				continue
			}
			frame.trailingHyphen = true
			i++
		default:
			frame.addOperand(false)
			_, width := utf8.DecodeRuneInString(pattern[i:])
			if width == 0 {
				width = 1
			}
			i += width
		}
	}
	return start, len(pattern), false
}

func scanRegexVQString(pattern string, start int) (next int, mayHaveStrings bool, errorPosition int, ok bool) {
	flags := RegexFlags{UnicodeSets: true}
	characterCount := 0
	for i := start + 3; i < len(pattern); {
		switch pattern[i] {
		case '\\':
			step, ok := SkipPatternEscape(pattern, i, flags)
			if !ok {
				return 0, false, i, false
			}
			characterCount++
			if high, fixed := regexFixedUnicodeEscape(pattern, i); fixed && high >= 0xD800 && high <= 0xDBFF {
				if low, lowFixed := regexFixedUnicodeEscape(pattern, i+step); lowFixed && low >= 0xDC00 && low <= 0xDFFF {
					step += 6
				}
			}
			i += step
		case '|':
			mayHaveStrings = mayHaveStrings || characterCount != 1
			characterCount = 0
			i++
		case '}':
			return i + 1, mayHaveStrings || characterCount != 1, 0, true
		case '/':
			return 0, false, i, false
		case '^', '$':
			if i+1 < len(pattern) && pattern[i+1] == pattern[i] {
				return 0, false, i, false
			}
			characterCount++
			i++
		default:
			characterCount++
			_, width := ecmascript.DecodeStringRune(pattern[i:])
			if width == 0 {
				width = 1
			}
			i += width
		}
	}
	return 0, false, len(pattern), false
}

func regexFixedUnicodeEscape(pattern string, start int) (uint32, bool) {
	if start < 0 || start+6 > len(pattern) || pattern[start] != '\\' || pattern[start+1] != 'u' ||
		!allHexStr(pattern[start+2:start+6]) {
		return 0, false
	}
	value, err := strconv.ParseUint(pattern[start+2:start+6], 16, 16)
	return uint32(value), err == nil
}

func regexVPropertyMayContainStrings(pattern string, start int) bool {
	if !strings.HasPrefix(pattern[start:], `\p{`) {
		return false
	}
	closeRel := strings.IndexByte(pattern[start+3:], '}')
	if closeRel < 0 {
		return false
	}
	switch pattern[start+3 : start+3+closeRel] {
	case "Basic_Emoji",
		"Emoji_Keycap_Sequence",
		"RGI_Emoji",
		"RGI_Emoji_Flag_Sequence",
		"RGI_Emoji_Modifier_Sequence",
		"RGI_Emoji_Tag_Sequence",
		"RGI_Emoji_ZWJ_Sequence":
		return true
	default:
		return false
	}
}

// regexPatternLiteral transports a u/v constructor pattern into a literal for
// tsgo's parser. It reads JavaScript UTF-16 code units, escapes only lexical
// delimiters, and rejects cases where encoding a line terminator or surrogate
// would turn an invalid identity escape into valid syntax.
func regexPatternLiteral(pattern string, flags RegexFlags) (string, bool) {
	literal, _, _, ok := buildRegexPatternLiteral(pattern, flags, false)
	return literal, ok
}

// RegexPatternCharacterEventCutoff returns the byte position where a syntax
// error stops regexpp-style character events. It transports constructor text
// through a RegExp literal while retaining an output-to-pattern position map.
func RegexPatternCharacterEventCutoff(pattern string, flags RegexFlags) (int, bool) {
	referencePosition, invalidReference := firstInvalidUnicodeNumericBackreference(pattern, flags)
	vClassPosition, invalidVClass := firstRegexVClassErrorPosition(pattern, flags)
	duplicatePosition, invalidDuplicate := firstDuplicateCaptureNamePosition(pattern, flags)
	position, invalid := regexPatternParserErrorPosition(pattern, flags, true)
	if invalidReference && (!invalid || referencePosition < position) {
		position, invalid = referencePosition, true
	}
	if invalidVClass && (!invalid || vClassPosition < position) {
		position, invalid = vClassPosition, true
	}
	if invalidDuplicate && (!invalid || duplicatePosition < position) {
		position, invalid = duplicatePosition, true
	}
	return position, invalid
}

func firstDuplicateCaptureNamePosition(pattern string, flags RegexFlags) (int, bool) {
	if strings.Count(pattern, "(?<") < 2 {
		return 0, false
	}
	_, analysis, _, _ := normalizeRegexCaptureNames(pattern, flags, false)
	return analysis.nonExclusivePosition, analysis.hasNonExclusiveDuplicate
}

func regexPatternParserErrorPosition(pattern string, flags RegexFlags, repairUnterminated bool) (int, bool) {
	if flags.Unicode && flags.UnicodeSets {
		// regexpp rejects conflicting modes before entering the pattern.
		return 0, true
	}
	if strings.HasPrefix(pattern, "*") {
		// Transporting this pattern as /*.../ would make the scanner read a
		// block comment. A quantifier cannot begin an ECMAScript pattern.
		return 0, true
	}
	literal, offsets, failureOffset, ok := buildRegexPatternLiteral(pattern, flags, true)
	if !ok {
		if repairUnterminated {
			if repairedPosition, found := regexPatternEarlierErrorAfterRepair(pattern, flags, failureOffset); found {
				return repairedPosition, true
			}
		}
		return failureOffset, true
	}
	position, unterminated, hasError := ecmascript.RegexLiteralCharacterEventCutoff(literal)
	if !hasError {
		return 0, false
	}
	if unterminated && repairUnterminated {
		if repairedPosition, found := regexPatternEarlierErrorAfterRepair(pattern, flags, len(pattern)); found {
			return repairedPosition, true
		}
		return len(pattern), true
	}
	if unterminated || position >= len(offsets) {
		return offsets[len(offsets)-1], true
	}
	return offsets[position], true
}

func regexPatternEarlierErrorAfterRepair(pattern string, flags RegexFlags, boundary int) (int, bool) {
	repaired := pattern
	trailingBackslashes := 0
	for i := len(repaired) - 1; i >= 0 && repaired[i] == '\\'; i-- {
		trailingBackslashes++
	}
	if trailingBackslashes%2 != 0 {
		repaired += `\`
	}
	repaired += strings.Repeat("}]", len(pattern)+1)
	position, invalid := regexPatternParserErrorPosition(repaired, flags, false)
	return position, invalid && position < boundary
}

func firstInvalidUnicodeNumericBackreference(pattern string, flags RegexFlags) (int, bool) {
	if !flags.UV() {
		return 0, false
	}
	groupCount, _ := esregexp.CapturingGroupCount(pattern)
	classDepth := 0
	for i := 0; i < len(pattern); {
		if pattern[i] == '\\' {
			if classDepth == 0 && i+1 < len(pattern) && pattern[i+1] >= '1' && pattern[i+1] <= '9' {
				value := 0
				for j := i + 1; j < len(pattern) && pattern[j] >= '0' && pattern[j] <= '9'; j++ {
					digit := int(pattern[j] - '0')
					if value > (groupCount-digit)/10 {
						return i, true
					}
					value = value*10 + digit
				}
				if value > groupCount {
					return i, true
				}
			}
			step, ok := SkipPatternEscape(pattern, i, flags)
			if !ok {
				i++
			} else {
				i += step
			}
			continue
		}
		switch pattern[i] {
		case '[':
			if classDepth == 0 || flags.UnicodeSets {
				classDepth++
			}
		case ']':
			if classDepth > 0 {
				classDepth--
			}
		}
		_, width := utf8.DecodeRuneInString(pattern[i:])
		if width == 0 {
			width = 1
		}
		i += width
	}
	return 0, false
}

func firstRegexVClassErrorPosition(pattern string, flags RegexFlags) (int, bool) {
	if !flags.UnicodeSets {
		return 0, false
	}
	for i := 0; i < len(pattern); {
		switch pattern[i] {
		case '\\':
			step, ok := SkipPatternEscape(pattern, i, flags)
			if !ok {
				return i, true
			}
			i += step
		case '[':
			end, errorPosition, ok := analyzeRegexVClass(pattern, i)
			if !ok {
				return errorPosition, true
			}
			i = end
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

func buildRegexPatternLiteral(pattern string, flags RegexFlags, trackOffsets bool) (literalText string, offsets []int, failureOffset int, ok bool) {
	pattern = ecmascript.CombineSurrogatePairs(pattern)
	originalLength := len(pattern)
	var normalizedOffsets []int
	if trackOffsets {
		if normalized, _, nameOffsets, namesOK := normalizeRegexCaptureNames(pattern, flags, true); namesOK {
			pattern = normalized
			normalizedOffsets = nameOffsets
		}
		if !flags.UV() {
			normalized, annexOffsets := esregexp.NormalizeAnnexBEscapesForParser(pattern)
			pattern = normalized
			if annexOffsets != nil {
				if normalizedOffsets != nil {
					for i, offset := range annexOffsets {
						annexOffsets[i] = normalizedOffsets[offset]
					}
				}
				normalizedOffsets = annexOffsets
			}
		}
	}
	patternOffset := func(offset int) int {
		if normalizedOffsets != nil && offset >= 0 && offset < len(normalizedOffsets) {
			return normalizedOffsets[offset]
		}
		return offset
	}
	units := ecmascript.StringCodeUnits(pattern)
	if ecmascript.StringFromCodeUnits(units) != pattern {
		return "", nil, 0, false
	}

	var literal strings.Builder
	literal.Grow(len(pattern) + 4)
	if trackOffsets {
		offsets = make([]int, 0, len(pattern)+5)
	}
	writeString := func(value string, patternOffset int) {
		literal.WriteString(value)
		if trackOffsets {
			for range len(value) {
				offsets = append(offsets, patternOffset)
			}
		}
	}
	writeRune := func(value rune, patternOffset int) {
		before := literal.Len()
		literal.WriteRune(value)
		if trackOffsets {
			for range literal.Len() - before {
				offsets = append(offsets, patternOffset)
			}
		}
	}
	writeUnicodeEscape := func(unit uint16, patternOffset int) {
		before := literal.Len()
		writeRegexUnicodeEscape(&literal, unit)
		if trackOffsets {
			for range literal.Len() - before {
				offsets = append(offsets, patternOffset)
			}
		}
	}

	writeString("/", 0)
	if len(units) == 0 {
		writeString("(?:)", 0)
	}

	escaped := false
	escapeStart := -1
	unitIndex := 0
	for sourceStart := 0; sourceStart < len(pattern); {
		r, size := ecmascript.DecodeStringRune(pattern[sourceStart:])
		originalSourceStart := patternOffset(sourceStart)
		if escaped && trackOffsets &&
			(r == '\n' || r == '\r' || r == 0x2028 || r == 0x2029 || r >= 0xD800 && r <= 0xDFFF || r > 0xFFFF) {
			// A constructor pattern may contain characters that cannot follow a
			// backslash in a RegExp literal carrier. Keep an invalid one-atom
			// escape in its place so tsgo can still reveal any earlier grammar
			// error; the inserted byte maps back to the original escape.
			writeString("q", escapeStart)
			escaped = false
			escapeStart = -1
			unitIndex++
			if r > 0xFFFF {
				unitIndex++
			}
			sourceStart += size
			continue
		}
		unitCount := 1
		if r > 0xFFFF {
			unitCount = 2
		}
		for range unitCount {
			unit := units[unitIndex]
			unitIndex++
			if unit == '\\' {
				writeString("\\", originalSourceStart)
				if escaped {
					escaped = false
					escapeStart = -1
				} else {
					escaped = true
					escapeStart = originalSourceStart
				}
				continue
			}

			switch unit {
			case '/':
				if !escaped {
					writeString("\\", originalSourceStart)
				}
				writeString("/", originalSourceStart)
			case '\n':
				if escaped {
					return "", nil, escapeStart, false
				}
				writeString(`\n`, originalSourceStart)
			case '\r':
				if escaped {
					return "", nil, escapeStart, false
				}
				writeString(`\r`, originalSourceStart)
			case 0x2028, 0x2029:
				if escaped {
					return "", nil, escapeStart, false
				}
				writeUnicodeEscape(unit, originalSourceStart)
			default:
				if unit >= 0xD800 && unit <= 0xDFFF {
					if escaped {
						return "", nil, escapeStart, false
					}
					writeUnicodeEscape(unit, originalSourceStart)
				} else {
					writeRune(rune(unit), originalSourceStart)
				}
			}
			escaped = false
			escapeStart = -1
		}
		sourceStart += size
	}
	if escaped {
		return "", nil, escapeStart, false
	}

	writeString("/", originalLength)
	if flags.UnicodeSets {
		writeString("v", originalLength)
	}
	if flags.Unicode {
		writeString("u", originalLength)
	}
	if trackOffsets {
		offsets = append(offsets, originalLength)
	}
	return literal.String(), offsets, 0, true
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

// NormalizeRegexCaptureName validates and decodes a RegExpIdentifierName.
// Callers that compare authored capture names must use the decoded value so
// raw and Unicode-escaped spellings name the same group.
func NormalizeRegexCaptureName(name string) (string, bool) {
	return normalizeRegexCaptureName(ecmascript.CombineSurrogatePairs(name))
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
