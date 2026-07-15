package minimatch

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
)

// ErrInvalidSearchPattern is returned when a search pattern cannot be parsed.
var ErrInvalidSearchPattern = errors.New("invalid search pattern")

// brace-expansion 5.x, used by Minimatch 10 in ESLint v10, truncates each
// expansion to its first 100,000 alternatives.
const maxBraceExpansions = 100_000
const maxSearchPatternUTF16Units = 65_536

// SearchPattern is an immutable, concurrency-safe matcher for one filesystem
// search pattern. Patterns and paths use normalized POSIX separators.
//
// Match implements an exact minimatch-style match. PartialMatch reports
// whether a directory is either a match itself or can contain a match, which
// is the behavior needed to prune a config-aware filesystem walk.
type SearchPattern struct {
	basePath        string
	rawPattern      string
	unmatchedKey    string
	matchRemovalKey string
	negated         bool
	comment         bool
	alternatives    []searchPathPattern
}

// CompileSearchPattern compiles pattern relative to basePath. Pattern may be
// absolute or relative. If basePath is empty, GlobParent(pattern) is used.
func CompileSearchPattern(pattern string, basePath string) (SearchPattern, error) {
	if err := validateSearchPatternLength(pattern); err != nil {
		return SearchPattern{}, err
	}
	result := SearchPattern{rawPattern: pattern}
	if pattern == "" {
		if basePath == "" {
			basePath = "."
		}
		result.basePath = cleanBasePath(basePath)
		result.alternatives = []searchPathPattern{{}}
		return result, nil
	}

	if basePath == "" {
		basePath = GlobParent(pattern)
	}
	result.basePath = cleanBasePath(basePath)
	patternWasAbsolute := isAbsoluteSearchPattern(pattern)
	unmatchedKey, err := patternRelativeToBase(pattern, result.basePath, patternWasAbsolute)
	if err != nil {
		return SearchPattern{}, err
	}
	return compilePatternBody(result, unmatchedKey, searchCompileOptions{})
}

// CompileRelativePattern compiles an already ConfigArray-normalized relative
// pattern without removing a leading "./". This is the raw Minimatch boundary
// used after ignore negation markers have been separated: a dot segment left
// behind by patterns such as "!!./src/**" must remain observable.
func CompileRelativePattern(pattern string) (SearchPattern, error) {
	return compileRelativePattern(pattern, searchCompileOptions{})
}

// CompileGitignorePattern compiles the converted Git-wildmatch dialect.
// Git patterns use ordinary * and ? wildcards but do not support Bash brace
// expansion or Minimatch extglob operators.
func CompileGitignorePattern(pattern string) (SearchPattern, error) {
	return compileRelativePattern(pattern, searchCompileOptions{
		noBrace:        true,
		noExt:          true,
		literalLeading: true,
		unitMode:       searchUnitModeBytes,
	})
}

type searchCompileOptions struct {
	noBrace        bool
	noExt          bool
	literalLeading bool
	unitMode       searchUnitMode
}

type searchUnitMode uint8

const (
	// The zero value selects Minimatch's per-segment regexp mode.
	searchUnitModeMinimatch searchUnitMode = iota
	searchUnitModeUTF16
	searchUnitModeUnicode
	searchUnitModeBytes
)

func searchStringUnits(value string, mode searchUnitMode) []rune {
	switch mode {
	case searchUnitModeUnicode:
		return []rune(value)
	case searchUnitModeUTF16:
		encoded := utf16.Encode([]rune(value))
		units := make([]rune, len(encoded))
		for index, unit := range encoded {
			units[index] = rune(unit)
		}
		return units
	case searchUnitModeBytes:
		encoded := []byte(value)
		units := make([]rune, len(encoded))
		for index, unit := range encoded {
			units[index] = rune(unit)
		}
		return units
	default:
		panic("search unit mode must be resolved before parsing or matching")
	}
}

func searchUnitsString(units []rune, mode searchUnitMode) string {
	switch mode {
	case searchUnitModeUnicode:
		return string(units)
	case searchUnitModeUTF16:
		encoded := make([]uint16, len(units))
		for index, unit := range units {
			encoded[index] = uint16(unit)
		}
		return string(utf16.Decode(encoded))
	case searchUnitModeBytes:
		encoded := make([]byte, len(units))
		for index, unit := range units {
			encoded[index] = byte(unit)
		}
		return string(encoded)
	default:
		panic("search unit mode must be resolved before decoding literals")
	}
}

func compileRelativePattern(pattern string, options searchCompileOptions) (SearchPattern, error) {
	if err := validateSearchPatternLength(pattern); err != nil {
		return SearchPattern{}, err
	}
	result := SearchPattern{
		basePath:   ".",
		rawPattern: pattern,
	}
	if pattern == "" {
		result.alternatives = []searchPathPattern{{}}
		return result, nil
	}
	return compilePatternBody(result, pattern, options)
}

func validateSearchPatternLength(pattern string) error {
	units := 0
	for _, character := range pattern {
		units++
		if character > 0xffff {
			units++
		}
		if units > maxSearchPatternUTF16Units {
			return fmt.Errorf("%w: pattern is too long", ErrInvalidSearchPattern)
		}
	}
	return nil
}

func compilePatternBody(result SearchPattern, unmatchedKey string, options searchCompileOptions) (SearchPattern, error) {
	result.unmatchedKey = unmatchedKey
	result.matchRemovalKey = strings.TrimLeft(unmatchedKey, "!")
	// Minimatch interprets comment and leading-negation syntax on the original
	// pattern before brace expansion. A # or ! introduced by a brace alternative
	// is therefore a literal (for example {#,x}*.js and {!,x}foo.js).
	matchPattern := unmatchedKey
	if !options.literalLeading {
		if strings.HasPrefix(matchPattern, "#") {
			result.comment = true
			return result, nil
		}
		for strings.HasPrefix(matchPattern, "!") {
			result.negated = !result.negated
			matchPattern = strings.TrimPrefix(matchPattern, "!")
		}
	}

	expanded := []string{matchPattern}
	if !options.noBrace {
		var err error
		expanded, err = expandSearchBraces(matchPattern)
		if err != nil {
			return SearchPattern{}, err
		}
	}
	result.alternatives = make([]searchPathPattern, 0, len(expanded))
	for _, expandedPattern := range expanded {
		compiled, compileErr := compileSearchPathPattern(expandedPattern, options)
		if compileErr != nil {
			return SearchPattern{}, compileErr
		}
		result.alternatives = append(result.alternatives, compiled)
	}
	return result, nil
}

// UnmatchedKey is the base-relative spelling ESLint inserts into its
// unmatched-pattern set. MatchRemovalKey mirrors Minimatch.pattern: leading
// negation markers are stripped before a successful match tries to delete
// from that set. Keeping both values intentionally preserves ESLint v10's
// observable leading-! unmatched-pattern behavior.
func (pattern SearchPattern) UnmatchedKey() string {
	return pattern.unmatchedKey
}

func (pattern SearchPattern) MatchRemovalKey() string {
	return pattern.matchRemovalKey
}

// BasePath returns the lexical search base used by the matcher.
func (pattern SearchPattern) BasePath() string {
	return pattern.basePath
}

// RawPattern returns the pattern passed to CompileSearchPattern.
func (pattern SearchPattern) RawPattern() string {
	return pattern.rawPattern
}

// LiteralPrefixes returns each brace-expanded alternative's leading concrete
// path prefix after Minimatch's default preprocessing. An empty prefix means
// the alternative starts with glob syntax and may reach any subtree.
func (pattern SearchPattern) LiteralPrefixes() []string {
	prefixes := make([]string, 0, len(pattern.alternatives))
	for _, alternative := range pattern.alternatives {
		segments := make([]string, 0, len(alternative.segments))
		for _, segment := range alternative.segments {
			if segment.globstar {
				break
			}
			literal, ok := segment.literalValue()
			if !ok {
				break
			}
			segments = append(segments, literal)
		}
		prefixes = append(prefixes, strings.Join(segments, "/"))
	}
	return prefixes
}

// Match reports whether relativePath exactly matches the pattern.
func (pattern SearchPattern) Match(relativePath string) bool {
	if pattern.comment {
		return false
	}
	relativePath = normalizeRelativeSearchPath(relativePath)
	matched := false
	for _, alternative := range pattern.alternatives {
		if alternative.match(relativePath) {
			matched = true
			break
		}
	}
	if pattern.negated {
		return !matched
	}
	return matched
}

// PartialMatch reports whether relativeDir is a viable prefix of an exact
// match. A directory that exactly matches also returns true, matching
// minimatch's partial option.
func (pattern SearchPattern) PartialMatch(relativeDir string) bool {
	if pattern.comment {
		return false
	}
	relativeDir = normalizeRelativeSearchPath(relativeDir)
	matched := false
	for _, alternative := range pattern.alternatives {
		if alternative.partialMatch(relativeDir) {
			matched = true
			break
		}
	}
	if pattern.negated {
		return !matched
	}
	return matched
}

// GlobParent returns the non-glob parent of pattern using glob-parent 6.x
// semantics. In particular, a bare '?' is not considered glob magic by the
// strict is-glob check used by ESLint.
func GlobParent(pattern string) string {
	if searchPatternHasSlashEnclosure(pattern) {
		pattern += "/"
	}

	candidate := pattern + "a"
	for {
		parent := path.Dir(candidate)
		candidate = parent
		if !isGlobbySearchParent(candidate) {
			break
		}
	}
	return unescapeGlobParent(candidate)
}

// isGlobPattern mirrors the strict is-glob classification ESLint uses before
// grouping a CLI pattern by GlobParent.
func IsGlobPattern(pattern string) bool {
	return isGlobbySearchPath(pattern)
}

func isGlobbySearchParent(value string) bool {
	if value == "" {
		return false
	}
	if value[0] == '{' || value[0] == '[' {
		return true
	}
	for index := 0; index < len(value); index++ {
		if value[index] == '\\' {
			index++
			continue
		}
		if value[index] == '{' || value[index] == '[' {
			return true
		}
		if value[index] == '(' && findMatchingParen(value, index) < 0 && index+1 < len(value) {
			return true
		}
	}
	return isGlobbySearchPath(value)
}

func cleanBasePath(basePath string) string {
	if basePath == "" {
		return "."
	}
	return path.Clean(basePath)
}

func patternRelativeToBase(pattern string, basePath string, patternIsAbsolute bool) (string, error) {
	if !patternIsAbsolute {
		return strings.TrimPrefix(pattern, "./"), nil
	}
	if basePath == "/" {
		return strings.TrimPrefix(pattern, "/"), nil
	}
	if relative, ok := stripAbsoluteSearchBase(pattern, basePath); ok {
		return relative, nil
	}
	// ESLint always feeds Node path.relative(basePath, pattern) into
	// Minimatch, even when slash normalization has moved an otherwise existing
	// POSIX directory pattern outside its native-spelled base. Such a pattern
	// simply cannot match entries produced by that walk; it is not a config or
	// parser error.
	relative, err := filepath.Rel(filepath.FromSlash(basePath), filepath.FromSlash(pattern))
	if err == nil {
		return filepath.ToSlash(relative), nil
	}
	// Different Windows volumes are the one case where filepath.Rel can fail.
	// Node returns the absolute target there, which likewise cannot match a
	// walk-relative entry.
	return pattern, nil
}

func isAbsoluteSearchPattern(pattern string) bool {
	if strings.HasPrefix(pattern, "/") {
		return true
	}
	return len(pattern) >= 3 && ((pattern[0] >= 'A' && pattern[0] <= 'Z') || (pattern[0] >= 'a' && pattern[0] <= 'z')) && pattern[1] == ':' && pattern[2] == '/'
}

func stripAbsoluteSearchBase(pattern string, basePath string) (string, bool) {
	var decoded strings.Builder
	for index := 0; index < len(pattern); index++ {
		character := pattern[index]
		if character == '\\' && index+1 < len(pattern) && strings.ContainsRune("!*?|[](){}", rune(pattern[index+1])) {
			index++
			character = pattern[index]
		}
		decoded.WriteByte(character)
		decodedPrefix := decoded.String()
		if !strings.HasPrefix(basePath, decodedPrefix) {
			return "", false
		}
		if decodedPrefix != basePath {
			continue
		}
		if index+1 == len(pattern) {
			return "", true
		}
		if pattern[index+1] == '/' {
			return pattern[index+2:], true
		}
		return "", false
	}
	if decoded.String() != basePath {
		return "", false
	}
	return "", true
}

func normalizeRelativeSearchPath(relativePath string) string {
	for strings.HasPrefix(relativePath, "./") {
		relativePath = strings.TrimPrefix(relativePath, "./")
	}
	return strings.TrimPrefix(relativePath, "/")
}

func searchPatternHasSlashEnclosure(pattern string) bool {
	if len(pattern) < 2 {
		return false
	}
	var opening byte
	switch pattern[len(pattern)-1] {
	case '}':
		opening = '{'
	case ']':
		opening = '['
	default:
		return false
	}
	openingIndex := strings.IndexByte(pattern, opening)
	return openingIndex >= 0 && strings.Contains(pattern[openingIndex+1:len(pattern)-1], "/")
}

func isGlobbySearchPath(value string) bool {
	if value == "" {
		return false
	}
	if value[0] == '!' {
		return true
	}
	for index := 0; index < len(value); index++ {
		if value[index] == '\\' {
			index++
			continue
		}
		if index+1 < len(value) && value[index+1] == '?' && strings.ContainsRune("].+)", rune(value[index])) {
			return true
		}
		switch value[index] {
		case '*':
			return true
		case '[', '{':
			closing := byte(']')
			if value[index] == '{' {
				closing = '}'
			}
			if findUnescapedByte(value, closing, index+1) > index+1 {
				return true
			}
		case '@', '?', '+', '!':
			if index+1 < len(value) && value[index+1] == '(' && findMatchingParen(value, index+1) >= 0 {
				return true
			}
		case '(':
			closing := findMatchingParen(value, index)
			if closing > index+1 {
				content := value[index+1 : closing]
				if (strings.HasPrefix(content, "?:") || strings.HasPrefix(content, "?!") || strings.HasPrefix(content, "?=") || strings.HasPrefix(content, "?<")) && len(content) > 2 {
					return true
				}
				if hasUnescapedPipe(content) && content[0] != '|' && content[len(content)-1] != '|' {
					return true
				}
			}
		}
	}
	return false
}

func hasUnescapedPipe(value string) bool {
	for index := 0; index < len(value); index++ {
		if value[index] == '\\' {
			index++
			continue
		}
		if value[index] == '|' {
			return true
		}
	}
	return false
}

func findUnescapedByte(value string, target byte, start int) int {
	for index := start; index < len(value); index++ {
		if value[index] == '\\' {
			index++
			continue
		}
		if value[index] == target {
			return index
		}
	}
	return -1
}

func findMatchingParen(value string, opening int) int {
	depth := 0
	for index := opening; index < len(value); index++ {
		if value[index] == '\\' {
			index++
			continue
		}
		switch value[index] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return index
			}
		}
	}
	return -1
}

func unescapeGlobParent(value string) string {
	var result strings.Builder
	for index := 0; index < len(value); index++ {
		if value[index] == '\\' && index+1 < len(value) && strings.ContainsRune("!*?|[](){}", rune(value[index+1])) {
			index++
		}
		result.WriteByte(value[index])
	}
	return result.String()
}

type searchPathPattern struct {
	segments []searchSegmentPattern
}

func compileSearchPathPattern(pattern string, options searchCompileOptions) (searchPathPattern, error) {
	if pattern == "" {
		return searchPathPattern{}, nil
	}
	parts := levelOneOptimizeSearchParts(splitSearchPath(pattern))
	compiled := searchPathPattern{segments: make([]searchSegmentPattern, 0, len(parts))}
	for _, part := range parts {
		if part == "**" {
			compiled.segments = append(compiled.segments, searchSegmentPattern{globstar: true})
			continue
		}
		segment, err := compileSearchSegment(part, options)
		if err != nil {
			return searchPathPattern{}, fmt.Errorf("%w: %q: %w", ErrInvalidSearchPattern, pattern, err)
		}
		compiled.segments = append(compiled.segments, segment)
	}
	return compiled, nil
}

// levelOneOptimizeSearchParts mirrors Minimatch's default optimization level:
// adjacent globstars collapse and a literal segment immediately followed by
// ".." cancels. A "." segment is deliberately retained.
func levelOneOptimizeSearchParts(parts []string) []string {
	optimized := make([]string, 0, len(parts))
	for _, part := range parts {
		if len(optimized) > 0 {
			previous := optimized[len(optimized)-1]
			if part == "**" && previous == "**" {
				continue
			}
			if part == ".." && previous != "" && previous != ".." &&
				previous != "." && previous != "**" {
				optimized = optimized[:len(optimized)-1]
				continue
			}
		}
		optimized = append(optimized, part)
	}
	if len(optimized) == 0 {
		return []string{""}
	}
	return optimized
}

func (pattern searchPathPattern) match(relativePath string) bool {
	parts := splitSearchPath(relativePath)
	if relativePath == "" && len(pattern.segments) == 0 {
		return true
	}
	firstGlobstar := -1
	lastGlobstar := -1
	for index, segment := range pattern.segments {
		if segment.globstar {
			if firstGlobstar == -1 {
				firstGlobstar = index
			}
			lastGlobstar = index
		}
	}
	if firstGlobstar == -1 {
		return len(parts) == len(pattern.segments) &&
			matchSearchSegmentSection(parts, pattern.segments, 0)
	}
	return pattern.matchGlobstar(parts, firstGlobstar, lastGlobstar)
}

const maxSearchGlobstarRecursion = 200

// matchGlobstar follows Minimatch 10's exact-match decomposition: the fixed
// head and tail are matched once, while each non-globstar body section is
// searched for between globstars. The 200-section ceiling is an intentional
// false-negative security bound in Minimatch.
func (pattern searchPathPattern) matchGlobstar(parts []string, firstGlobstar int, lastGlobstar int) bool {
	head := pattern.segments[:firstGlobstar]
	if !matchSearchSegmentSection(parts, head, 0) {
		return false
	}
	fileIndex := len(head)

	tail := pattern.segments[lastGlobstar+1:]
	fileTailMatch := 0
	if len(tail) != 0 {
		if fileIndex+len(tail) > len(parts) {
			return false
		}
		tailStart := len(parts) - len(tail)
		if matchSearchSegmentSection(parts, tail, tailStart) {
			fileTailMatch = len(tail)
		} else {
			// Minimatch lets a fixed tail match immediately before a trailing
			// empty path portion (for example a/**/b against a/b/).
			if len(parts) == 0 || parts[len(parts)-1] != "" || fileIndex+len(tail) == len(parts) {
				return false
			}
			tailStart--
			if tailStart < fileIndex || !matchSearchSegmentSection(parts, tail, tailStart) {
				return false
			}
			fileTailMatch = len(tail) + 1
		}
	}

	var body []searchSegmentPattern
	if firstGlobstar != lastGlobstar {
		body = pattern.segments[firstGlobstar+1 : lastGlobstar]
	}
	if len(body) == 0 {
		sawSome := fileTailMatch != 0
		for index := fileIndex; index < len(parts)-fileTailMatch; index++ {
			sawSome = true
			if isSearchGlobstarTraversalSegment(parts[index]) {
				return false
			}
		}
		return sawSome
	}

	sections := [][]searchSegmentPattern{{}}
	nonGlobstarParts := 0
	nonGlobstarPartSums := []int{0}
	for _, segment := range body {
		if segment.globstar {
			nonGlobstarPartSums = append(nonGlobstarPartSums, nonGlobstarParts)
			sections = append(sections, []searchSegmentPattern{})
			continue
		}
		sections[len(sections)-1] = append(sections[len(sections)-1], segment)
		nonGlobstarParts++
	}

	// Preserve Minimatch 10's actual section bound calculation, including its
	// reverse association between cumulative part counts and body sections.
	// Unequal-sized sections make this observably stricter than a conventional
	// suffix-length bound, so simplifying the formula changes match results.
	fileEnd := len(parts) - fileTailMatch
	after := make([]int, len(sections))
	sumIndex := len(sections) - 1
	for index, section := range sections {
		after[index] = fileEnd - (nonGlobstarPartSums[sumIndex] + len(section))
		sumIndex--
	}

	type bodyState struct{ sectionIndex, fileIndex int }
	memo := make(map[bodyState]bool)
	visited := make(map[bodyState]struct{})
	var matchBody func(int, int, int) bool
	matchBody = func(sectionIndex int, candidate int, globstarDepth int) bool {
		state := bodyState{sectionIndex: sectionIndex, fileIndex: candidate}
		if _, ok := visited[state]; ok {
			return memo[state]
		}
		visited[state] = struct{}{}

		if sectionIndex == len(sections) {
			sawTail := fileTailMatch != 0
			for index := candidate; index < len(parts); index++ {
				sawTail = true
				if isSearchGlobstarTraversalSegment(parts[index]) {
					memo[state] = false
					return false
				}
			}
			memo[state] = sawTail
			return sawTail
		}

		section := sections[sectionIndex]
		for candidate <= after[sectionIndex] {
			if matchSearchSegmentSection(parts, section, candidate) &&
				globstarDepth < maxSearchGlobstarRecursion &&
				matchBody(sectionIndex+1, candidate+len(section), globstarDepth+1) {
				memo[state] = true
				return true
			}
			if candidate < len(parts) && isSearchGlobstarTraversalSegment(parts[candidate]) {
				memo[state] = false
				return false
			}
			candidate++
		}
		memo[state] = false
		return false
	}
	return matchBody(0, fileIndex, 0)
}

func matchSearchSegmentSection(parts []string, section []searchSegmentPattern, fileIndex int) bool {
	if fileIndex < 0 || fileIndex+len(section) > len(parts) {
		return false
	}
	for offset, segment := range section {
		if segment.globstar || !segment.match(parts[fileIndex+offset]) {
			return false
		}
	}
	return true
}

func isSearchGlobstarTraversalSegment(value string) bool {
	return value == "." || value == ".."
}

func (pattern searchPathPattern) partialMatch(relativeDir string) bool {
	parts := splitSearchPathForPartial(relativeDir)
	states := map[int]struct{}{0: {}}
	states = pattern.globstarClosure(states)
	for _, part := range parts {
		next := make(map[int]struct{})
		for patternIndex := range states {
			if patternIndex == len(pattern.segments) {
				continue
			}
			segment := pattern.segments[patternIndex]
			if segment.globstar {
				if part != "." && part != ".." {
					next[patternIndex] = struct{}{}
				}
				continue
			}
			if segment.match(part) {
				next[patternIndex+1] = struct{}{}
			}
		}
		states = pattern.globstarClosure(next)
		if len(states) == 0 {
			return false
		}
	}
	return len(states) != 0
}

func (pattern searchPathPattern) globstarClosure(states map[int]struct{}) map[int]struct{} {
	for patternIndex := range states {
		for patternIndex < len(pattern.segments) && pattern.segments[patternIndex].globstar {
			patternIndex++
			states[patternIndex] = struct{}{}
		}
	}
	return states
}

func splitSearchPath(relativePath string) []string {
	return strings.Split(coalesceSearchSlashes(relativePath), "/")
}

func splitSearchPathForPartial(relativePath string) []string {
	if relativePath == "" {
		return []string{""}
	}
	return splitSearchPath(relativePath)
}

func coalesceSearchSlashes(value string) string {
	if !strings.Contains(value, "//") {
		return value
	}
	var result strings.Builder
	result.Grow(len(value))
	previousSlash := false
	for _, character := range value {
		if character == '/' {
			if previousSlash {
				continue
			}
			previousSlash = true
		} else {
			previousSlash = false
		}
		result.WriteRune(character)
	}
	return result.String()
}

type searchSegmentPattern struct {
	globstar bool
	sequence searchSegmentSequence
	raw      string
	unitMode searchUnitMode
}

func (pattern searchSegmentPattern) literalValue() (string, bool) {
	var literal strings.Builder
	for _, node := range pattern.sequence {
		if node.kind != searchSegmentLiteral {
			return "", false
		}
		literal.WriteString(searchUnitsString(node.literal, pattern.unitMode))
	}
	return literal.String(), true
}

func compileSearchSegment(pattern string, options searchCompileOptions) (searchSegmentPattern, error) {
	unitMode := options.unitMode
	var sequence searchSegmentSequence
	var err error
	if unitMode == searchUnitModeMinimatch {
		// Minimatch normally renders a JavaScript regexp without /u, so its
		// wildcards consume UTF-16 code units. A property-based POSIX class
		// makes the regexp for this one path segment Unicode-aware instead.
		sequence, err = parseSearchSegment(pattern, options.noExt, searchUnitModeUnicode)
		if err == nil && !sequence.requiresUnicode() {
			unitMode = searchUnitModeUTF16
			sequence, err = parseSearchSegment(pattern, options.noExt, unitMode)
		} else {
			unitMode = searchUnitModeUnicode
		}
	} else {
		sequence, err = parseSearchSegment(pattern, options.noExt, unitMode)
	}
	if err != nil {
		return searchSegmentPattern{}, err
	}
	sequence = flattenSearchSequence(sequence)
	sequence = literalizeStandaloneEmptyExtglobs(sequence, unitMode)
	sequence = annotateNegativeExtglobTails(sequence, nil)
	return searchSegmentPattern{sequence: sequence, raw: pattern, unitMode: unitMode}, nil
}

func parseSearchSegment(pattern string, noExt bool, unitMode searchUnitMode) (searchSegmentSequence, error) {
	parser := searchSegmentParser{input: searchStringUnits(pattern, unitMode), noExt: noExt, unitMode: unitMode}
	sequence, terminator, _, err := parser.parseSequence(false, 0, 0)
	if err != nil {
		return nil, err
	}
	if terminator != 0 || parser.index != len(parser.input) {
		return nil, errors.New("unexpected extglob terminator")
	}
	return sequence, nil
}

func flattenSearchSequence(sequence searchSegmentSequence) searchSegmentSequence {
	result := append(searchSegmentSequence(nil), sequence...)
	for index := range result {
		if result[index].kind == searchSegmentExtglob {
			result[index] = flattenSearchExtglob(result[index])
		}
	}
	return result
}

// Minimatch flattens nested extglobs before rendering a path segment, then
// renders a non-negative empty extglob as a literal only when it occupies the
// whole segment. For example, @(@()) flattens to the literal "@()", while the
// same empty tree embedded in x@(@())y still consumes no characters.
func literalizeStandaloneEmptyExtglobs(sequence searchSegmentSequence, unitMode searchUnitMode) searchSegmentSequence {
	if len(sequence) != 1 || sequence[0].kind != searchSegmentExtglob || sequence[0].extglobType == '!' {
		return sequence
	}
	node := sequence[0]
	if !extglobBodyMatchesOnlyEmpty(node.alternatives) {
		return sequence
	}
	literal := string(node.extglobType) + "(" + strings.Repeat("|", len(node.alternatives)-1) + ")"
	return searchSegmentSequence{{kind: searchSegmentLiteral, literal: searchStringUnits(literal, unitMode)}}
}

func flattenSearchExtglob(node searchSegmentNode) searchSegmentNode {
	for alternativeIndex := range node.alternatives {
		for childIndex := range node.alternatives[alternativeIndex] {
			child := &node.alternatives[alternativeIndex][childIndex]
			if child.kind == searchSegmentExtglob {
				*child = flattenSearchExtglob(*child)
			}
		}
	}
	for range 10 {
		changed := false
		for alternativeIndex := 0; alternativeIndex < len(node.alternatives); alternativeIndex++ {
			alternative := node.alternatives[alternativeIndex]
			if len(alternative) != 1 || alternative[0].kind != searchSegmentExtglob {
				continue
			}
			child := alternative[0]
			if searchExtglobCanAdopt(node.extglobType, child.extglobType) {
				node.alternatives = spliceSearchAlternatives(node.alternatives, alternativeIndex, child.alternatives)
				changed = true
				alternativeIndex--
				continue
			}
			if searchExtglobCanAdoptWithEmpty(node.extglobType, child.extglobType) {
				adopted := append(append([]searchSegmentSequence(nil), child.alternatives...), searchSegmentSequence{})
				node.alternatives = spliceSearchAlternatives(node.alternatives, alternativeIndex, adopted)
				changed = true
				alternativeIndex--
				continue
			}
			if len(node.alternatives) == 1 {
				if adoptedType, ok := searchExtglobUsurpedType(node.extglobType, child.extglobType); ok {
					node.extglobType = adoptedType
					node.alternatives = append([]searchSegmentSequence(nil), child.alternatives...)
					node.emptyExt = false
					changed = true
					break
				}
			}
		}
		if !changed {
			break
		}
	}
	return node
}

// Minimatch fills every negative extglob with the suffix that follows it
// through all ordinary sequence ancestors, then anchors the negative
// lookahead at the end of the path segment. Keeping that suffix as matcher
// metadata avoids requiring regexp lookahead while preserving the same
// endpoint semantics.
func annotateNegativeExtglobTails(sequence searchSegmentSequence, outerTail searchSegmentSequence) searchSegmentSequence {
	result := append(searchSegmentSequence(nil), sequence...)
	for index := len(result) - 1; index >= 0; index-- {
		tail := concatSearchSequences(result[index+1:], outerTail)
		node := result[index]
		if node.kind != searchSegmentExtglob {
			continue
		}
		for alternativeIndex := range node.alternatives {
			node.alternatives[alternativeIndex] = annotateNegativeExtglobTails(
				node.alternatives[alternativeIndex],
				tail,
			)
		}
		if node.extglobType == '!' && !node.emptyExt {
			node.lookaheadAlternatives = make([]searchSegmentSequence, len(node.alternatives))
			for alternativeIndex, alternative := range node.alternatives {
				node.lookaheadAlternatives[alternativeIndex] = concatSearchSequences(alternative, tail)
			}
		}
		result[index] = node
	}
	return result
}

func concatSearchSequences(left searchSegmentSequence, right searchSegmentSequence) searchSegmentSequence {
	result := make(searchSegmentSequence, 0, len(left)+len(right))
	result = append(result, left...)
	return append(result, right...)
}

func spliceSearchAlternatives(
	alternatives []searchSegmentSequence,
	index int,
	replacement []searchSegmentSequence,
) []searchSegmentSequence {
	result := make([]searchSegmentSequence, 0, len(alternatives)-1+len(replacement))
	result = append(result, alternatives[:index]...)
	result = append(result, replacement...)
	return append(result, alternatives[index+1:]...)
}

func searchExtglobCanAdopt(parent rune, child rune) bool {
	switch parent {
	case '!':
		return child == '@'
	case '?':
		return child == '?' || child == '@'
	case '@':
		return child == '@'
	case '*':
		return child == '*' || child == '+' || child == '?' || child == '@'
	case '+':
		return child == '+' || child == '@'
	default:
		return false
	}
}

func searchExtglobCanAdoptWithEmpty(parent rune, child rune) bool {
	switch parent {
	case '!':
		return child == '?'
	case '@':
		return child == '?'
	case '+':
		return child == '?' || child == '*'
	default:
		return false
	}
}

func searchExtglobCanAdoptAny(parent rune, child rune) bool {
	switch parent {
	case '!':
		return child == '?' || child == '@'
	case '?', '@':
		return child == '?' || child == '@'
	case '*':
		return child == '*' || child == '+' || child == '?' || child == '@'
	case '+':
		return child == '+' || child == '@' || child == '?' || child == '*'
	default:
		return false
	}
}

func searchExtglobUsurpedType(parent rune, child rune) (rune, bool) {
	if parent == '@' {
		return child, true
	}
	switch parent {
	case '!':
		if child == '!' {
			return '@', true
		}
	case '?':
		if child == '*' || child == '+' {
			return '*', true
		}
	case '+':
		if child == '?' || child == '*' {
			return '*', true
		}
	}
	return 0, false
}

func extglobBodyMatchesOnlyEmpty(alternatives []searchSegmentSequence) bool {
	if len(alternatives) == 0 {
		return false
	}
	for _, alternative := range alternatives {
		if len(alternative) != 0 {
			return false
		}
	}
	return true
}

func (pattern searchSegmentPattern) match(value string) bool {
	input := searchStringUnits(value, pattern.unitMode)
	ends := pattern.sequence.matchFrom(input, 0)
	if _, ok := ends[len(input)]; !ok {
		return false
	}
	if (value == "." || value == "..") && pattern.sequence.hasTraversalWildcard() {
		return false
	}
	if len(input) == 0 && pattern.raw != "" && !pattern.sequence.explicitlyAllowsEmpty() {
		return false
	}
	return true
}

func (sequence searchSegmentSequence) hasTraversalWildcard() bool {
	for _, node := range sequence {
		switch node.kind {
		case searchSegmentStar, searchSegmentQuestion:
			return true
		case searchSegmentClass:
			if node.class.negated {
				return true
			}
		case searchSegmentExtglob:
			if node.extglobType == '!' {
				continue
			}
			for _, alternative := range node.alternatives {
				if alternative.hasTraversalWildcard() {
					return true
				}
			}
		}
	}
	return false
}

type searchSegmentNodeKind uint8

const (
	searchSegmentLiteral searchSegmentNodeKind = iota
	searchSegmentStar
	searchSegmentQuestion
	searchSegmentClass
	searchSegmentExtglob
)

type searchSegmentNode struct {
	kind                  searchSegmentNodeKind
	literal               []rune
	class                 searchRuneClass
	extglobType           rune
	alternatives          []searchSegmentSequence
	emptyExt              bool
	lookaheadAlternatives []searchSegmentSequence
}

type searchSegmentSequence []searchSegmentNode

func (sequence searchSegmentSequence) requiresUnicode() bool {
	for _, node := range sequence {
		if node.kind == searchSegmentClass {
			for _, name := range node.class.posix {
				if searchPOSIXClassRequiresUnicode(name) {
					return true
				}
			}
		}
		if node.kind == searchSegmentExtglob {
			for _, alternative := range node.alternatives {
				if alternative.requiresUnicode() {
					return true
				}
			}
		}
	}
	return false
}

func (sequence searchSegmentSequence) matchFrom(input []rune, start int) map[int]struct{} {
	positions := map[int]struct{}{start: {}}
	for _, node := range sequence {
		next := make(map[int]struct{})
		for position := range positions {
			for end := range node.advance(input, position) {
				next[end] = struct{}{}
			}
		}
		positions = next
		if len(positions) == 0 {
			break
		}
	}
	return positions
}

func (sequence searchSegmentSequence) explicitlyAllowsEmpty() bool {
	if len(sequence) == 0 {
		return true
	}
	for _, node := range sequence {
		switch node.kind {
		case searchSegmentExtglob:
			if node.extglobType != '?' && node.extglobType != '*' && node.extglobType != '!' {
				return false
			}
		case searchSegmentLiteral:
			if len(node.literal) != 0 {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func (node searchSegmentNode) advance(input []rune, start int) map[int]struct{} {
	result := make(map[int]struct{})
	switch node.kind {
	case searchSegmentLiteral:
		if start+len(node.literal) <= len(input) {
			matches := true
			for offset, character := range node.literal {
				if input[start+offset] != character {
					matches = false
					break
				}
			}
			if matches {
				result[start+len(node.literal)] = struct{}{}
			}
		}
	case searchSegmentStar:
		for end := start; end <= len(input); end++ {
			result[end] = struct{}{}
		}
	case searchSegmentQuestion:
		if start < len(input) {
			result[start+1] = struct{}{}
		}
	case searchSegmentClass:
		if start < len(input) && node.class.matches(input[start]) {
			result[start+1] = struct{}{}
		}
	case searchSegmentExtglob:
		return node.advanceExtglob(input, start)
	}
	return result
}

func (node searchSegmentNode) advanceExtglob(input []rune, start int) map[int]struct{} {
	matchOnce := func(position int) map[int]struct{} {
		result := make(map[int]struct{})
		for _, alternative := range node.alternatives {
			for end := range alternative.matchFrom(input, position) {
				result[end] = struct{}{}
			}
		}
		return result
	}

	switch node.extglobType {
	case '@':
		return matchOnce(start)
	case '?':
		result := matchOnce(start)
		result[start] = struct{}{}
		return result
	case '+', '*':
		result := make(map[int]struct{})
		frontier := matchOnce(start)
		if node.extglobType == '*' {
			result[start] = struct{}{}
		}
		for len(frontier) != 0 {
			next := make(map[int]struct{})
			for position := range frontier {
				if _, seen := result[position]; seen {
					continue
				}
				result[position] = struct{}{}
				for end := range matchOnce(position) {
					if end != position {
						next[end] = struct{}{}
					}
				}
			}
			frontier = next
		}
		return result
	case '!':
		result := make(map[int]struct{})
		if node.emptyExt {
			// Minimatch's syntactic empty-ext flag is set when the raw
			// accumulator is empty at ')', including when the body ends in a
			// nested extglob. It renders that case as starNoEmpty.
			for end := start + 1; end <= len(input); end++ {
				result[end] = struct{}{}
			}
			return result
		}
		lookahead := node.lookaheadAlternatives
		if len(lookahead) == 0 {
			lookahead = node.alternatives
		}
		for _, alternative := range lookahead {
			if _, excluded := alternative.matchFrom(input, start)[len(input)]; excluded {
				return result
			}
		}
		for end := start; end <= len(input); end++ {
			result[end] = struct{}{}
		}
		return result
	default:
		return nil
	}
}

type searchRuneRange struct {
	low  rune
	high rune
}

type searchRuneClass struct {
	negated  bool
	never    bool
	byteMode bool
	ranges   []searchRuneRange
	posix    []string
}

func (class searchRuneClass) matches(character rune) bool {
	if class.never {
		return false
	}
	matched := false
	for _, characterRange := range class.ranges {
		if character >= characterRange.low && character <= characterRange.high {
			matched = true
			break
		}
	}
	if !matched {
		for _, name := range class.posix {
			if class.matchesPOSIX(name, character) {
				matched = true
				break
			}
		}
	}
	if class.negated {
		return !matched
	}
	return matched
}

func (class searchRuneClass) matchesPOSIX(name string, character rune) bool {
	if class.byteMode {
		return matchesSearchBytePOSIXClass(name, character)
	}
	return matchesSearchPOSIXClass(name, character)
}

func matchesSearchBytePOSIXClass(name string, character rune) bool {
	isLower := character >= 'a' && character <= 'z'
	isUpper := character >= 'A' && character <= 'Z'
	isDigit := character >= '0' && character <= '9'
	switch name {
	case "alnum":
		return isLower || isUpper || isDigit
	case "alpha":
		return isLower || isUpper
	case "ascii":
		return character <= unicode.MaxASCII
	case "blank":
		return character == ' ' || character == '\t'
	case "cntrl":
		return character < ' ' || character == unicode.MaxASCII
	case "digit":
		return isDigit
	case "graph":
		return character >= '!' && character <= '~'
	case "lower":
		return isLower
	case "print":
		return character >= ' ' && character <= '~'
	case "punct":
		return character >= '!' && character <= '~' && !isLower && !isUpper && !isDigit
	case "space":
		return strings.ContainsRune(" \t\r\n\v\f", character)
	case "upper":
		return isUpper
	case "word":
		return isLower || isUpper || isDigit || character == '_'
	case "xdigit":
		return isDigit || character >= 'a' && character <= 'f' || character >= 'A' && character <= 'F'
	default:
		return false
	}
}

func matchesSearchPOSIXClass(name string, character rune) bool {
	switch name {
	case "alnum":
		return isSearchUnicodeLetter(character) || isSearchUnicode16DecimalDigit(character)
	case "alpha":
		return isSearchUnicodeLetter(character)
	case "ascii":
		return character <= unicode.MaxASCII
	case "blank":
		return character == '\t' || unicode.Is(unicode.Zs, character)
	case "cntrl":
		return unicode.Is(unicode.Cc, character)
	case "digit":
		return isSearchUnicode16DecimalDigit(character)
	case "graph":
		return !unicode.Is(unicode.Categories["Z"], character) &&
			!isSearchUnicode16Other(character)
	case "lower":
		return isSearchUnicode16Lower(character)
	case "print":
		// minimatch 10.2.x currently translates [:print:] to Unicode's C
		// category (rather than its complement); mirror that exact behavior.
		return isSearchUnicode16Other(character)
	case "punct":
		return isSearchUnicode16Punctuation(character)
	case "space":
		return unicode.Is(unicode.Categories["Z"], character) ||
			strings.ContainsRune("\t\r\n\v\f", character)
	case "upper":
		return isSearchUnicode16Upper(character)
	case "word":
		return isSearchUnicodeLetter(character) || isSearchUnicode16DecimalDigit(character) ||
			unicode.Is(unicode.Pc, character)
	case "xdigit":
		return character >= '0' && character <= '9' ||
			character >= 'a' && character <= 'f' ||
			character >= 'A' && character <= 'F'
	default:
		return false
	}
}

func isSearchUnicodeLetter(character rune) bool {
	// JavaScript's Unicode Letter property includes the Letter Number (Nl)
	// category, while Go's unicode.IsLetter intentionally does not.
	return isSearchUnicode16Letter(character)
}

type searchSegmentParser struct {
	input    []rune
	index    int
	noExt    bool
	unitMode searchUnitMode
}

func (parser *searchSegmentParser) parseSequence(
	inExtglob bool,
	parentType rune,
	extglobDepth int,
) (searchSegmentSequence, rune, bool, error) {
	var sequence searchSegmentSequence
	var literal []rune
	// Minimatch tracks the raw accumulator separately from its parsed parts.
	// A successful nested extglob flushes the accumulator back to empty; this
	// syntactic detail controls the special negative empty-ext rendering.
	rawAccumulatorEmpty := true
	flushLiteral := func() {
		if len(literal) == 0 {
			return
		}
		sequence = append(sequence, searchSegmentNode{kind: searchSegmentLiteral, literal: literal})
		literal = nil
	}

	for parser.index < len(parser.input) {
		character := parser.input[parser.index]
		if inExtglob && (character == '|' || character == ')') {
			flushLiteral()
			parser.index++
			return sequence, character, rawAccumulatorEmpty, nil
		}
		if character == '\\' {
			rawAccumulatorEmpty = false
			parser.index++
			if parser.index >= len(parser.input) {
				literal = append(literal, '\\')
				continue
			}
			literal = append(literal, parser.input[parser.index])
			parser.index++
			continue
		}

		canAdopt := inExtglob && searchExtglobCanAdoptAny(parentType, character)
		if !parser.noExt && strings.ContainsRune("@?+*!", character) &&
			parser.index+1 < len(parser.input) && parser.input[parser.index+1] == '(' &&
			parser.hasClosingParen(parser.index+1) && (extglobDepth <= 2 || canAdopt) {
			flushLiteral()
			parser.index += 2
			childDepth := extglobDepth + 1
			if canAdopt {
				childDepth = extglobDepth
			}
			alternatives, err := parser.parseExtglobAlternatives(character, childDepth)
			if err != nil {
				return nil, 0, false, err
			}
			sequence = append(sequence, searchSegmentNode{
				kind:         searchSegmentExtglob,
				extglobType:  character,
				alternatives: alternatives.sequences,
				emptyExt:     alternatives.emptyExt,
			})
			rawAccumulatorEmpty = true
			continue
		}

		rawAccumulatorEmpty = false
		switch character {
		case '*':
			flushLiteral()
			if len(sequence) == 0 || sequence[len(sequence)-1].kind != searchSegmentStar {
				sequence = append(sequence, searchSegmentNode{kind: searchSegmentStar})
			}
			parser.index++
		case '?':
			flushLiteral()
			sequence = append(sequence, searchSegmentNode{kind: searchSegmentQuestion})
			parser.index++
		case '[':
			class, ok := parser.parseClass()
			if !ok {
				literal = append(literal, '[')
				parser.index++
				continue
			}
			flushLiteral()
			sequence = append(sequence, searchSegmentNode{kind: searchSegmentClass, class: class})
		default:
			literal = append(literal, character)
			parser.index++
		}
	}
	flushLiteral()
	if inExtglob {
		return nil, 0, false, errors.New("unclosed extglob")
	}
	return sequence, 0, rawAccumulatorEmpty, nil
}

func (parser *searchSegmentParser) hasClosingParen(opening int) bool {
	depth := 0
	for index := opening; index < len(parser.input); index++ {
		if parser.input[index] == '\\' {
			index++
			continue
		}
		switch parser.input[index] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return true
			}
		}
	}
	return false
}

type parsedSearchExtglobAlternatives struct {
	sequences []searchSegmentSequence
	emptyExt  bool
}

func (parser *searchSegmentParser) parseExtglobAlternatives(
	parentType rune,
	extglobDepth int,
) (parsedSearchExtglobAlternatives, error) {
	var alternatives []searchSegmentSequence
	for {
		alternative, terminator, rawAccumulatorEmpty, err := parser.parseSequence(true, parentType, extglobDepth)
		if err != nil {
			return parsedSearchExtglobAlternatives{}, err
		}
		alternatives = append(alternatives, alternative)
		switch terminator {
		case '|':
			continue
		case ')':
			return parsedSearchExtglobAlternatives{
				sequences: alternatives,
				emptyExt:  rawAccumulatorEmpty,
			}, nil
		default:
			return parsedSearchExtglobAlternatives{}, errors.New("unclosed extglob")
		}
	}
}

func (parser *searchSegmentParser) parseClass() (searchRuneClass, bool) {
	index := parser.index + 1
	class := searchRuneClass{byteMode: parser.unitMode == searchUnitModeBytes}
	if index < len(parser.input) && (parser.input[index] == '!' || parser.input[index] == '^') {
		class.negated = true
		index++
	}

	type classItem struct {
		character rune
		posix     string
	}
	var items []classItem
	if index < len(parser.input) && parser.input[index] == ']' {
		items = append(items, classItem{character: ']'})
		index++
	}
	for index < len(parser.input) && parser.input[index] != ']' {
		if parser.input[index] == '[' && index+3 < len(parser.input) && parser.input[index+1] == ':' {
			closing := index + 2
			for closing+1 < len(parser.input) && (parser.input[closing] != ':' || parser.input[closing+1] != ']') {
				closing++
			}
			if closing+1 < len(parser.input) {
				name := string(parser.input[index+2 : closing])
				if isSearchPOSIXClass(name) {
					items = append(items, classItem{posix: name})
					index = closing + 2
					continue
				}
			}
		}
		character := parser.input[index]
		if character == '\\' {
			index++
			if index >= len(parser.input) {
				return searchRuneClass{}, false
			}
			character = parser.input[index]
		}
		items = append(items, classItem{character: character})
		index++
	}
	if index >= len(parser.input) || parser.input[index] != ']' {
		return searchRuneClass{}, false
	}
	parser.index = index + 1
	if len(items) == 0 {
		return searchRuneClass{}, false
	}

	for index := 0; index < len(items); index++ {
		if items[index].posix != "" {
			class.posix = append(class.posix, items[index].posix)
			continue
		}
		if index+2 < len(items) && items[index+1].posix == "" && items[index+1].character == '-' && items[index+2].posix == "" {
			if items[index].character > items[index+2].character {
				class.never = true
				return class, true
			}
			class.ranges = append(class.ranges, searchRuneRange{low: items[index].character, high: items[index+2].character})
			index += 2
			continue
		}
		class.ranges = append(class.ranges, searchRuneRange{low: items[index].character, high: items[index].character})
	}
	return class, true
}

func isSearchPOSIXClass(name string) bool {
	switch name {
	case "alnum", "alpha", "ascii", "blank", "cntrl", "digit", "graph", "lower", "print", "punct", "space", "upper", "word", "xdigit":
		return true
	default:
		return false
	}
}

func searchPOSIXClassRequiresUnicode(name string) bool {
	// brace-expressions emits plain ASCII ranges for these two classes. Every
	// other supported class uses a Unicode property escape and makes the
	// JavaScript regexp for this path segment use the /u flag.
	return name != "ascii" && name != "xdigit"
}

func expandSearchBraces(pattern string) ([]string, error) {
	tokens := newBraceEscapeTokens(pattern)
	escaped := escapeSearchBraces(pattern, tokens)
	if strings.HasPrefix(escaped, "{}") {
		escaped = tokens.open + tokens.close + escaped[2:]
	}
	expanded := expandSearchBracesRecursive(escaped, maxBraceExpansions, true, tokens)
	seen := make(map[string]struct{}, len(expanded))
	result := make([]string, 0, len(expanded))
	for _, item := range expanded {
		item = unescapeSearchBraces(item, tokens)
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result, nil
}

type braceEscapeTokens struct {
	slash  string
	open   string
	close  string
	comma  string
	period string
}

func newBraceEscapeTokens(pattern string) braceEscapeTokens {
	prefix := "\x00RSLINT_BRACE\x00"
	for strings.Contains(pattern, prefix) {
		prefix += "X"
	}
	return braceEscapeTokens{
		slash:  prefix + "SLASH\x00",
		open:   prefix + "OPEN\x00",
		close:  prefix + "CLOSE\x00",
		comma:  prefix + "COMMA\x00",
		period: prefix + "PERIOD\x00",
	}
}

func escapeSearchBraces(pattern string, tokens braceEscapeTokens) string {
	var result strings.Builder
	result.Grow(len(pattern))
	for index := 0; index < len(pattern); index++ {
		if pattern[index] != '\\' || index+1 >= len(pattern) {
			result.WriteByte(pattern[index])
			continue
		}
		var replacement string
		switch pattern[index+1] {
		case '\\':
			replacement = tokens.slash
		case '{':
			replacement = tokens.open
		case '}':
			replacement = tokens.close
		case ',':
			replacement = tokens.comma
		case '.':
			replacement = tokens.period
		default:
			result.WriteByte(pattern[index])
			continue
		}
		result.WriteString(replacement)
		index++
	}
	return result.String()
}

func unescapeSearchBraces(pattern string, tokens braceEscapeTokens) string {
	replacer := strings.NewReplacer(
		tokens.slash, "\\",
		tokens.open, "{",
		tokens.close, "}",
		tokens.comma, ",",
		tokens.period, ".",
	)
	return replacer.Replace(pattern)
}

type balancedBraceMatch struct {
	pre  string
	body string
	post string
}

func findBalancedSearchBrace(pattern string) (balancedBraceMatch, bool) {
	opening := strings.IndexByte(pattern, '{')
	if opening < 0 {
		return balancedBraceMatch{}, false
	}
	closingOffset := strings.IndexByte(pattern[opening+1:], '}')
	if closingOffset < 0 {
		return balancedBraceMatch{}, false
	}
	closing := opening + 1 + closingOffset
	stack := make([]int, 0, 4)
	left := len(pattern)
	right := -1
	for nextOpening, nextClosing := opening, closing; nextOpening >= 0 || nextClosing >= 0; {
		if nextOpening >= 0 && (nextClosing < 0 || nextOpening < nextClosing) {
			stack = append(stack, nextOpening)
			if offset := strings.IndexByte(pattern[nextOpening+1:], '{'); offset >= 0 {
				nextOpening += 1 + offset
			} else {
				nextOpening = -1
			}
			continue
		}
		if len(stack) == 1 {
			start := stack[0]
			return balancedBraceMatch{
				pre:  pattern[:start],
				body: pattern[start+1 : nextClosing],
				post: pattern[nextClosing+1:],
			}, true
		}
		start := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if start < left {
			left = start
			right = nextClosing
		}
		if offset := strings.IndexByte(pattern[nextClosing+1:], '}'); offset >= 0 {
			nextClosing += 1 + offset
		} else {
			nextClosing = -1
		}
	}
	if len(stack) > 0 && right >= 0 {
		return balancedBraceMatch{
			pre:  pattern[:left],
			body: pattern[left+1 : right],
			post: pattern[right+1:],
		}, true
	}
	return balancedBraceMatch{}, false
}

func parseSearchBraceCommaParts(pattern string) []string {
	if pattern == "" {
		return []string{""}
	}
	match, ok := findBalancedSearchBrace(pattern)
	if !ok {
		return strings.Split(pattern, ",")
	}
	parts := strings.Split(match.pre, ",")
	last := len(parts) - 1
	parts[last] += "{" + match.body + "}"
	postParts := parseSearchBraceCommaParts(match.post)
	if match.post != "" {
		parts[last] += postParts[0]
		parts = append(parts, postParts[1:]...)
	}
	return parts
}

var (
	numericSearchBraceSequence = regexp.MustCompile(`^-?[0-9]+\.\.-?[0-9]+(?:\.\.-?[0-9]+)?$`)
	alphaSearchBraceSequence   = regexp.MustCompile(`^[a-zA-Z]\.\.[a-zA-Z](?:\.\.-?[0-9]+)?$`)
)

func expandSearchBracesRecursive(
	pattern string,
	limit int,
	top bool,
	tokens braceEscapeTokens,
) []string {
	match, ok := findBalancedSearchBrace(pattern)
	if !ok {
		return []string{pattern}
	}
	post := []string{""}
	if match.post != "" {
		post = expandSearchBracesRecursive(match.post, limit, false, tokens)
	}
	result := make([]string, 0)
	if strings.HasSuffix(match.pre, "$") {
		for index := range min(len(post), limit) {
			result = append(result, match.pre+"{"+match.body+"}"+post[index])
		}
		return result
	}

	numericSequence := numericSearchBraceSequence.MatchString(match.body)
	alphaSequence := alphaSearchBraceSequence.MatchString(match.body)
	sequence := numericSequence || alphaSequence
	options := strings.Contains(match.body, ",")
	if !sequence && !options {
		if searchBracePostHasRecoverableClose(match.post) {
			recovered := match.pre + "{" + match.body + tokens.close + match.post
			return expandSearchBracesRecursive(recovered, limit, true, tokens)
		}
		return []string{pattern}
	}

	var items []string
	if sequence {
		items = expandSearchBraceSequence(match.body, alphaSequence, limit)
	} else {
		parts := parseSearchBraceCommaParts(match.body)
		if len(parts) == 1 {
			nested := expandSearchBracesRecursive(parts[0], limit, false, tokens)
			items = make([]string, 0, len(nested))
			for _, item := range nested {
				items = append(items, "{"+item+"}")
			}
			if len(items) == 1 {
				result := make([]string, 0, len(post))
				for _, suffix := range post {
					result = append(result, match.pre+items[0]+suffix)
				}
				return result
			}
		} else {
			for _, part := range parts {
				items = append(items, expandSearchBracesRecursive(part, limit, false, tokens)...)
			}
		}
	}
	for _, item := range items {
		for _, suffix := range post {
			if len(result) >= limit {
				return result
			}
			expansion := match.pre + item + suffix
			if !top || sequence || expansion != "" {
				result = append(result, expansion)
			}
		}
	}
	return result
}

func searchBracePostHasRecoverableClose(post string) bool {
	for index := range post {
		if post[index] != ',' || index+1 < len(post) && post[index+1] == ',' {
			continue
		}
	search:
		for _, character := range post[index+1:] {
			switch character {
			case '}':
				return true
			case '\n', '\r', '\u2028', '\u2029':
				break search
			}
		}
	}
	return false
}

func expandSearchBraceSequence(body string, alpha bool, limit int) []string {
	parts := strings.Split(body, "..")
	if len(parts) < 2 {
		return nil
	}
	start, end := 0, 0
	if alpha {
		start, end = int(parts[0][0]), int(parts[1][0])
	} else {
		var err error
		start, err = strconv.Atoi(parts[0])
		if err != nil {
			return nil
		}
		end, err = strconv.Atoi(parts[1])
		if err != nil {
			return nil
		}
	}
	step := 1
	if len(parts) == 3 {
		parsed, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil
		}
		if parsed < 0 {
			parsed = -parsed
		}
		if parsed > step {
			step = parsed
		}
	}
	if end < start {
		step = -step
	}
	padded := false
	for _, part := range parts {
		trimmed := strings.TrimPrefix(part, "-")
		if len(trimmed) > 1 && trimmed[0] == '0' {
			padded = true
			break
		}
	}
	width := len(parts[0])
	if len(parts[1]) > width {
		width = len(parts[1])
	}
	result := make([]string, 0)
	for value := start; len(result) < limit; value += step {
		if step > 0 && value > end || step < 0 && value < end {
			break
		}
		if alpha {
			if value == int('\\') {
				result = append(result, "")
			} else {
				result = append(result, string(rune(value)))
			}
			continue
		}
		formatted := strconv.Itoa(value)
		if padded {
			need := width - len(formatted)
			if need > 0 {
				zeros := strings.Repeat("0", need)
				if value < 0 {
					formatted = "-" + zeros + formatted[1:]
				} else {
					formatted = zeros + formatted
				}
			}
		}
		result = append(result, formatted)
	}
	return result
}
