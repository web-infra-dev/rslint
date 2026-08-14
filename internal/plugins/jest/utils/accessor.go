package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

// AccessorQuestionDotToken returns the optional-chain token owned by accessor.
func AccessorQuestionDotToken(accessor *ast.Node) *ast.Node {
	return testFramework.AccessorQuestionDotToken(accessor)
}

// RemoveMemberAccessorFixes removes one member from a parsed Jest chain while
// preserving comments and the chain's optional boundary.
//
// The removal itself is framework-neutral, so it lives in test_framework and is
// shared with the Rstest plugin. That version derives the operation following
// the removed accessor from the AST rather than from the next element of the
// parsed chain, which is both what an alias-resolved entry needs — its
// neighbors are not in that chain at all — and one fewer way for two copies of
// this logic to drift apart. It is also why this takes the entry itself: the
// surrounding slice and an index into it are no longer needed.
func RemoveMemberAccessorFixes(
	ctx rule.RuleContext,
	entry *ParsedJestFnMemberEntry,
) ([]rule.RuleFix, bool) {
	ranges, ok := testFramework.RemoveAccessorEntryRanges(
		ctx.SourceFile,
		ctx.Comments.All(),
		entry,
	)
	if !ok {
		return nil, false
	}

	fixes := make([]rule.RuleFix, 0, len(ranges))
	for _, textRange := range ranges {
		fixes = append(fixes, rule.RuleFixRemoveRange(textRange))
	}
	return fixes, true
}

// ReplaceMemberNameFix renames an accessor without changing its syntax,
// optional-chain token, surrounding parentheses, or comments.
func ReplaceMemberNameFix(
	ctx rule.RuleContext,
	entry *ParsedJestFnMemberEntry,
	name string,
) (rule.RuleFix, bool) {
	_, accessor := GetAccessorReceiverAndParent(entry)
	if accessor == nil {
		return rule.RuleFix{}, false
	}

	switch accessor.Kind {
	case ast.KindPropertyAccessExpression:
		return rule.RuleFixReplace(ctx.SourceFile, entry.Node, name), true
	case ast.KindElementAccessExpression:
		return rule.RuleFixReplace(ctx.SourceFile, entry.Node, "'"+name+"'"), true
	default:
		return rule.RuleFix{}, false
	}
}

// InsertMemberBeforeAccessorFix inserts a property before entry while keeping
// an optional boundary on the original receiver.
func InsertMemberBeforeAccessorFix(
	ctx rule.RuleContext,
	entry *ParsedJestFnMemberEntry,
	name string,
) (rule.RuleFix, bool) {
	receiver, accessor := GetAccessorReceiverAndParent(entry)
	if receiver == nil || accessor == nil {
		return rule.RuleFix{}, false
	}

	questionDot := AccessorQuestionDotToken(accessor)
	if questionDot == nil {
		return rule.RuleFixReplaceRange(
			core.NewTextRange(receiver.End(), receiver.End()),
			"."+name,
		), true
	}

	suffix := name
	if accessor.Kind == ast.KindPropertyAccessExpression {
		suffix += "."
	}
	return rule.RuleFixReplaceRange(
		core.NewTextRange(questionDot.End(), questionDot.End()),
		suffix,
	), true
}

// ReplaceCallSuffixFix replaces a call's type arguments and arguments while
// preserving an optional-call token owned by the call.
//
// The range is derived from the call node alone: it starts at the first token
// after the callee (or after the call's optional token), so a comment sitting
// between the callee and the type/argument list is preserved rather than
// swallowed.
func ReplaceCallSuffixFix(
	sourceFile *ast.SourceFile,
	callNode *ast.Node,
	replacement string,
) (rule.RuleFix, bool) {
	callExpr := callNode.AsCallExpression()
	if callExpr == nil {
		return rule.RuleFix{}, false
	}

	start := callExpr.Expression.End()
	if callExpr.QuestionDotToken != nil {
		start = callExpr.QuestionDotToken.End()
	}
	// Advance past trivia to the `<` or `(` that actually opens the suffix.
	start = scanner.GetRangeOfTokenAtPosition(sourceFile, start).Pos()
	return rule.RuleFixReplaceRange(
		core.NewTextRange(start, callNode.End()),
		replacement,
	), true
}
