package utils

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/diagnostics"
	"github.com/microsoft/typescript-go/shim/scanner"
)

// IsValidRegexLiteral reports whether literal is a complete ECMAScript RegExp
// literal, including leading/trailing slashes and flags, under tsgo's latest
// regex grammar. Use this before offering a fix that emits a regex literal.
func IsValidRegexLiteral(literal string) bool {
	_, flags := ExtractRegexPatternAndFlags(literal)
	s := scanner.NewScanner()
	hasError := false
	s.SetScriptTarget(core.ScriptTargetLatest)
	s.SetOnError(func(message *diagnostics.Message, _ int, _ int, _ ...any) {
		if !isIgnorableRegexLiteralDiagnostic(message, flags) {
			hasError = true
		}
	})
	s.SetText(literal)
	s.ResetTokenState(0)
	if s.Scan() != ast.KindSlashToken {
		return false
	}
	if s.ReScanSlashToken(true) != ast.KindRegularExpressionLiteral {
		return false
	}
	return !hasError && s.TokenFlags()&ast.TokenFlagsUnterminated == 0 && s.TokenEnd() == len(literal)
}

func isIgnorableRegexLiteralDiagnostic(message *diagnostics.Message, flags string) bool {
	if strings.ContainsAny(flags, "uv") {
		return false
	}
	return message == diagnostics.This_backreference_refers_to_a_group_that_does_not_exist_There_are_only_0_capturing_groups_in_this_regular_expression ||
		message == diagnostics.This_backreference_refers_to_a_group_that_does_not_exist_There_are_no_capturing_groups_in_this_regular_expression
}
