package reactutil

import (
	"regexp"
	"strings"
	"sync"
)

// GlobToRegex converts a minimatch-style glob into a fully anchored
// regular expression. Supports the subset of minimatch syntax the
// eslint-plugin-react ecosystem relies on:
//
//   - `*`             — any run of characters (`**` collapses to `*`)
//   - `?`             — a single character
//   - `[abc]`         — character class
//   - `[!abc]` / `[^abc]` — negated character class
//   - `{a,b,c}`       — brace expansion (nestable)
//   - `?(a|b)`        — zero or one of alternatives (extglob)
//   - `*(a|b)`        — zero or more of alternatives (extglob)
//   - `+(a|b)`        — one or more of alternatives (extglob)
//   - `@(a|b)`        — exactly one of alternatives (extglob)
//   - `!(a|b)`        — extglob negation (RE2 lacks lookarounds; approximated
//                       as "zero or one" — exact-match semantics not supported)
//   - `\X`            — literal X
//
// Leading `!` (whole-pattern negation) is intentionally NOT handled here:
// it inverts the whole-pattern match result and so cannot be expressed in
// a single anchored regex. Callers that need it should use `MatchGlob`,
// which special-cases the `!` prefix.
//
// Compilation is cached per-pattern; the returned `*regexp.Regexp` is
// safe to share across goroutines. Returns nil only on malformed `[...]`
// classes that would produce a regex RE2 rejects (callers treat nil as
// "exact-match-only fallback"); upstream minimatch never throws here, so
// nil should not arise for any real-world glob.
func GlobToRegex(pattern string) *regexp.Regexp {
	if v, ok := globToRegexCache.Load(pattern); ok {
		if re, ok := v.(*regexp.Regexp); ok {
			return re
		}
	}
	body := globBody([]rune(pattern))
	re, err := regexp.Compile("^" + body + "$")
	if err != nil {
		// Pattern was malformed enough to produce an invalid regex (e.g. a
		// `[...]` body the converter couldn't repair). Cache nil so we do
		// not retry on subsequent matches; callers fall back to "no match".
		globToRegexCache.Store(pattern, (*regexp.Regexp)(nil))
		return nil
	}
	globToRegexCache.Store(pattern, re)
	return re
}

// MatchGlob reports whether `text` matches the minimatch-style `pattern`.
// Returns false for empty `text`. Supports leading `!` whole-pattern
// negation: a pattern of `!X` matches everything except what `X` matches.
// `!!X` is treated as a literal pattern starting with `!` (matches `!X`),
// mirroring minimatch's odd-count-of-`!` rule.
func MatchGlob(text, pattern string) bool {
	if text == "" {
		return false
	}
	negate := false
	for strings.HasPrefix(pattern, "!") {
		negate = !negate
		pattern = pattern[1:]
	}
	re := GlobToRegex(pattern)
	if re == nil {
		return false
	}
	matched := re.MatchString(text)
	if negate {
		return !matched
	}
	return matched
}

var globToRegexCache sync.Map

// globBody recursively translates a glob fragment (already split on rune
// boundaries) into a regex body. Operates on `[]rune` so indices and
// slicing are codepoint-aligned — essential when patterns contain
// multi-byte characters (CJK, emoji); mixing rune indices with `string`
// byte offsets would misalign after the first multi-byte rune.
func globBody(runes []rune) string {
	var sb strings.Builder
	i := 0
	for i < len(runes) {
		r := runes[i]
		// Extglob: ?(...), *(...), +(...), @(...), !(...). Each is
		// `<sigil>(<alt>|<alt>|...)` — split on top-level `|` and
		// recursively convert each alternative.
		if i+1 < len(runes) && runes[i+1] == '(' && strings.ContainsRune("?*+@!", r) {
			if end, ok := findMatchingParen(runes, i+2); ok {
				alts := splitTopLevel(runes[i+2:end], '|', '(', ')')
				parts := make([]string, len(alts))
				for j, a := range alts {
					parts[j] = globBody(a)
				}
				body := strings.Join(parts, "|")
				switch r {
				case '?':
					sb.WriteString("(?:" + body + ")?")
				case '*':
					sb.WriteString("(?:" + body + ")*")
				case '+':
					sb.WriteString("(?:" + body + ")+")
				case '@':
					sb.WriteString("(?:" + body + ")")
				case '!':
					// RE2 lacks lookarounds; approximate as "zero or one"
					// so the pattern still compiles. Exact extglob `!(...)`
					// negation cannot be modeled in RE2.
					sb.WriteString("(?:" + body + ")?")
				}
				i = end + 1
				continue
			}
		}
		// Brace expansion: `{a,b,c}` (nestable). Split on top-level `,`
		// and recursively convert each branch.
		if r == '{' {
			if end, ok := findMatchingBrace(runes, i+1); ok {
				alts := splitTopLevel(runes[i+1:end], ',', '{', '}')
				// Single-branch braces (no `,`) are NOT brace expansion in
				// minimatch — they're treated as literal `{x}`. Mirror that.
				if len(alts) <= 1 {
					sb.WriteString(regexp.QuoteMeta("{"))
					i++
					continue
				}
				parts := make([]string, len(alts))
				for j, a := range alts {
					parts[j] = globBody(a)
				}
				sb.WriteString("(?:" + strings.Join(parts, "|") + ")")
				i = end + 1
				continue
			}
		}
		switch r {
		case '*':
			// Collapse `**` and longer runs to a single `.*`. Default
			// minimatch on this codebase runs with `noglobstar: true`
			// (see jsx_pascal_case docs); upstream eslint-plugin-react
			// matches that since component-name patterns never contain
			// path separators anyway.
			for i < len(runes) && runes[i] == '*' {
				i++
			}
			sb.WriteString(".*")
		case '?':
			sb.WriteString(".")
			i++
		case '[':
			// Character class. Find the matching `]` (don't escape inner
			// content beyond `!`/`^` negation, since the glob class syntax
			// is a strict subset of regex class syntax).
			closeIdx := -1
			for j := i + 1; j < len(runes); j++ {
				if runes[j] == ']' {
					closeIdx = j
					break
				}
			}
			// At least two chars between `[` and `]` so `[^x]` / `[!x]`
			// negation doesn't collapse to an empty inverted class (`[^]`),
			// which RE2 rejects.
			if closeIdx > i+1 {
				body := string(runes[i+1 : closeIdx])
				if len(body) > 1 && (body[0] == '!' || body[0] == '^') {
					body = "^" + body[1:]
				}
				sb.WriteString("[" + body + "]")
				i = closeIdx + 1
			} else {
				// Unbalanced `[`: treat as literal.
				sb.WriteString("\\[")
				i++
			}
		case '\\':
			// Escape the next character as a literal.
			if i+1 < len(runes) {
				sb.WriteString(regexp.QuoteMeta(string(runes[i+1])))
				i += 2
			} else {
				sb.WriteString("\\\\")
				i++
			}
		default:
			sb.WriteString(regexp.QuoteMeta(string(r)))
			i++
		}
	}
	return sb.String()
}

func findMatchingParen(runes []rune, start int) (int, bool) {
	depth := 1
	for j := start; j < len(runes); j++ {
		switch runes[j] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return j, true
			}
		}
	}
	return -1, false
}

func findMatchingBrace(runes []rune, start int) (int, bool) {
	depth := 1
	for j := start; j < len(runes); j++ {
		switch runes[j] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return j, true
			}
		}
	}
	return -1, false
}

// splitTopLevel splits `runes` on every `sep` that is at top level —
// ignoring separators inside paired `openCh`/`closeCh` delimiters. Used
// for brace alternatives (`,` outside nested `{...}`) and extglob
// alternatives (`|` outside nested `(...)`).
func splitTopLevel(runes []rune, sep, openCh, closeCh rune) [][]rune {
	var parts [][]rune
	depth := 0
	start := 0
	for i := range runes {
		switch runes[i] {
		case openCh:
			depth++
		case closeCh:
			depth--
		case sep:
			if depth == 0 {
				parts = append(parts, runes[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, runes[start:])
	return parts
}
