package utils

import "github.com/web-infra-dev/rslint/internal/utils/ecmascript"

// IsValidRegexLiteral reports whether literal is a complete ECMAScript RegExp
// literal, including leading/trailing slashes and flags, under tsgo's latest
// regex grammar. Use this before offering a fix that emits a regex literal.
func IsValidRegexLiteral(literal string) bool {
	return ecmascript.IsValidRegexLiteral(literal)
}
