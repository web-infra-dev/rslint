package prefer_math_trunc

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-math-trunc.js
var PreferMathTruncRule = rule.Rule{
	Name:   "unicorn/prefer-math-trunc",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindBinaryExpression:      func(node *ast.Node) { checkBitwise(&ctx, node) },
			ast.KindPrefixUnaryExpression: func(node *ast.Node) { checkDoubleNot(&ctx, node) },
			ast.KindCallExpression:        func(node *ast.Node) { checkParseInt(&ctx, node) },
		}
	},
}

func checkBitwise(ctx *rule.RuleContext, node *ast.Node) {
	binary := node.AsBinaryExpression()
	operator := binary.OperatorToken
	switch operator.Kind {
	case ast.KindBarToken, ast.KindGreaterThanGreaterThanToken, ast.KindLessThanLessThanToken, ast.KindCaretToken,
		ast.KindBarEqualsToken, ast.KindGreaterThanGreaterThanEqualsToken, ast.KindLessThanLessThanEqualsToken, ast.KindCaretEqualsToken:
	default:
		return
	}
	right := utils.ESTreeRuntimeExpression(binary.Right)
	if !isNumber(right, "0") {
		return
	}
	operatorText := scanner.TokenToString(operator.Kind)
	value := scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, right, false)
	report(ctx, node, binary.Left, "bitwise", operatorText+" "+value, map[string]string{"operator": operatorText, "value": value}, func(sourceRange core.TextRange) []rule.RuleFix {
		left := binary.Left
		if !ast.IsAssignmentOperator(operator.Kind) {
			return replaceWithTruncatedOperand(ctx, node, sourceRange, left)
		}
		if utils.HasCommentInSpan(ctx.Comments.All(), sourceRange.Pos(), sourceRange.End()) ||
			assignmentTargetHasSideEffect(left) {
			return nil
		}
		return []rule.RuleFix{
			rule.RuleFixReplace(ctx.SourceFile, operator, "="),
			rule.RuleFixReplace(ctx.SourceFile, right, truncCallForOperand(ctx, left)),
		}
	})
}

func assignmentTargetHasSideEffect(target *ast.Node) bool {
	target = utils.SkipAssertionsAndParens(target)
	evaluator := utils.NewStaticStringEvaluatorWithoutScope()
	if receiver, key := utils.MemberExpressionParts(target); receiver != nil {
		// The rewrite evaluates the receiver and computed key twice, while
		// reading the final property value once.
		return evaluator.HasSideEffect(receiver, true) || evaluator.HasSideEffect(key, true)
	}
	return evaluator.HasSideEffect(target, true)
}

func checkDoubleNot(ctx *rule.RuleContext, node *ast.Node) {
	outer := node.AsPrefixUnaryExpression()
	if outer.Operator != ast.KindTildeToken {
		return
	}
	inner := utils.ESTreeRuntimeExpression(outer.Operand)
	if !isBitwiseNot(inner) {
		return
	}
	operand := inner.AsPrefixUnaryExpression().Operand
	if isBitwiseNot(utils.ESTreeRuntimeExpression(operand)) {
		return
	}
	report(ctx, node, operand, "bitwise-not", "~~", nil, func(sourceRange core.TextRange) []rule.RuleFix {
		return replaceWithTruncatedOperand(ctx, node, sourceRange, operand)
	})
}

type bigIntKind uint8

const (
	bigIntUnknown bigIntKind = iota
	bigIntPossible
	bigIntKnown
)

func classifyBigInt(ctx *rule.RuleContext, operand *ast.Node) bigIntKind {
	syntax := bigIntFromSyntax(ctx, operand)
	if syntax == bigIntKnown || ctx.TypeChecker == nil {
		return syntax
	}
	t := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, operand)
	parts := utils.UnionTypeParts(t)
	allBigInt, hasBigInt, uncertain := len(parts) > 0, false, false
	for _, part := range parts {
		if utils.IsTypeFlagSet(part, checker.TypeFlagsBigIntLike) {
			hasBigInt = true
		} else {
			allBigInt = false
			uncertain = uncertain || utils.IsTypeFlagSet(part, checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsTypeParameter)
		}
	}
	if allBigInt {
		return bigIntKnown
	}
	if hasBigInt {
		return bigIntPossible
	}
	// An unresolved type must not erase BigInt evidence from the syntax.
	if uncertain || t == nil {
		return syntax
	}
	return bigIntUnknown
}

// Classify the result without evaluating source-controlled arithmetic or
// requiring values for conditions, identifiers or BigInt call arguments.
func bigIntFromSyntax(ctx *rule.RuleContext, operand *ast.Node) bigIntKind {
	operand = utils.SkipAssertionsAndParens(operand)
	if operand == nil {
		return bigIntUnknown
	}
	// Bitwise NOT preserves BigInt, which Math.trunc rejects.
	switch operand.Kind {
	case ast.KindBigIntLiteral:
		return bigIntKnown
	case ast.KindPrefixUnaryExpression:
		unary := operand.AsPrefixUnaryExpression()
		if unary.Operator == ast.KindMinusToken || unary.Operator == ast.KindTildeToken {
			return bigIntFromSyntax(ctx, unary.Operand)
		}
	case ast.KindCallExpression:
		if isGlobalIdentifier(ctx, utils.ESTreeCallCallee(operand.Expression()), "BigInt") {
			if ast.IsOptionalChainRoot(operand) {
				return bigIntPossible
			}
			return bigIntKnown
		}
	case ast.KindConditionalExpression:
		conditional := operand.AsConditionalExpression()
		condition := utils.SkipAssertionsAndParens(conditional.Condition)
		if condition != nil && condition.Kind == ast.KindTrueKeyword {
			return bigIntFromSyntax(ctx, conditional.WhenTrue)
		}
		if condition != nil && condition.Kind == ast.KindFalseKeyword {
			return bigIntFromSyntax(ctx, conditional.WhenFalse)
		}
		return mergeBigIntKinds(bigIntFromSyntax(ctx, conditional.WhenTrue), bigIntFromSyntax(ctx, conditional.WhenFalse))
	case ast.KindBinaryExpression:
		binary := operand.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindCommaToken, ast.KindEqualsToken:
			return bigIntFromSyntax(ctx, binary.Right)
		case ast.KindPlusToken:
			// String concatenation produces a string even with a BigInt operand.
			if ast.IsStringLiteralLike(utils.SkipAssertionsAndParens(binary.Left)) ||
				ast.IsStringLiteralLike(utils.SkipAssertionsAndParens(binary.Right)) {
				return bigIntUnknown
			}
		case ast.KindMinusToken, ast.KindAsteriskToken, ast.KindSlashToken, ast.KindPercentToken,
			ast.KindAsteriskAsteriskToken, ast.KindBarToken, ast.KindAmpersandToken, ast.KindCaretToken,
			ast.KindLessThanLessThanToken, ast.KindGreaterThanGreaterThanToken,
			ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
		default:
			return bigIntUnknown
		}
		return mergeBigIntKinds(bigIntFromSyntax(ctx, binary.Left), bigIntFromSyntax(ctx, binary.Right))
	}
	return bigIntUnknown
}

func mergeBigIntKinds(left, right bigIntKind) bigIntKind {
	if left == bigIntKnown && right == bigIntKnown {
		return bigIntKnown
	}
	if left != bigIntUnknown || right != bigIntUnknown {
		return bigIntPossible
	}
	return bigIntUnknown
}

func checkParseInt(ctx *rule.RuleContext, node *ast.Node) {
	args := node.Arguments()
	if len(args) != 2 || ast.IsOptionalChainRoot(node) {
		return
	}
	stringCall := utils.ESTreeRuntimeExpression(args[0])
	if stringCall == nil || !ast.IsCallExpression(stringCall) || ast.IsOptionalChainRoot(stringCall) ||
		!isNumber(utils.ESTreeRuntimeExpression(args[1]), "10") {
		return
	}
	name := "parseInt"
	if !isGlobalIdentifier(ctx, utils.ESTreeCallCallee(node.Expression()), name) {
		call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{Method: name})
		if !ok || !isGlobalIdentifier(ctx, utils.ESTreeRuntimeExpression(call.Object), "Number") {
			return
		}
		name = "Number.parseInt"
	}
	if !isGlobalIdentifier(ctx, utils.ESTreeCallCallee(stringCall.Expression()), "String") {
		return
	}
	stringArgs := stringCall.ArgumentList()
	if stringArgs == nil || len(stringArgs.Nodes) != 1 || ast.IsSpreadElement(stringArgs.Nodes[0]) {
		return
	}
	report(ctx, node, stringArgs.Nodes[0], "parse-int", name+"(String(...), 10)", map[string]string{"name": name}, func(sourceRange core.TextRange) []rule.RuleFix {
		// The argument list starts just after the opening parenthesis. Copy its
		// interior verbatim, including comments, parentheses and a trailing comma.
		closeParen, ok := utils.TokenAtOrAfter(ctx.SourceFile, stringArgs.End())
		if !ok || closeParen.Kind != ast.KindCloseParenToken || closeParen.End != stringCall.End() {
			return nil
		}
		start, end := stringArgs.Pos(), closeParen.Start
		if hasCommentsOutside(ctx, sourceRange, core.NewTextRange(start, end)) {
			return nil
		}
		return replaceExpression(ctx, node, sourceRange, "Math.trunc("+ctx.SourceFile.Text()[start:end]+")")
	})
}

func isGlobalIdentifier(ctx *rule.RuleContext, node *ast.Node, name string) bool {
	return node != nil && ast.IsIdentifier(node) && node.Text() == name &&
		ctx.Globals.Access(name).IsDeclared() && unicornutil.IsGlobalReference(*ctx, node)
}

// The tsgo scanner stores numeric literals as their canonical Number text.
func isNumber(node *ast.Node, value string) bool {
	return node != nil && ast.IsNumericLiteral(node) && node.Text() == value
}

func isBitwiseNot(node *ast.Node) bool {
	return node != nil && ast.IsPrefixUnaryExpression(node) && node.AsPrefixUnaryExpression().Operator == ast.KindTildeToken
}

func replaceWithTruncatedOperand(ctx *rule.RuleContext, node *ast.Node, sourceRange core.TextRange, operand *ast.Node) []rule.RuleFix {
	if hasCommentsOutside(ctx, sourceRange, utils.TrimNodeTextRange(ctx.SourceFile, operand)) {
		return nil
	}
	return replaceExpression(ctx, node, sourceRange, truncCallForOperand(ctx, operand))
}

func truncCallForOperand(ctx *rule.RuleContext, operand *ast.Node) string {
	expression := utils.ESTreeRuntimeExpression(operand)
	if expression != operand && hasCommentsOutside(ctx,
		utils.TrimNodeTextRange(ctx.SourceFile, operand), utils.TrimNodeTextRange(ctx.SourceFile, expression)) {
		return "Math.trunc(" + scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, operand, false) + ")"
	}
	text := scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, expression, false)
	if ast.IsCommaExpression(expression) {
		text = "(" + text + ")"
	}
	return "Math.trunc(" + text + ")"
}

func hasCommentsOutside(ctx *rule.RuleContext, span, preserved core.TextRange) bool {
	comments := ctx.Comments.All()
	return utils.HasCommentInSpan(comments, span.Pos(), preserved.Pos()) ||
		utils.HasCommentInSpan(comments, preserved.End(), span.End())
}

func replaceExpression(ctx *rule.RuleContext, node *ast.Node, sourceRange core.TextRange, text string) []rule.RuleFix {
	return append(unicornutil.SpaceAroundKeywordFixes(ctx.SourceFile, node), rule.RuleFixReplaceRange(sourceRange, text))
}

func canSuggestMathTrunc(ctx *rule.RuleContext, node *ast.Node) bool {
	// The replacement needs the built-in Math object at this location.
	if !ctx.Globals.Access("Math").IsDeclared() {
		return false
	}
	if ctx.Refs != nil {
		return ctx.Refs.IsGlobalNameReference(node, "Math", ast.SymbolFlagsValue)
	}
	return !utils.IsShadowed(node, "Math")
}

func report(ctx *rule.RuleContext, node, operand *ast.Node, kind, pattern string, data map[string]string, buildFixes func(core.TextRange) []rule.RuleFix) {
	bigInt := classifyBigInt(ctx, operand)
	if bigInt == bigIntKnown {
		return
	}
	sourceRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	ctx.ReportRangeWithDeferredSuggestions(sourceRange, rule.RuleMessage{
		Id:          "error-" + kind,
		Description: "Use `Math.trunc` instead of `" + pattern + "`.",
		Data:        data,
	}, func() []rule.RuleSuggestion {
		if bigInt == bigIntPossible || !canSuggestMathTrunc(ctx, node) {
			return nil
		}
		fixes := buildFixes(sourceRange)
		if len(fixes) == 0 {
			return nil
		}
		return []rule.RuleSuggestion{{
			Message: rule.RuleMessage{
				Id:          "suggestion-" + kind,
				Description: "Replace `" + pattern + "` with `Math.trunc`.",
				Data:        data,
			},
			FixesArr: fixes,
		}}
	})
}
