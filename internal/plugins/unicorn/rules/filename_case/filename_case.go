package filename_case

import (
	_ "embed"
	"fmt"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

//go:embed filename_case.schema.json
var schemaJSON []byte

// caseStyle is one of the supported filename case styles.
type caseStyle struct {
	key  string
	name string
	fn   func(string) string
}

const caseStyleCount = 5

// allCases follows the upstream case registry order. Go maps do not preserve
// the authored order of `cases`, so this order also keeps diagnostics and
// suggestions deterministic when multiple cases are enabled.
var allCases = [caseStyleCount]caseStyle{
	{key: "camelCase", name: "camel case", fn: toCamelCase},
	{key: "camelCaseWithAcronyms", name: "camel case with acronyms", fn: toCamelCaseWithAcronyms},
	{key: "kebabCase", name: "kebab case", fn: toKebabCase},
	{key: "snakeCase", name: "snake case", fn: toSnakeCase},
	{key: "pascalCase", name: "pascal case", fn: toPascalCaseWithLeadingAcronym},
}

func caseForKey(key string) (caseStyle, bool) {
	for _, c := range allCases {
		if c.key == key {
			return c, true
		}
	}
	return caseStyle{}, false
}

// isIgnoredByDefault mirrors the upstream's hardcoded set of files that cannot
// change case (notably required by Node / build tooling).
func isIgnoredByDefault(basename string) bool {
	switch basename {
	case "index.js", "index.mjs", "index.cjs", "index.ts", "index.tsx", "index.vue":
		return true
	default:
		return false
	}
}

// invalidIgnore captures a single user-supplied `ignore` pattern that failed
// to compile. The rule reports each one as its own diagnostic so the user can
// see which configuration entry is broken.
type invalidIgnore struct {
	pattern string
	err     error
}

// options is the parsed shape of the user's rule configuration. Cases use a
// fixed-size array because the schema permits exactly five known styles; this
// keeps the option parsing done for every file allocation-free unless ignore
// patterns are configured.
type options struct {
	cases                  [caseStyleCount]caseStyle
	caseCount              int
	ignores                []*esregexp.RegExp
	invalidIgnores         []invalidIgnore
	multipleFileExtensions bool
	checkDirectories       bool
}

func parseOptions(rawOpts []any) options {
	opts := options{caseCount: 1, multipleFileExtensions: true, checkDirectories: true}
	opts.cases[0] = allCases[2] // kebabCase
	if len(rawOpts) == 0 {
		return opts
	}
	optsMap, _ := rawOpts[0].(map[string]any)

	if v, ok := optsMap["case"].(string); ok {
		if c, found := caseForKey(v); found {
			opts.cases[0] = c
			opts.caseCount = 1
		}
	} else if casesMap, ok := optsMap["cases"].(map[string]interface{}); ok {
		var chosen [caseStyleCount]caseStyle
		chosenCount := 0
		for _, c := range allCases {
			if b, ok := casesMap[c.key].(bool); ok && b {
				chosen[chosenCount] = c
				chosenCount++
			}
		}
		if chosenCount > 0 {
			opts.cases = chosen
			opts.caseCount = chosenCount
		}
	}

	if v, ok := optsMap["ignore"].([]interface{}); ok {
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				continue
			}
			if re, err := compileIgnoreRegexp(s); err == nil {
				opts.ignores = append(opts.ignores, re)
			} else {
				opts.invalidIgnores = append(opts.invalidIgnores, invalidIgnore{pattern: s, err: err})
			}
		}
	}

	if v, ok := optsMap["multipleFileExtensions"].(bool); ok {
		opts.multipleFileExtensions = v
	}
	if v, ok := optsMap["checkDirectories"].(bool); ok {
		opts.checkDirectories = v
	}

	return opts
}

func (o *options) selectedCases() []caseStyle {
	return o.cases[:o.caseCount]
}

// Compiling is considerably more expensive than matching, while a
// configured ignore list is shared by every file. Cache compiled patterns at
// rule scope so parallel files reuse them. The fixed-size FIFO bound prevents
// long-lived API/LSP processes from retaining an unbounded stream of config
// values; a compiled pattern is safe for concurrent use.
const ignoreRegexpCacheCapacity = 128

type compiledIgnoreRegexp struct {
	regexp *esregexp.RegExp
	err    error
}

var ignoreRegexpCache = struct {
	sync.Mutex
	entries map[string]compiledIgnoreRegexp
	order   [ignoreRegexpCacheCapacity]string
	count   int
	next    int
}{}

func compileIgnoreRegexp(pattern string) (*esregexp.RegExp, error) {
	ignoreRegexpCache.Lock()
	defer ignoreRegexpCache.Unlock()

	if cached, ok := ignoreRegexpCache.entries[pattern]; ok {
		return cached.regexp, cached.err
	}

	re, err := esregexp.Compile(pattern, "u")
	if ignoreRegexpCache.entries == nil {
		ignoreRegexpCache.entries = make(map[string]compiledIgnoreRegexp, ignoreRegexpCacheCapacity)
	}
	if ignoreRegexpCache.count == ignoreRegexpCacheCapacity {
		delete(ignoreRegexpCache.entries, ignoreRegexpCache.order[ignoreRegexpCache.next])
	} else {
		ignoreRegexpCache.count++
	}
	ignoreRegexpCache.order[ignoreRegexpCache.next] = pattern
	ignoreRegexpCache.next = (ignoreRegexpCache.next + 1) % ignoreRegexpCacheCapacity
	ignoreRegexpCache.entries[pattern] = compiledIgnoreRegexp{regexp: re, err: err}
	return re, err
}

// nodeExtname mirrors Node.js `path.extname`. Returns the suffix from the
// last `.`, with these special cases (matching Node's behaviour):
//
//   - no dot in basename            → ""        (e.g. `foo`)
//   - leading-only dot              → ""        (e.g. `.foo`)
//   - basename is all dots          → ""        (e.g. `..`, `...`)
//   - trailing dot                  → "."       (e.g. `foo.`)
//   - regular extension             → ".<ext>"  (e.g. `foo.js`, `.foo.js`)
//
// `tspath.GetAnyExtensionFromPath` returns `.foo` for a basename like `.foo`
// (Go-style), which is not what we want for hidden-style filenames such as
// `.test_utils.js`.
func nodeExtname(basename string) string {
	lastDot := strings.LastIndex(basename, ".")
	if lastDot <= 0 {
		return ""
	}
	// If the basename is composed entirely of dots (`..`, `...`, etc.) Node
	// treats it as extensionless. `..js` / `...js` are NOT all-dots — they
	// have real characters and Node returns `.js`.
	allDots := true
	for i := range len(basename) {
		if basename[i] != '.' {
			allDots = false
			break
		}
	}
	if allDots {
		return ""
	}
	return basename[lastDot:]
}

// splitWords reproduces change-case@5.4's `split()`. The exact upstream regex
// pipeline is:
//
//  1. /([\p{Ll}\d])(\p{Lu})/gu       → insert delimiter before an uppercase
//     that follows a lowercase or digit.
//  2. /(\p{Lu})([\p{Lu}][\p{Ll}])/gu → insert delimiter between two
//     uppercases when the second is the
//     start of a TitleCase word
//     (e.g. `XMLHttp` → `XML Http`).
//  3. /[^\p{L}\d]+/giu               → collapse non-alphanumeric runs into
//     the same delimiter.
//
// Then trim leading/trailing delimiters and split. The three replacements only
// establish word boundaries, so the implementation below recognizes the same
// boundaries in one pass and returns slices of the original string. This
// avoids building three intermediate rune buffers for every case candidate.
func splitWords(s string) []string {
	var words []string
	wordStart := -1
	var previous rune
	for pos, current := range s {
		if !unicode.IsLetter(current) && !isASCIIDigit(current) {
			if wordStart >= 0 {
				words = append(words, s[wordStart:pos])
				wordStart = -1
			}
			continue
		}

		if wordStart < 0 {
			wordStart = pos
		} else {
			boundary := (unicode.IsLower(previous) || isASCIIDigit(previous)) && unicode.IsUpper(current)
			if !boundary && unicode.IsUpper(previous) && unicode.IsUpper(current) {
				_, size := utf8.DecodeRuneInString(s[pos:])
				next, _ := utf8.DecodeRuneInString(s[pos+size:])
				boundary = unicode.IsLower(next)
			}
			if boundary {
				words = append(words, s[wordStart:pos])
				wordStart = pos
			}
		}
		previous = current
	}
	if wordStart >= 0 {
		words = append(words, s[wordStart:])
	}
	return words
}

func isASCIIDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// pascalLikeTransform is change-case@5.4's `pascalCaseTransformFactory`:
// when a non-first word starts with a digit, prepend `_` to keep the join
// readable and round-trippable; otherwise capitalize the first letter.
func pascalLikeTransform(word string, index int) string {
	if word == "" {
		return ""
	}
	char0, size := utf8.DecodeRuneInString(word)
	first := word[:size]
	rest := ecmascript.StringToLowerCase(word[size:])
	if index > 0 && isASCIIDigit(char0) {
		return "_" + first + rest
	}
	return ecmascript.StringToUpperCase(first) + rest
}

func toCamelCase(s string) string {
	words := splitWords(s)
	if len(words) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.Grow(len(s) + len(words))
	sb.WriteString(ecmascript.StringToLowerCase(words[0]))
	for i := 1; i < len(words); i++ {
		sb.WriteString(pascalLikeTransform(words[i], i))
	}
	return sb.String()
}

func toCamelCaseWithAcronyms(s string) string {
	if isCamelCaseWithAcronyms(s) {
		return s
	}
	return toCamelCase(s)
}

func isCamelCaseWithAcronyms(s string) bool {
	if len(s) == 0 || !isASCIILower(s[0]) {
		return false
	}
	for index := 1; index < len(s); index++ {
		char := s[index]
		if isASCIILower(char) || isASCIIByteDigit(char) {
			continue
		}
		if !isASCIIUpper(char) {
			return false
		}

		uppercaseStart := index
		for index+1 < len(s) && isASCIIUpper(s[index+1]) {
			index++
		}
		if index == uppercaseStart {
			continue
		}
		if index+1 < len(s) && isASCIILower(s[index+1]) {
			index--
			continue
		}
		for index+1 < len(s) && isASCIIByteDigit(s[index+1]) {
			index++
		}
		if index == len(s)-1 {
			return true
		}
		if !isASCIIUpper(s[index+1]) {
			return false
		}
	}
	return true
}

func isASCIILower(char byte) bool     { return char >= 'a' && char <= 'z' }
func isASCIIUpper(char byte) bool     { return char >= 'A' && char <= 'Z' }
func isASCIIByteDigit(char byte) bool { return char >= '0' && char <= '9' }

func toPascalCase(s string) string {
	words := splitWords(s)
	if len(words) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.Grow(len(s) + len(words))
	for i, w := range words {
		sb.WriteString(pascalLikeTransform(w, i))
	}
	return sb.String()
}

func toPascalCaseWithLeadingAcronym(s string) string {
	if acronymEnd := leadingPascalAcronymEnd(s); acronymEnd > 0 {
		suffix := s[acronymEnd:]
		if toPascalCase(suffix) == suffix {
			return s
		}
	}
	return toPascalCase(s)
}

// leadingPascalAcronymEnd matches upstream's
// /^[A-Z]{3,}(?=\d*[A-Z](?:[a-z]|\d+[a-z]))/ without routing source text
// through a regexp engine.
func leadingPascalAcronymEnd(s string) int {
	for index := range len(s) {
		char := s[index]
		if !isASCIILower(char) && !isASCIIUpper(char) && !isASCIIByteDigit(char) {
			return 0
		}
	}

	uppercaseEnd := 0
	for uppercaseEnd < len(s) && isASCIIUpper(s[uppercaseEnd]) {
		uppercaseEnd++
	}
	for acronymEnd := uppercaseEnd; acronymEnd >= 3; acronymEnd-- {
		index := acronymEnd
		for index < len(s) && isASCIIByteDigit(s[index]) {
			index++
		}
		if index >= len(s) || !isASCIIUpper(s[index]) {
			continue
		}
		index++
		if index < len(s) && isASCIILower(s[index]) {
			return acronymEnd
		}
		digitStart := index
		for index < len(s) && isASCIIByteDigit(s[index]) {
			index++
		}
		if index > digitStart && index < len(s) && isASCIILower(s[index]) {
			return acronymEnd
		}
	}
	return 0
}

func toKebabCase(s string) string { return joinNoCase(splitWords(s), "-") }
func toSnakeCase(s string) string { return joinNoCase(splitWords(s), "_") }

func joinNoCase(words []string, delim string) string {
	if len(words) == 0 {
		return ""
	}
	size := len(delim) * (len(words) - 1)
	for _, word := range words {
		size += len(word)
	}
	var sb strings.Builder
	sb.Grow(size)
	for i, w := range words {
		if i > 0 {
			sb.WriteString(delim)
		}
		sb.WriteString(ecmascript.StringToLowerCase(w))
	}
	return sb.String()
}

// filenameWord is one chunk of the filename produced by splitFilename: a run
// of either filename-relevant characters (letters/digits/`-`/`_`) or
// "decoration" characters (`[`, `]`, `$`, …) the rule should preserve verbatim.
type filenameWord struct {
	word    string
	ignored bool
}

// isIgnoredChar mirrors the upstream's `/^[a-z\d-_]$/i` test: returns true
// when a character is NOT one of the case-relevant filename characters.
func isIgnoredChar(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z':
		return false
	case r >= 'A' && r <= 'Z':
		return false
	case r >= '0' && r <= '9':
		return false
	case r == '-', r == '_':
		return false
	}
	return true
}

func appendFilenameWord(words []filenameWord, word string, ignored bool, invalidUTF8 bool) []filenameWord {
	if invalidUTF8 {
		// Match range-over-runes semantics from the original implementation:
		// malformed bytes become U+FFFD instead of leaking invalid UTF-8 into a
		// diagnostic suggestion. Valid filenames stay on the zero-copy path.
		word = string([]rune(word))
	}
	return append(words, filenameWord{word: word, ignored: ignored})
}

// splitFilename mirrors the upstream helper of the same name. Leading
// underscores are captured separately so they're preserved verbatim in the
// rename suggestion.
func splitFilename(filename string) (leading string, words []filenameWord) {
	i := 0
	for i < len(filename) && filename[i] == '_' {
		i++
	}
	leading = filename[:i]
	tailing := filename[i:]
	if tailing == "" {
		return leading, nil
	}

	wordStart := 0
	var lastIgnored bool
	wordHasInvalidUTF8 := false
	hasLast := false
	for pos, r := range tailing {
		ignored := isIgnoredChar(r)
		invalidUTF8 := false
		if r == utf8.RuneError {
			_, size := utf8.DecodeRuneInString(tailing[pos:])
			invalidUTF8 = size == 1
		}
		if !hasLast {
			lastIgnored = ignored
			wordHasInvalidUTF8 = invalidUTF8
			hasLast = true
			continue
		}
		if lastIgnored != ignored {
			words = appendFilenameWord(words, tailing[wordStart:pos], lastIgnored, wordHasInvalidUTF8)
			wordStart = pos
			lastIgnored = ignored
			wordHasInvalidUTF8 = invalidUTF8
		} else if invalidUTF8 {
			wordHasInvalidUTF8 = true
		}
	}
	words = appendFilenameWord(words, tailing[wordStart:], lastIgnored, wordHasInvalidUTF8)
	return leading, words
}

type caseCandidates [caseStyleCount]string

// validateFilename returns whether every non-ignored chunk already matches at
// least one chosen case style. On failure it also returns the first invalid
// chunk and the candidates already computed for it, so fixFilename does not
// repeat those conversions.
func validateFilename(words []filenameWord, cases []caseStyle) (valid bool, invalidWord int, candidates caseCandidates) {
	for wordIndex, w := range words {
		if w.ignored {
			continue
		}
		for caseIndex, c := range cases {
			candidate := c.fn(w.word)
			if candidate == w.word {
				candidates = caseCandidates{}
				break
			}
			candidates[caseIndex] = candidate
			if caseIndex == len(cases)-1 {
				return false, wordIndex, candidates
			}
		}
	}
	return true, -1, caseCandidates{}
}

// fixFilename builds the deduplicated, ordered list of suggested filenames by
// taking the cartesian product of each non-ignored chunk's case-conversion
// candidates. Candidate order follows allCases so output remains stable even
// though the source configuration was decoded into a Go map.
type filenameReplacements struct {
	items [caseStyleCount]string
	count int
}

func (r *filenameReplacements) add(candidate string) {
	for i := range r.count {
		if r.items[i] == candidate {
			return
		}
	}
	r.items[r.count] = candidate
	r.count++
}

func fixFilename(
	words []filenameWord,
	cases []caseStyle,
	invalidWord int,
	invalidCandidates caseCandidates,
	leading string,
	trailing string,
) []string {
	replacements := make([]filenameReplacements, len(words))
	combinationCount := 1
	maxFilenameLen := len(leading) + len(trailing)
	for i, w := range words {
		if w.ignored {
			replacements[i].add(w.word)
		} else {
			for j, c := range cases {
				candidate := invalidCandidates[j]
				if i != invalidWord {
					candidate = c.fn(w.word)
				}
				replacements[i].add(candidate)
			}
		}

		// Case candidates contain only filename-relevant characters, while
		// adjacent converted chunks are separated by a fixed ignored chunk.
		// Removing duplicates inside each chunk therefore also guarantees that
		// the complete cartesian-product outputs are unique, without a map.
		maxReplacementLen := 0
		for j := range replacements[i].count {
			maxReplacementLen = max(maxReplacementLen, len(replacements[i].items[j]))
		}
		maxFilenameLen += maxReplacementLen
		if combinationCount <= 256/replacements[i].count {
			combinationCount *= replacements[i].count
		} else {
			combinationCount = 256
		}
	}

	out := make([]string, 0, combinationCount)
	buffer := make([]byte, 0, maxFilenameLen)
	buffer = append(buffer, leading...)
	var visit func(idx int)
	visit = func(idx int) {
		if idx == len(replacements) {
			prefixLen := len(buffer)
			buffer = append(buffer, trailing...)
			out = append(out, string(buffer))
			buffer = buffer[:prefixLen]
			return
		}
		for i := range replacements[idx].count {
			prefixLen := len(buffer)
			buffer = append(buffer, replacements[idx].items[i]...)
			visit(idx + 1)
			buffer = buffer[:prefixLen]
		}
	}
	visit(0)
	return out
}

// englishishJoin reproduces `Intl.ListFormat('en-US', {type: 'disjunction'})`:
// `a`, `a or b`, `a, b, or c`, …
func englishishJoin(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return items[0] + " or " + items[1]
	}
	return strings.Join(items[:len(items)-1], ", ") + ", or " + items[len(items)-1]
}

func englishishCaseNames(cases []caseStyle) string {
	var names [caseStyleCount]string
	for i, c := range cases {
		names[i] = c.name
	}
	return englishishJoin(names[:len(cases)])
}

func englishishBacktickJoin(items []string) string {
	if len(items) == 0 {
		return ""
	}
	size := 2 * len(items)
	for _, item := range items {
		size += len(item)
	}
	var sb strings.Builder
	sb.Grow(size + 6*(len(items)-1))
	for i, item := range items {
		if i > 0 {
			switch {
			case i < len(items)-1:
				sb.WriteString(", ")
			case len(items) == 2:
				sb.WriteString(" or ")
			default:
				sb.WriteString(", or ")
			}
		}
		sb.WriteByte('`')
		sb.WriteString(item)
		sb.WriteByte('`')
	}
	return sb.String()
}

func isInsideCwd(relativePath string) bool {
	return relativePath != "" &&
		relativePath != ".." &&
		!strings.HasPrefix(relativePath, "../") &&
		!tspath.IsRootedDiskPath(relativePath)
}

// getCanonicalPOSIXPathSegments handles the hot path produced by the normal
// lint planner: absolute, normalized POSIX filenames under an absolute,
// normalized cwd. Less common path forms fall back to tspath below so their
// Node-compatible normalization stays centralized there.
func getCanonicalPOSIXPathSegments(fileName string, cwd string) ([]string, bool) {
	if !isCanonicalAbsolutePOSIXPath(fileName) || !isCanonicalAbsolutePOSIXPath(cwd) {
		return nil, false
	}

	basename := fileName[strings.LastIndexByte(fileName, '/')+1:]
	if fileName == cwd {
		return []string{basename}, true
	}

	var relative string
	if cwd == "/" {
		relative = fileName[1:]
	} else {
		if !strings.HasPrefix(fileName, cwd) || len(fileName) <= len(cwd) || fileName[len(cwd)] != '/' {
			return []string{basename}, true
		}
		relative = fileName[len(cwd)+1:]
	}
	if relative == "" {
		return []string{basename}, true
	}
	return strings.Split(relative, "/"), true
}

func isCanonicalAbsolutePOSIXPath(path string) bool {
	if path == "/" {
		return true
	}
	if len(path) < 2 || path[0] != '/' || path[len(path)-1] == '/' || strings.IndexByte(path, '\\') >= 0 {
		return false
	}

	segmentStart := 1
	for index := 1; index <= len(path); index++ {
		if index < len(path) && path[index] != '/' {
			continue
		}
		segment := path[segmentStart:index]
		if segment == "" || segment == "." || segment == ".." || segment == "^" ||
			len(segment) == 2 && tspath.IsVolumeCharacter(segment[0]) && segment[1] == ':' {
			return false
		}
		segmentStart = index + 1
	}
	return true
}

func getPathSegments(fileName string, cwd string) []string {
	if segments, ok := getCanonicalPOSIXPathSegments(fileName, cwd); ok {
		return segments
	}
	return getPathSegmentsWithTspath(fileName, cwd)
}

func getPathSegmentsWithTspath(fileName string, cwd string) []string {
	basename := tspath.GetBaseFileName(fileName)
	if cwd == "" {
		return []string{basename}
	}
	resolved := tspath.ResolvePath(cwd, fileName)
	if (tspath.GetRootLength(cwd) > 0) != (tspath.GetRootLength(resolved) > 0) {
		return []string{basename}
	}
	relative := tspath.GetRelativePathFromDirectory(cwd, resolved, tspath.ComparePathsOptions{
		CurrentDirectory:          cwd,
		UseCaseSensitiveFileNames: true,
	})
	if !isInsideCwd(relative) {
		return []string{basename}
	}
	segments := strings.Split(relative, "/")
	filtered := segments[:0]
	for _, segment := range segments {
		if segment != "." && segment != "" {
			filtered = append(filtered, segment)
		}
	}
	if len(filtered) == 0 {
		return []string{basename}
	}
	return filtered
}

const (
	filenameCaseRuleName       = "unicorn/filename-case"
	messageIDFilenameCase      = "filename-case"
	messageIDDirectoryCase     = "directory-case"
	messageIDFilenameExtension = "filename-extension"
)

var FilenameCaseRule = rule.Rule{
	Name:   filenameCaseRuleName,
	Schema: rule.NewSchema(schemaJSON),
	// The rule is purely filename-driven — it does not inspect any AST node.
	// `Run` is invoked once per source file, so we do the work here and
	// return an empty listener map. (The linter's visitor walks SourceFile
	// children but never the SourceFile node itself, so a
	// `ast.KindSourceFile` listener would silently never fire.)
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if ctx.SourceFile == nil {
			return nil
		}
		fileName := ctx.SourceFile.FileName()
		if fileName == "" {
			return nil
		}
		pathSegments := getPathSegments(fileName, ctx.ProcessCurrentDirectory())
		basename := pathSegments[len(pathSegments)-1]
		// Skip ESLint's stdin / inline-source virtual filenames.
		if basename == "<input>" || basename == "<text>" {
			return nil
		}

		opts := parseOptions(options)
		reportRange := core.NewTextRange(0, 0)

		// Configuration error: any malformed `ignore` pattern aborts
		// case-checking on this file. Mirrors ESLint's behaviour, where
		// `new RegExp(item, 'u')` throws at rule-create time and the rule
		// produces no further diagnostics until the config is fixed —
		// returning case reports based on a partially-broken ignore list
		// would be misleading.
		if len(opts.invalidIgnores) > 0 {
			if ctx.DisableManager.IsRuleDisabled(filenameCaseRuleName, reportRange.Pos()) {
				return nil
			}
			for _, bad := range opts.invalidIgnores {
				ctx.ReportRange(reportRange, rule.RuleMessage{
					Id: "invalidIgnorePattern",
					Description: fmt.Sprintf(
						"Invalid regular expression in `ignore` option: `%s`: %s",
						bad.pattern, bad.err.Error(),
					),
				})
			}
			return nil
		}

		for _, segment := range pathSegments {
			for _, re := range opts.ignores {
				if re.TestOrTimeout(segment) {
					return nil
				}
			}
		}
		cases := opts.selectedCases()
		if opts.checkDirectories {
			for _, directory := range pathSegments[:len(pathSegments)-1] {
				if strings.HasPrefix(directory, "$") {
					continue
				}
				leading, words := splitFilename(directory)
				valid, invalidWord, invalidCandidates := validateFilename(words, cases)
				if valid {
					continue
				}
				if ctx.DisableManager.IsRuleDisabled(filenameCaseRuleName, reportRange.Pos()) {
					return nil
				}
				renamed := fixFilename(words, cases, invalidWord, invalidCandidates, leading, "")
				ctx.ReportRange(reportRange, rule.RuleMessage{
					Id: messageIDDirectoryCase,
					Description: "Directory name `" + directory + "` is not in " +
						englishishCaseNames(cases) + ". Rename it to " + englishishBacktickJoin(renamed) + ".",
				})
				return nil
			}
		}
		if isIgnoredByDefault(basename) {
			return nil
		}

		ext := nodeExtname(basename)
		filename := basename[:len(basename)-len(ext)]
		middle := ""
		if opts.multipleFileExtensions {
			if i := strings.IndexByte(filename, '.'); i >= 0 {
				middle = filename[i:]
				filename = filename[:i]
			}
		}

		leading, words := splitFilename(filename)
		valid, invalidWord, invalidCandidates := validateFilename(words, cases)
		if strings.HasPrefix(filename, "$") {
			valid = true
		}
		lowerExt := ecmascript.StringToLowerCase(ext)
		if valid {
			if ext != lowerExt {
				if ctx.DisableManager.IsRuleDisabled(filenameCaseRuleName, reportRange.Pos()) {
					return nil
				}
				ctx.ReportRange(reportRange, rule.RuleMessage{
					Id: messageIDFilenameExtension,
					Description: "File extension `" + ext + "` is not in lowercase. Rename it to `" +
						filename + middle + lowerExt + "`.",
				})
			}
			return nil
		}

		// A filename diagnostic has no optional fix/suggestion artifact for the
		// deferred reporting APIs to gate. Check suppression here, after cheap
		// validation but before building the potentially combinatorial rename
		// list and message; ReportRange repeats only the cached lookup.
		if ctx.DisableManager.IsRuleDisabled(filenameCaseRuleName, reportRange.Pos()) {
			return nil
		}

		renamed := fixFilename(words, cases, invalidWord, invalidCandidates, leading, middle+lowerExt)
		ctx.ReportRange(reportRange, rule.RuleMessage{
			Id: messageIDFilenameCase,
			Description: "Filename is not in " + englishishCaseNames(cases) + ". Rename it to " +
				englishishBacktickJoin(renamed) + ".",
		})
		return nil
	},
}
