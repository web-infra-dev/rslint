package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslintUtils "github.com/web-infra-dev/rslint/internal/utils"
)

// AccessorQuestionDotToken returns the optional-chain token owned by accessor.
func AccessorQuestionDotToken(accessor *ast.Node) *ast.Node {
	if accessor == nil {
		return nil
	}

	switch accessor.Kind {
	case ast.KindPropertyAccessExpression:
		return accessor.AsPropertyAccessExpression().QuestionDotToken
	case ast.KindElementAccessExpression:
		return accessor.AsElementAccessExpression().QuestionDotToken
	default:
		return nil
	}
}

func removeAccessorSyntax(
	ctx rule.RuleContext,
	accessor *ast.Node,
	start int,
) []rule.RuleFix {
	if !rslintUtils.HasCommentInSpan(ctx.Comments.All(), start, accessor.End()) {
		return []rule.RuleFix{
			rule.RuleFixRemoveRange(core.NewTextRange(start, accessor.End())),
		}
	}

	fixes := []rule.RuleFix{}
	for _, token := range rslintUtils.TokensOfNode(ctx.SourceFile, accessor) {
		if token.Start >= start && token.End <= accessor.End() {
			fixes = append(fixes, rule.RuleFixRemoveRange(token.Range()))
		}
	}
	return fixes
}

// RemoveMemberAccessorFixes removes one member from a parsed Jest chain while
// preserving comments and the chain's optional boundary.
func RemoveMemberAccessorFixes(
	ctx rule.RuleContext,
	entries []ParsedJestFnMemberEntry,
	index int,
) ([]rule.RuleFix, bool) {
	if index < 0 || index >= len(entries) {
		return nil, false
	}

	entry := &entries[index]
	receiver, accessor := GetAccessorReceiverAndParent(entry)
	if receiver == nil || accessor == nil {
		return nil, false
	}

	questionDot := AccessorQuestionDotToken(accessor)
	start := receiver.End()
	if questionDot != nil {
		start = questionDot.End()
	}
	fixes := removeAccessorSyntax(ctx, accessor, start)
	if len(fixes) == 0 {
		return nil, false
	}

	if questionDot == nil {
		return fixes, true
	}

	// The retained `?.` must directly introduce the next member or call.
	// Remove a duplicate connector from that operation, if present.
	if index+1 < len(entries) {
		nextEntry := &entries[index+1]
		nextReceiver, nextAccessor := GetAccessorReceiverAndParent(nextEntry)
		if nextReceiver != accessor {
			return nil, false
		}

		switch nextAccessor.Kind {
		case ast.KindPropertyAccessExpression:
			connector, ok := rslintUtils.TokenAtOrAfter(ctx.SourceFile, accessor.End())
			if !ok || (connector.Text != "." && connector.Text != "?.") {
				return nil, false
			}
			nextNameStart := rslintUtils.TrimNodeTextRange(ctx.SourceFile, nextEntry.Node).Pos()
			if rslintUtils.HasCommentInSpan(ctx.Comments.All(), accessor.End(), nextNameStart) {
				fixes = append(fixes, rule.RuleFixRemoveRange(connector.Range()))
			} else {
				fixes = append(fixes, rule.RuleFixRemoveRange(
					core.NewTextRange(accessor.End(), nextNameStart),
				))
			}
		case ast.KindElementAccessExpression:
			if nextQuestionDot := AccessorQuestionDotToken(nextAccessor); nextQuestionDot != nil {
				nextQuestionDotRange := rslintUtils.TrimNodeTextRange(ctx.SourceFile, nextQuestionDot)
				if rslintUtils.HasCommentInSpan(ctx.Comments.All(), accessor.End(), nextQuestionDotRange.End()) {
					fixes = append(fixes, rule.RuleFixRemoveRange(nextQuestionDotRange))
				} else {
					fixes = append(fixes, rule.RuleFixRemoveRange(
						core.NewTextRange(accessor.End(), nextQuestionDotRange.End()),
					))
				}
			}
		default:
			return nil, false
		}
		return fixes, true
	}

	if entry.Call == nil {
		return nil, false
	}
	callExpr := entry.Call.AsCallExpression()
	if callExpr == nil || callExpr.Expression != accessor {
		return nil, false
	}
	if callExpr.QuestionDotToken != nil {
		callQuestionDotRange := rslintUtils.TrimNodeTextRange(ctx.SourceFile, callExpr.QuestionDotToken)
		if rslintUtils.HasCommentInSpan(ctx.Comments.All(), accessor.End(), callQuestionDotRange.End()) {
			fixes = append(fixes, rule.RuleFixRemoveRange(callQuestionDotRange))
		} else {
			fixes = append(fixes, rule.RuleFixRemoveRange(
				core.NewTextRange(accessor.End(), callQuestionDotRange.End()),
			))
		}
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
func ReplaceCallSuffixFix(
	callNode *ast.Node,
	callee *ast.Node,
	replacement string,
) (rule.RuleFix, bool) {
	callExpr := callNode.AsCallExpression()
	if callExpr == nil || callee == nil {
		return rule.RuleFix{}, false
	}

	start := callee.End()
	if callExpr.QuestionDotToken != nil {
		start = callExpr.QuestionDotToken.End()
	}
	return rule.RuleFixReplaceRange(
		core.NewTextRange(start, callNode.End()),
		replacement,
	), true
}
