package no_useless_undefined

import (
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

func reportExpression(ctx rule.RuleContext, node, expression *ast.Node, replacement string) {
	if !unicornutil.IsGlobalReference(ctx, node) {
		return
	}
	ctx.ReportNodeWithDeferredFixes(node, uselessUndefined, func() []rule.RuleFix {
		source := ctx.SourceFile
		text := source.Text()
		replace := func(start, end int, value string) rule.RuleFix {
			before := ecmascript.SkipTrailingWhitespace(text, 0, start)
			// Keep the first line break when it immediately follows a token or comment.
			space := text[before:start]
			switch {
			case strings.HasPrefix(space, "\r\n"):
				value = "\r\n" + value
			case strings.HasPrefix(space, "\r"), strings.HasPrefix(space, "\n"):
				value = space[:1] + value
			}
			return rule.RuleFixReplaceRange(core.NewTextRange(before, end), value)
		}

		var fixes []rule.RuleFix
		for current := expression; current != node; {
			if ast.IsParenthesizedExpression(current) {
				span := utils.TrimNodeTextRange(source, current)
				if span.Len() < 2 || text[span.End()-1] != ')' {
					return nil
				}
				// Edit the parentheses individually so comments within them survive.
				fixes = append(fixes, replace(span.Pos(), span.Pos()+1, ""), replace(span.End()-1, span.End(), ""))
				current = current.Expression()
			} else {
				current = utils.JSDocTypeCastExpression(current)
			}
		}
		span := utils.TrimNodeTextRange(source, node)
		return append(fixes, replace(span.Pos(), span.End(), replacement))
	})
}

func reportInitializer(ctx rule.RuleContext, node, left, value *ast.Node, isTypeScript bool) {
	if left == nil || utils.ESTreeRuntimeExpression(value) != node || !unicornutil.IsGlobalReference(ctx, node) {
		return
	}
	ctx.ReportNodeWithDeferredFixes(node, uselessUndefined, func() []rule.RuleFix {
		source := ctx.SourceFile
		start := utils.GetESTreeBindingIdentifierRange(source, left).End()
		optionalAt := -1
		if parameter := left.Parent; parameter != nil && ast.IsParameterDeclaration(parameter) && parameter.Name() == left {
			annotation := parameter.Type()
			typed := isTypeScript || annotation != nil
			if typed && ast.IsBindingPattern(left) {
				// TypeScript rejects an optional binding pattern in an implementation.
				return nil
			}
			if typed && parameter.QuestionToken() == nil && utils.IsFunctionLikeContainer(parameter.Parent) {
				if annotation == nil || !canMakeParameterOptional(parameter) {
					return nil
				}
				colon, ok := utils.TokenBeforePosition(source, utils.TrimNodeTextRange(source, annotation).Pos())
				if !ok || colon.Kind != ast.KindColonToken {
					return nil
				}
				optionalAt = colon.Start
			}
		}
		if utils.HasCommentInSpan(ctx.Comments.All(), start, value.End()) {
			return nil
		}
		fixes := []rule.RuleFix{rule.RuleFixRemoveRange(core.NewTextRange(start, value.End()))}
		if optionalAt >= 0 {
			fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(optionalAt, optionalAt), "?"))
		}
		return fixes
	})
}

func canMakeParameterOptional(parameter *ast.Node) bool {
	// A parameter property's optional marker would also change the class shape.
	if ast.IsParameterPropertyDeclaration(parameter, parameter.Parent) {
		return false
	}
	parameters := parameter.Parent.Parameters()
	for _, next := range parameters[slices.Index(parameters, parameter)+1:] {
		declaration := next.AsParameterDeclaration()
		if declaration.Initializer == nil && declaration.QuestionToken == nil && declaration.DotDotDotToken == nil {
			return false
		}
	}
	return true
}

func lastArgumentFix(ctx rule.RuleContext, argument *ast.Node, onlyArgument bool) []rule.RuleFix {
	span := utils.TrimNodeTextRange(ctx.SourceFile, argument)
	start, end := span.Pos(), span.End()
	if onlyArgument {
		if comma, ok := utils.TokenAtOrAfter(ctx.SourceFile, end); ok && comma.Kind == ast.KindCommaToken {
			end = comma.End
		}
	} else {
		comma, ok := utils.TokenBeforePosition(ctx.SourceFile, start)
		if !ok || comma.Kind != ast.KindCommaToken {
			return nil
		}
		// Keep a trailing comma for the previous argument when one is present.
		start = comma.Start
	}
	if utils.HasCommentInSpan(ctx.Comments.All(), start, end) {
		return nil
	}
	return []rule.RuleFix{rule.RuleFixRemoveRange(core.NewTextRange(start, end))}
}
