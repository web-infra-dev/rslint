// This file is the port of is-extglob 2.1.1, the package is-glob hands the
// one question its own scan does not answer: does the string carry an extended
// glob list? is-glob asks it first and returns early when the answer is yes,
// so the two are separate here for the same reason they are separate upstream.

package isglob

import "regexp"

// extglobPattern is is-extglob's own regexp, kept verbatim. The first group
// swallows an escaped character so that `\@(a)` does not read as a list.
var extglobPattern = regexp.MustCompile(`(\\).|([@?!+*]\(.*\))`)

// IsExtglob reports whether str carries an extended glob list — `!(a)`,
// `@(a|b)`, `+(a)`, `?(a)` or `*(a)`.
func IsExtglob(str string) bool {
	for str != "" {
		match := extglobPattern.FindStringSubmatchIndex(str)
		if match == nil {
			return false
		}
		if match[4] != -1 {
			return true
		}
		str = str[match[1]:]
	}
	return false
}
