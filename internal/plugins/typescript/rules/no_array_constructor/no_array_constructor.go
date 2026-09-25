package no_array_constructor

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func useLiteralMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useLiteral",
		Description: "The array literal notation [] is preferable.",
	}
}

func sourceMayUseArrayConstructor(sourceFile *ast.SourceFile) bool {
	// Stay conservative for direct callers without a parsed source file.
	if sourceFile == nil || sourceFile.AsNode().Kind != ast.KindSourceFile {
		return true
	}
	text := sourceFile.Text()
	// Leave Unicode decoding to the compiler's normalized-name cache. Check
	// escapes first so repeated escaped names do not cause a failed substring
	// search for the literal spelling on every run.
	if strings.Contains(text, `\u`) {
		return sourceFile.HasIdentifier("Array")
	}
	// Plain spellings need no AST-wide name index. Comments, strings and longer
	// names can keep the listeners, which still check the actual callee.
	return strings.Contains(text, "Array")
}

func buildArrayConstructorFixes(
	sourceText string,
	node *ast.Node,
	args *ast.NodeList,
	reportRange core.TextRange,
) []rule.RuleFix {
	if args == nil {
		return []rule.RuleFix{rule.RuleFixReplaceRange(reportRange, "[]")}
	}

	closeParenPos := node.End() - 1
	if args.Pos() < reportRange.Pos() ||
		closeParenPos < args.Pos() ||
		closeParenPos >= len(sourceText) ||
		sourceText[closeParenPos] != ')' {
		// Keep malformed/recovery ASTs safe and preserve the rule's previous
		// fallback for a missing closing parenthesis.
		return []rule.RuleFix{rule.RuleFixReplaceRange(reportRange, "[]")}
	}

	// Replace only the call boundaries. The argument text, including comments,
	// whitespace, and trailing commas, stays in place without another scan or a
	// copy into a replacement string.
	return []rule.RuleFix{
		rule.RuleFixReplaceRange(core.NewTextRange(reportRange.Pos(), args.Pos()), "["),
		rule.RuleFixReplaceRange(core.NewTextRange(closeParenPos, node.End()), "]"),
	}
}

// Keep listener construction outside Run so the context captured by check
// does not escape on files rejected by the identifier fast path.
func noArrayConstructorListeners(ctx rule.RuleContext) rule.RuleListeners {
	check := func(node *ast.Node) {
		var callee *ast.Node
		var args *ast.NodeList
		var typeArgs *ast.NodeList

		switch node.Kind {
		case ast.KindCallExpression:
			callExpr := node.AsCallExpression()
			callee = callExpr.Expression
			args = callExpr.Arguments
			typeArgs = callExpr.TypeArguments
		case ast.KindNewExpression:
			newExpr := node.AsNewExpression()
			callee = newExpr.Expression
			args = newExpr.Arguments
			typeArgs = newExpr.TypeArguments
		default:
			return
		}

		// These exceptions do not depend on the callee. Check them before
		// unwrapping parentheses or inspecting the identifier, including the
		// upstream exception for a single spread argument.
		if (args != nil && len(args.Nodes) == 1) ||
			(typeArgs != nil && len(typeArgs.Nodes) > 0) {
			return
		}

		// ESTree does not expose grouping parentheses around a callee. Match
		// that behavior without unwrapping TypeScript-only outer expressions
		// such as non-null or `as` expressions.
		if callee == nil {
			return
		}
		if callee.Kind == ast.KindParenthesizedExpression {
			callee = ast.SkipParentheses(callee)
		}

		// Check if callee is an Identifier named "Array"
		if callee.Kind != ast.KindIdentifier {
			return
		}
		identifier := callee.AsIdentifier()
		if identifier.Text != "Array" {
			return
		}

		sourceText := ctx.SourceFile.Text()
		reportRange := core.NewTextRange(scanner.SkipTrivia(sourceText, node.Pos()), node.End())
		ctx.ReportRangeWithDeferredFixes(reportRange, useLiteralMessage(), func() []rule.RuleFix {
			return buildArrayConstructorFixes(sourceText, node, args, reportRange)
		})
	}

	return rule.RuleListeners{
		ast.KindCallExpression: check,
		ast.KindNewExpression:  check,
	}
}

var NoArrayConstructorRule = rule.CreateRule(rule.Rule{
	Name:   "no-array-constructor",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if !sourceMayUseArrayConstructor(ctx.SourceFile) {
			return nil
		}
		return noArrayConstructorListeners(ctx)
	},
})
