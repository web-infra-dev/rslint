package filename_case

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/dlclark/regexp2"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// caseStyle is one of the four supported filename case styles.
type caseStyle struct {
	key  string
	name string
	fn   func(string) string
}

// allCases keeps a stable canonical iteration order for the `cases` option.
//
// NOTE: Unlike ESLint, where Object.keys() over `options.cases` preserves the
// user's literal-property insertion order, we always use this canonical order
// in the diagnostic message. The reason is that rslint receives options as a
// `map[string]interface{}` after JSON parsing; Go map iteration is not
// order-preserving, and the original key order is unrecoverable. Locking the
// order here keeps message text deterministic.
var allCases = []caseStyle{
	{key: "camelCase", name: "camel case", fn: toCamelCase},
	{key: "snakeCase", name: "snake case", fn: toSnakeCase},
	{key: "kebabCase", name: "kebab case", fn: toKebabCase},
	{key: "pascalCase", name: "pascal case", fn: toPascalCase},
}

var caseByKey = func() map[string]caseStyle {
	m := make(map[string]caseStyle, len(allCases))
	for _, c := range allCases {
		m[c.key] = c
	}
	return m
}()

// ignoredByDefault mirrors the upstream's hardcoded set of files that cannot
// change case (notably required by Node / build tooling).
var ignoredByDefault = map[string]bool{
	"index.js":  true,
	"index.mjs": true,
	"index.cjs": true,
	"index.ts":  true,
	"index.tsx": true,
	"index.vue": true,
}

const reOpts = regexp2.ECMAScript | regexp2.Unicode

// invalidIgnore captures a single user-supplied `ignore` pattern that failed
// to compile. The rule reports each one as its own diagnostic so the user can
// see which configuration entry is broken.
type invalidIgnore struct {
	pattern string
	err     error
}

// Options is the parsed shape of the user's rule configuration.
type Options struct {
	Cases                  []caseStyle
	Ignores                []*regexp2.Regexp
	InvalidIgnores         []invalidIgnore
	MultipleFileExtensions bool
}

func parseOptions(rawOpts any) Options {
	opts := Options{
		Cases:                  []caseStyle{caseByKey["kebabCase"]},
		MultipleFileExtensions: true,
	}
	optsMap := utils.GetOptionsMap(rawOpts)
	if optsMap == nil {
		return opts
	}

	if v, ok := optsMap["case"].(string); ok {
		if c, found := caseByKey[v]; found {
			opts.Cases = []caseStyle{c}
		}
	} else if casesMap, ok := optsMap["cases"].(map[string]interface{}); ok {
		var chosen []caseStyle
		for _, c := range allCases {
			if b, ok := casesMap[c.key].(bool); ok && b {
				chosen = append(chosen, c)
			}
		}
		if len(chosen) > 0 {
			opts.Cases = chosen
		}
	}

	if v, ok := optsMap["ignore"].([]interface{}); ok {
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				continue
			}
			if re, err := regexp2.Compile(s, reOpts); err == nil {
				opts.Ignores = append(opts.Ignores, re)
			} else {
				opts.InvalidIgnores = append(opts.InvalidIgnores, invalidIgnore{pattern: s, err: err})
			}
		}
	}

	if v, ok := optsMap["multipleFileExtensions"].(bool); ok {
		opts.MultipleFileExtensions = v
	}

	return opts
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
//   1. /([\p{Ll}\d])(\p{Lu})/gu       → insert delimiter before an uppercase
//                                       that follows a lowercase or digit.
//   2. /(\p{Lu})([\p{Lu}][\p{Ll}])/gu → insert delimiter between two
//                                       uppercases when the second is the
//                                       start of a TitleCase word
//                                       (e.g. `XMLHttp` → `XML Http`).
//   3. /[^\p{L}\d]+/giu               → collapse non-alphanumeric runs into
//                                       the same delimiter.
//
// Then trim leading/trailing delimiters and split. We use rune `0` as the
// internal delimiter (matching the upstream's `\0`).
func splitWords(s string) []string {
	runes := []rune(s)

	// Pass 1: lower/digit + upper → split.
	var pass1 []rune
	for i, r := range runes {
		if i > 0 {
			prev := runes[i-1]
			if (unicode.IsLower(prev) || isASCIIDigit(prev)) && unicode.IsUpper(r) {
				pass1 = append(pass1, 0)
			}
		}
		pass1 = append(pass1, r)
	}

	// Pass 2: upper + (upper + lower) → split between the first and second
	// uppercase.
	var pass2 []rune
	for i, r := range pass1 {
		if i > 0 && i < len(pass1)-1 {
			prev, next := pass1[i-1], pass1[i+1]
			if unicode.IsUpper(prev) && unicode.IsUpper(r) && unicode.IsLower(next) {
				pass2 = append(pass2, 0)
			}
		}
		pass2 = append(pass2, r)
	}

	// Pass 3: collapse runs of non-alphanumeric into one delimiter.
	var stripped []rune
	inDelim := false
	for _, r := range pass2 {
		if !unicode.IsLetter(r) && !isASCIIDigit(r) {
			if !inDelim {
				stripped = append(stripped, 0)
				inDelim = true
			}
			continue
		}
		stripped = append(stripped, r)
		inDelim = false
	}

	// Trim leading and trailing delimiters.
	start, end := 0, len(stripped)
	for start < end && stripped[start] == 0 {
		start++
	}
	for end > start && stripped[end-1] == 0 {
		end--
	}
	if start >= end {
		return nil
	}

	var words []string
	var cur []rune
	for _, r := range stripped[start:end] {
		if r == 0 {
			if len(cur) > 0 {
				words = append(words, string(cur))
				cur = cur[:0]
			}
			continue
		}
		cur = append(cur, r)
	}
	if len(cur) > 0 {
		words = append(words, string(cur))
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
	runes := []rune(word)
	char0 := runes[0]
	rest := strings.ToLower(string(runes[1:]))
	if index > 0 && isASCIIDigit(char0) {
		return "_" + string(char0) + rest
	}
	return strings.ToUpper(string(char0)) + rest
}

func toCamelCase(s string) string {
	words := splitWords(s)
	if len(words) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(strings.ToLower(words[0]))
	for i := 1; i < len(words); i++ {
		sb.WriteString(pascalLikeTransform(words[i], i))
	}
	return sb.String()
}

func toPascalCase(s string) string {
	words := splitWords(s)
	if len(words) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, w := range words {
		sb.WriteString(pascalLikeTransform(w, i))
	}
	return sb.String()
}

func toKebabCase(s string) string { return joinNoCase(splitWords(s), "-") }
func toSnakeCase(s string) string { return joinNoCase(splitWords(s), "_") }

func joinNoCase(words []string, delim string) string {
	if len(words) == 0 {
		return ""
	}
	parts := make([]string, len(words))
	for i, w := range words {
		parts[i] = strings.ToLower(w)
	}
	return strings.Join(parts, delim)
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

	var hasLast, lastIgnored bool
	var lastWord []rune
	flush := func() {
		if !hasLast {
			return
		}
		words = append(words, filenameWord{word: string(lastWord), ignored: lastIgnored})
		lastWord = lastWord[:0]
		hasLast = false
	}
	for _, r := range tailing {
		ig := isIgnoredChar(r)
		if hasLast && lastIgnored == ig {
			lastWord = append(lastWord, r)
			continue
		}
		flush()
		lastWord = append(lastWord[:0], r)
		lastIgnored = ig
		hasLast = true
	}
	flush()
	return leading, words
}

// validateFilename returns true when every non-ignored chunk already matches
// at least one of the chosen case styles.
func validateFilename(words []filenameWord, cases []caseStyle) bool {
	for _, w := range words {
		if w.ignored {
			continue
		}
		ok := false
		for _, c := range cases {
			if c.fn(w.word) == w.word {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	return true
}

// fixFilename builds the deduplicated, ordered list of suggested filenames by
// taking the cartesian product of each non-ignored chunk's case-conversion
// candidates. The order matches change-case's left-to-right output so the
// message reads `fooBar`, `FooBar`, `foo-bar` for cases ordered camel,
// pascal, kebab.
func fixFilename(words []filenameWord, cases []caseStyle, leading, trailing string) []string {
	replacements := make([][]string, len(words))
	for i, w := range words {
		if w.ignored {
			replacements[i] = []string{w.word}
			continue
		}
		cand := make([]string, len(cases))
		for j, c := range cases {
			cand[j] = c.fn(w.word)
		}
		replacements[i] = cand
	}

	seen := map[string]bool{}
	var out []string
	var visit func(idx int, acc string)
	visit = func(idx int, acc string) {
		if idx == len(replacements) {
			full := leading + acc + trailing
			if !seen[full] {
				seen[full] = true
				out = append(out, full)
			}
			return
		}
		for _, item := range replacements[idx] {
			visit(idx+1, acc+item)
		}
	}
	visit(0, "")
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

func backtickList(items []string) []string {
	out := make([]string, len(items))
	for i, s := range items {
		out[i] = "`" + s + "`"
	}
	return out
}

func isLowerCase(s string) bool { return s == strings.ToLower(s) }

var FilenameCaseRule = rule.Rule{
	Name: "unicorn/filename-case",
	// The rule is purely filename-driven — it does not inspect any AST node.
	// `Run` is invoked once per source file, so we do the work here and
	// return an empty listener map. (The linter's visitor walks SourceFile
	// children but never the SourceFile node itself, so a
	// `ast.KindSourceFile` listener would silently never fire.)
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		if ctx.SourceFile == nil {
			return rule.RuleListeners{}
		}
		fileName := ctx.SourceFile.FileName()
		if fileName == "" {
			return rule.RuleListeners{}
		}
		// `tspath.GetBaseFileName` normalizes `\` → `/` first, so a Windows
		// path like `src\foo\bar.js` resolves the basename correctly.
		basename := tspath.GetBaseFileName(fileName)
		// Skip ESLint's stdin / inline-source virtual filenames.
		if basename == "<input>" || basename == "<text>" {
			return rule.RuleListeners{}
		}

		opts := parseOptions(options)

		// Configuration error: any malformed `ignore` pattern aborts
		// case-checking on this file. Mirrors ESLint's behaviour, where
		// `new RegExp(item, 'u')` throws at rule-create time and the rule
		// produces no further diagnostics until the config is fixed —
		// returning case reports based on a partially-broken ignore list
		// would be misleading.
		if len(opts.InvalidIgnores) > 0 {
			for _, bad := range opts.InvalidIgnores {
				ctx.ReportRange(core.NewTextRange(0, 0), rule.RuleMessage{
					Id: "invalidIgnorePattern",
					Description: fmt.Sprintf(
						"Invalid regular expression in `ignore` option: `%s`: %s",
						bad.pattern, bad.err.Error(),
					),
				})
			}
			return rule.RuleListeners{}
		}

		if ignoredByDefault[basename] {
			return rule.RuleListeners{}
		}
		for _, re := range opts.Ignores {
			if matched, _ := re.MatchString(basename); matched {
				return rule.RuleListeners{}
			}
		}

		ext := nodeExtname(basename)
		filename := strings.TrimSuffix(basename, ext)
		middle := ""
		if opts.MultipleFileExtensions {
			if i := strings.IndexByte(filename, '.'); i >= 0 {
				middle = filename[i:]
				filename = filename[:i]
			}
		}

		leading, words := splitFilename(filename)
		if validateFilename(words, opts.Cases) {
			if !isLowerCase(ext) {
				ctx.ReportRange(core.NewTextRange(0, 0), rule.RuleMessage{
					Id: "filenameExtension",
					Description: fmt.Sprintf(
						"File extension `%s` is not in lowercase. Rename it to `%s`.",
						ext, filename+middle+strings.ToLower(ext),
					),
				})
			}
			return rule.RuleListeners{}
		}

		renamed := fixFilename(words, opts.Cases, leading, middle+strings.ToLower(ext))
		caseNames := make([]string, len(opts.Cases))
		for i, c := range opts.Cases {
			caseNames[i] = c.name
		}
		ctx.ReportRange(core.NewTextRange(0, 0), rule.RuleMessage{
			Id: "filenameCase",
			Description: fmt.Sprintf(
				"Filename is not in %s. Rename it to %s.",
				englishishJoin(caseNames),
				englishishJoin(backtickList(renamed)),
			),
		})
		return rule.RuleListeners{}
	},
}
