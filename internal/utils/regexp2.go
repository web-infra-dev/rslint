package utils

import "github.com/dlclark/regexp2"

const (
	// JSRegexOptions enables regexp2's ECMAScript mode for JavaScript RegExp
	// patterns that do not pass explicit flags.
	JSRegexOptions regexp2.RegexOptions = regexp2.ECMAScript
	// JSUnicodeRegexOptions mirrors `new RegExp(pattern, "u")`.
	JSUnicodeRegexOptions regexp2.RegexOptions = regexp2.ECMAScript | regexp2.Unicode
)

// CompileRegexp2 compiles a regexp2 pattern with the caller's exact options.
// Use JSRegexOptions / JSUnicodeRegexOptions for ESLint rule options that model
// JavaScript RegExp patterns.
func CompileRegexp2(pattern string, options regexp2.RegexOptions) (*regexp2.Regexp, error) {
	return regexp2.Compile(pattern, options)
}

// Regexp2MatchString reports whether re matches s. regexp2 only returns an
// error for runtime failures such as timeouts, which lint rules treat as no
// match instead of reporting a diagnostic.
func Regexp2MatchString(re *regexp2.Regexp, s string) bool {
	if re == nil {
		return false
	}
	matched, err := re.MatchString(s)
	return err == nil && matched
}
