package reactutil

import (
	"regexp"
	"strings"
	"unicode"

	"unicode/utf8"
)

// hookNameRegex matches the React hook naming convention `^use[A-Z0-9].*$`.
// Mirrors eslint-plugin-react-hooks' `RE_HOOKS` and the `HOOK_REGEXP` used by
// no-unstable-nested-components. Exposed here because every React-flavored
// rule that needs to recognize hook calls re-derives the same predicate.
var hookNameRegex = regexp.MustCompile(`^use[A-Z0-9].*$`)

// IsHookName reports whether `name` matches the React hook naming convention.
// Returns false for empty input.
func IsHookName(name string) bool {
	if name == "" {
		return false
	}
	return hookNameRegex.MatchString(name)
}

// IsFirstLetterCapitalized is the exported alias for the package-private
// helper. Mirrors eslint-plugin-react's `lib/util/isFirstLetterCapitalized.js`
// — strips leading underscores then compares `unicode.ToUpper(r) == r`.
// Non-cased characters (CJK, digits, emoji) count as "capitalized" because
// they have no upper/lower mapping. Use this for any parent-name / binding
// capitalization check that needs to align with upstream's component
// detection semantics.
func IsFirstLetterCapitalized(s string) bool {
	return isFirstLetterCapitalized(s)
}

// IsLowercaseFirstLetter is the companion of IsFirstLetterCapitalized that
// matches upstream's exact lowercase-skip predicate from
// `lib/rules/no-unstable-nested-components.js`:
//
//	parentName[0] === parentName[0].toLowerCase()
//
// Notably this is NOT the negation of IsFirstLetterCapitalized: this
// helper does NOT strip leading underscores, so `_Foo` is treated as
// lowercase here (the `_` round-trips through `ToLower`) even though
// IsFirstLetterCapitalized returns true for `_Foo` (after stripping `_`,
// `Foo` is capitalized). Both helpers exist because upstream uses each
// in different code paths — keep them paired.
func IsLowercaseFirstLetter(s string) bool {
	if s == "" {
		return false
	}
	r, _ := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return false
	}
	return unicode.ToLower(r) == r
}

// IsCasedLowercaseFirstLetter mirrors upstream's
// `s[0] !== s[0].toUpperCase()` test (used by `forbid-component-props`'s
// componentName check and `forbid-dom-props`'s tag check): returns true iff
// the first rune is a cased letter currently in its lowercase form. Digits,
// `_`, `$`, and uppercase letters all return false. Distinct from
// IsLowercaseFirstLetter, which uses the looser `r === r.toLowerCase()`
// predicate (so `_Foo` returns true there, false here).
func IsCasedLowercaseFirstLetter(s string) bool {
	if s == "" {
		return false
	}
	r, _ := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return false
	}
	return unicode.ToLower(r) == r && unicode.ToUpper(r) != r
}

// isFirstLetterCapitalized mirrors eslint-plugin-react's helper of the same
// name (`lib/util/isFirstLetterCapitalized.js`). The semantics are:
//
//  1. Strip leading underscores: `_Foo` → "Foo" (so `_Foo` is treated the
//     same as `Foo`, matching upstream's `word.replace(/^_+/, ”)`).
//  2. A character is "capitalized" iff `unicode.ToUpper(r) == r` —
//     equivalent to upstream's `firstLetter.toUpperCase() === firstLetter`.
//
// Step 2 means non-cased characters (CJK, digits, emoji, symbols) all
// count as "capitalized" because they have no upper/lower mapping. This
// matters for non-ASCII identifiers like `function 不稳定组件()` — upstream
// classifies the function as a component (CJK char ≠ lowercase letter),
// and we must do the same to stay output-aligned.
func isFirstLetterCapitalized(s string) bool {
	stripped := strings.TrimLeft(s, "_")
	if stripped == "" {
		return false
	}
	r, _ := utf8.DecodeRuneInString(stripped)
	if r == utf8.RuneError {
		return false
	}
	return unicode.ToUpper(r) == r
}
