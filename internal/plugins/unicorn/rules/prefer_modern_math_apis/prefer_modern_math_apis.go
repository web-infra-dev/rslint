package prefer_modern_math_apis

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

const messageID = "prefer-modern-math-apis"

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-modern-math-apis.js
var PreferModernMathApisRule = rule.Rule{
	Name:   "unicorn/prefer-modern-math-apis",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) { checkLog(ctx, node) },
			ast.KindCallExpression:   func(node *ast.Node) { checkSqrt(ctx, node) },
		}
	},
}

func mathCall(node *ast.Node, method string) (unicornutil.DotMethodCall, bool) {
	one := 1
	call, ok := unicornutil.MatchDotMethodCall(utils.ESTreeRuntimeExpression(node), unicornutil.DotMethodCallOptions{
		Method: method, ArgumentsLength: &one, RejectSpreadElement: true,
	})
	return call, ok && isIdentifier(call.Object, "Math")
}

func isIdentifier(node *ast.Node, name string) bool {
	node = utils.ESTreeRuntimeExpression(node)
	return node != nil && ast.IsIdentifier(node) && node.Text() == name
}

func mathProperty(node *ast.Node) string {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || !ast.IsPropertyAccessExpression(node) || ast.IsOptionalChain(node) {
		return ""
	}
	member := node.AsPropertyAccessExpression()
	if !isIdentifier(member.Expression, "Math") || !ast.IsIdentifier(member.Name()) {
		return ""
	}
	return member.Name().Text()
}

func checkLog(ctx rule.RuleContext, node *ast.Node) {
	binary := node.AsBinaryExpression()
	operator := binary.OperatorToken.Kind
	if operator != ast.KindAsteriskToken && operator != ast.KindSlashToken {
		return
	}
	call, ok := mathCall(binary.Left, "log")
	constant := mathProperty(binary.Right)
	reversed := false
	if !ok && operator == ast.KindAsteriskToken {
		call, ok = mathCall(binary.Right, "log")
		constant = mathProperty(binary.Left)
		reversed = true
	}
	if !ok {
		return
	}
	method := ""
	if operator == ast.KindAsteriskToken {
		switch constant {
		case "LOG10E":
			method = "log10"
		case "LOG2E":
			method = "log2"
		}
	} else {
		switch constant {
		case "LN10":
			method = "log10"
		case "LN2":
			method = "log2"
		}
	}
	if method == "" {
		return
	}
	description := "Math.log(…) " + scanner.TokenToString(operator) + " Math." + constant
	if reversed {
		description = "Math." + constant + " * Math.log(…)"
	}
	argument := call.Call.Arguments()[0]
	autofix := isNumericValueUnused(node)
	if !autofix {
		if value, known := staticNumber(ctx, argument); known {
			autofix = value <= 0 || value == 1 || !isFinite(value)
		}
	}
	// A locally supplied Math object need not implement the builtin operations.
	autofix = autofix && ctx.Globals.Access("Math").IsDeclared() &&
		unicornutil.IsGlobalReference(ctx, utils.ESTreeRuntimeExpression(call.Object))
	report(ctx, node, method, description, autofix, func() []rule.RuleFix {
		nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
		argumentRange := utils.TrimNodeTextRange(ctx.SourceFile, argument)
		comments := ctx.Comments.All()
		if utils.HasCommentInSpan(comments, nodeRange.Pos(), argumentRange.Pos()) ||
			utils.HasCommentInSpan(comments, argumentRange.End(), nodeRange.End()) {
			return nil
		}
		return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node,
			"Math."+method+"("+utils.TrimmedNodeText(ctx.SourceFile, argument)+")")}
	})
}

func checkSqrt(ctx rule.RuleContext, node *ast.Node) {
	call, ok := mathCall(node, "sqrt")
	if !ok {
		return
	}
	var keys unicornutil.ComputedKeyEvaluator
	var additions, squares []*ast.Node
	var collect func(*ast.Node) bool
	collect = func(expression *ast.Node) bool {
		expression = utils.ESTreeRuntimeExpression(expression)
		if expression == nil || !ast.IsBinaryExpression(expression) {
			return false
		}
		binary := expression.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindPlusToken:
			additions = append(additions, expression)
			return collect(binary.Left) && collect(binary.Right)
		case ast.KindAsteriskToken:
			// This comparator includes static strings and erased TypeScript
			// wrappers, as upstream's isSameReference does in v79.
			if !unicornutil.IsSameComparisonOperandWithKeyEvaluator(binary.Left, binary.Right, keys.Eval) {
				return false
			}
		case ast.KindAsteriskAsteriskToken:
			right := utils.ESTreeRuntimeExpression(binary.Right)
			if right == nil || !ast.IsNumericLiteral(right) || jsnum.FromString(right.Text()) != 2 {
				return false
			}
		default:
			return false
		}
		squares = append(squares, expression)
		return true
	}
	if !collect(call.Call.Arguments()[0]) {
		return
	}
	method := "abs"
	if len(squares) > 1 {
		method = "hypot"
	}
	autofix := canAutoFixSquares(ctx, node, squares) && ctx.Globals.Access("Math").IsDeclared() &&
		unicornutil.IsGlobalReference(ctx, utils.ESTreeRuntimeExpression(call.Object))
	report(ctx, call.Property, method, "Math.sqrt(…)", autofix, func() []rule.RuleFix {
		fixes := []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, call.Property, method)}
		for _, addition := range additions {
			plus := utils.TrimNodeTextRange(ctx.SourceFile, addition.AsBinaryExpression().OperatorToken)
			text := ctx.SourceFile.Text()
			start := len(strings.TrimRightFunc(text[:plus.Pos()], ecmascript.IsWhiteSpaceOrLineTerminator))
			whitespace := text[start:plus.Pos()]
			lineBreak := ""
			if strings.HasPrefix(whitespace, "\r\n") {
				lineBreak = "\r\n"
			} else if strings.HasPrefix(whitespace, "\n") || strings.HasPrefix(whitespace, "\r") {
				lineBreak = whitespace[:1]
			}
			fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(start, plus.End()), lineBreak+","))
			for expression := addition; expression.Parent != nil; {
				parent := expression.Parent
				// ESTree omits the assertion wrappers tsgo synthesizes for
				// JSDoc casts, including those between pairs of parentheses.
				if !ast.IsParenthesizedExpression(parent) {
					if utils.IsJSDocTypeCastWrapper(parent) {
						expression = parent
						continue
					}
					break
				}
				parentRange := utils.TrimNodeTextRange(ctx.SourceFile, parent)
				fixes = append(fixes,
					rule.RuleFixRemoveRange(core.NewTextRange(parentRange.Pos(), parentRange.Pos()+1)),
					rule.RuleFixRemoveRange(core.NewTextRange(parentRange.End()-1, parentRange.End())))
				expression = parent
			}
		}
		for _, square := range squares {
			left := square.AsBinaryExpression().Left
			start, end := left.End(), square.End()
			if utils.HasCommentInSpan(ctx.Comments.All(), start, end) {
				return nil
			}
			fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(start, end)))
		}
		return fixes
	})
}

func report(ctx rule.RuleContext, node *ast.Node, method, description string, autofix bool, build func() []rule.RuleFix) {
	replacement := "Math." + method + "(…)"
	message := rule.RuleMessage{
		Id: messageID, Description: "Prefer `" + replacement + "` over `" + description + "`.",
		Data: map[string]string{"replacement": replacement, "description": description},
	}
	if autofix {
		ctx.ReportNodeWithDeferredFixes(node, message, build)
	} else {
		ctx.ReportNodeWithDeferredSuggestions(node, message, func() []rule.RuleSuggestion {
			fixes := build()
			if len(fixes) == 0 {
				return nil
			}
			return []rule.RuleSuggestion{{
				Message: rule.RuleMessage{Id: "suggestion", Description: "Replace with `" + replacement + "`.",
					Data: map[string]string{"replacement": replacement}},
				FixesArr: fixes,
			}}
		})
	}
}
