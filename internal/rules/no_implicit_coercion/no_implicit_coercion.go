package no_implicit_coercion

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type options struct {
	boolean                   bool
	number                    bool
	str                       bool
	disallowTemplateShorthand bool
	allow                     *utils.Set[string]
}

func parseOptions(raw any) options {
	opts := options{
		boolean: true,
		number:  true,
		str:     true,
		allow:   utils.NewSetWithSizeHint[string](0),
	}
	m := utils.GetOptionsMap(raw)
	if m == nil {
		return opts
	}
	if v, ok := m["boolean"].(bool); ok {
		opts.boolean = v
	}
	if v, ok := m["number"].(bool); ok {
		opts.number = v
	}
	if v, ok := m["string"].(bool); ok {
		opts.str = v
	}
	if v, ok := m["disallowTemplateShorthand"].(bool); ok {
		opts.disallowTemplateShorthand = v
	}
	for _, op := range utils.ToStringSlice(m["allow"]) {
		opts.allow.Add(op)
	}
	return opts
}

// NoImplicitCoercionRule disallows shorthand type conversions, suggesting
// explicit `Boolean()` / `Number()` / `String()` calls instead.
// https://eslint.org/docs/latest/rules/no-implicit-coercion
var NoImplicitCoercionRule = rule.Rule{
	Name: "no-implicit-coercion",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)

		// report emits the diagnostic for `node` (replaced with `recommendation`).
		// ESLint's semantics:
		//   - shouldFix  → autofix applied (no suggestion).
		//   - shouldSuggest (and !shouldFix) → suggestion only.
		//   - neither → plain report (no fix, no suggestion).
		report := func(node *ast.Node, recommendation string, shouldSuggest, shouldFix bool) {
			msg := rule.RuleMessage{
				Id:          "implicitCoercion",
				Description: "Unexpected implicit coercion encountered. Use `" + recommendation + "` instead.",
			}
			replacement := recommendation
			// Guard `typeof+foo` → `typeof Number(foo)`: without a space, the fix
			// would lex as a single `typeofNumber` identifier.
			start := utils.TrimNodeTextRange(ctx.SourceFile, node).Pos()
			if utils.NeedsLeadingSpaceForReplacement(ctx.SourceFile.Text(), start, recommendation) {
				replacement = " " + recommendation
			}

			if shouldFix {
				ctx.ReportNodeWithFixes(node, msg, rule.RuleFixReplace(ctx.SourceFile, node, replacement))
				return
			}
			if shouldSuggest {
				ctx.ReportNodeWithSuggestions(node, msg, rule.RuleSuggestion{
					Message: rule.RuleMessage{
						Id:          "useRecommendation",
						Description: "Use `" + recommendation + "` instead.",
					},
					FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, replacement)},
				})
				return
			}
			ctx.ReportNode(node, msg)
		}

		checkUnary := func(node *ast.Node) {
			pue := node.AsPrefixUnaryExpression()
			if pue == nil || pue.Operand == nil {
				return
			}

			// !!foo → Boolean(foo) — autofix unless `Boolean` is locally shadowed.
			// Parens between the two `!`s (e.g. `!(!foo)`) are transparent in
			// ESLint's AST, so we peel them off to match that behavior.
			if opts.boolean && !opts.allow.Has("!!") && isDoubleLogicalNegating(pue) {
				innerPue := ast.SkipParentheses(pue.Operand).AsPrefixUnaryExpression()
				target := ast.SkipParentheses(innerPue.Operand)
				recommendation := "Boolean(" + utils.TrimmedNodeText(ctx.SourceFile, target) + ")"
				shouldFix := !utils.IsShadowed(node, "Boolean")
				report(node, recommendation, true, shouldFix)
			}

			// ~foo.indexOf(bar) → foo.indexOf(bar) !== -1 (or `>= 0` on an
			// optional chain, since `?.indexOf(x) !== -1` flips to true when
			// `foo` is nullish). No autofix: the rewrite changes semantics on
			// non-array targets (e.g. `String.prototype.indexOf`).
			if opts.boolean && !opts.allow.Has("~") && isBinaryNegatingOfIndexOf(pue) {
				callNode := ast.SkipParentheses(pue.Operand)
				comparison := "!== -1"
				if ast.IsOptionalChain(callNode) {
					comparison = ">= 0"
				}
				recommendation := utils.TrimmedNodeText(ctx.SourceFile, callNode) + " " + comparison
				report(node, recommendation, false, false)
			}

			// +foo → Number(foo) (suggestion only — `+` on BigInt throws).
			if opts.number && !opts.allow.Has("+") && pue.Operator == ast.KindPlusToken {
				operand := ast.SkipParentheses(pue.Operand)
				if !isNumeric(operand) {
					recommendation := "Number(" + utils.TrimmedNodeText(ctx.SourceFile, operand) + ")"
					report(node, recommendation, true, false)
				}
			}

			// -(-foo) → Number(foo) (suggestion only — same BigInt caveat).
			if opts.number && !opts.allow.Has("- -") && pue.Operator == ast.KindMinusToken {
				inner := ast.SkipParentheses(pue.Operand)
				if inner != nil && inner.Kind == ast.KindPrefixUnaryExpression {
					innerPue := inner.AsPrefixUnaryExpression()
					operand := ast.SkipParentheses(innerPue.Operand)
					if innerPue.Operator == ast.KindMinusToken && !isNumeric(operand) {
						recommendation := "Number(" + utils.TrimmedNodeText(ctx.SourceFile, operand) + ")"
						report(node, recommendation, true, false)
					}
				}
			}
		}

		checkBinary := func(node *ast.Node) {
			bin := node.AsBinaryExpression()
			if bin == nil || bin.OperatorToken == nil {
				return
			}

			switch bin.OperatorToken.Kind {
			case ast.KindAsteriskToken:
				// `1 * foo` / `foo * 1` → Number(foo). `a * 1 / b` is skipped:
				// the reader naturally parses it as `a * (1 / b)` rather than
				// `(a * 1) / b`, so there's no coercion intent (issue eslint#16373).
				if opts.number && !opts.allow.Has("*") && isMultiplyByOne(bin) && !isMultiplyByFractionOfOne(node) {
					if operand := getNonNumericOperand(bin); operand != nil {
						recommendation := "Number(" + utils.TrimmedNodeText(ctx.SourceFile, operand) + ")"
						report(node, recommendation, true, false)
					}
				}
			case ast.KindMinusToken:
				// foo - 0 → Number(foo).
				if opts.number && !opts.allow.Has("-") {
					left := ast.SkipParentheses(bin.Left)
					if isLiteralZero(ast.SkipParentheses(bin.Right)) && !isNumeric(left) {
						recommendation := "Number(" + utils.TrimmedNodeText(ctx.SourceFile, left) + ")"
						report(node, recommendation, true, false)
					}
				}
			case ast.KindPlusToken:
				// "" + foo / foo + "" → String(foo).
				if opts.str && !opts.allow.Has("+") && isConcatWithEmptyString(bin) {
					operand := ast.SkipParentheses(getNonEmptyOperand(bin))
					recommendation := "String(" + utils.TrimmedNodeText(ctx.SourceFile, operand) + ")"
					report(node, recommendation, true, false)
				}
			case ast.KindPlusEqualsToken:
				// foo += "" → foo = String(foo).
				if opts.str && !opts.allow.Has("+") && isEmptyString(ast.SkipParentheses(bin.Right)) {
					leftText := utils.TrimmedNodeText(ctx.SourceFile, ast.SkipParentheses(bin.Left))
					recommendation := leftText + " = String(" + leftText + ")"
					report(node, recommendation, true, false)
				}
			}
		}

		checkTemplate := func(node *ast.Node) {
			if !opts.disallowTemplateShorthand {
				return
			}
			// tag`${foo}` is not a coercion — the tag function decides the result.
			if node.Parent != nil && node.Parent.Kind == ast.KindTaggedTemplateExpression {
				return
			}
			te := node.AsTemplateExpression()
			if te == nil || te.Head == nil || te.TemplateSpans == nil {
				return
			}
			// Only `` `${expr}` `` — exactly one span and empty head/tail.
			if len(te.TemplateSpans.Nodes) != 1 || te.Head.Text() != "" {
				return
			}
			span := te.TemplateSpans.Nodes[0].AsTemplateSpan()
			if span == nil || span.Literal == nil || span.Expression == nil || span.Literal.Text() != "" {
				return
			}
			// Already a string — no coercion happening.
			expr := ast.SkipParentheses(span.Expression)
			if isStringType(expr) {
				return
			}
			recommendation := "String(" + utils.TrimmedNodeText(ctx.SourceFile, expr) + ")"
			report(node, recommendation, true, false)
		}

		return rule.RuleListeners{
			ast.KindPrefixUnaryExpression: checkUnary,
			// :exit on BinaryExpression matches ESLint's ordering and keeps us
			// from reporting the same multiplicative chain twice.
			rule.ListenerOnExit(ast.KindBinaryExpression): checkBinary,
			ast.KindTemplateExpression:                    checkTemplate,
		}
	},
}

// isDoubleLogicalNegating reports whether node is `!!x` (with any number of
// intervening parentheses around the inner `!x`).
func isDoubleLogicalNegating(pue *ast.PrefixUnaryExpression) bool {
	if pue.Operator != ast.KindExclamationToken {
		return false
	}
	operand := ast.SkipParentheses(pue.Operand)
	if operand == nil || operand.Kind != ast.KindPrefixUnaryExpression {
		return false
	}
	inner := operand.AsPrefixUnaryExpression()
	return inner != nil && inner.Operator == ast.KindExclamationToken
}

// isBinaryNegatingOfIndexOf reports whether node is `~x.indexOf(y)` or
// `~x.lastIndexOf(y)` (optionally chained, optionally parenthesised).
func isBinaryNegatingOfIndexOf(pue *ast.PrefixUnaryExpression) bool {
	if pue.Operator != ast.KindTildeToken {
		return false
	}
	call := ast.SkipParentheses(pue.Operand)
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	if callee == nil {
		return false
	}
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		name := callee.AsPropertyAccessExpression().Name()
		return name != nil && isIndexOfName(name.Text())
	case ast.KindElementAccessExpression:
		// Strip parens on the computed key so `foo[('indexOf')](x)` matches —
		// ESLint treats parens transparently in `isSpecificMemberAccess`.
		arg := ast.SkipParentheses(callee.AsElementAccessExpression().ArgumentExpression)
		if arg == nil {
			return false
		}
		switch arg.Kind {
		case ast.KindStringLiteral:
			return isIndexOfName(arg.AsStringLiteral().Text)
		case ast.KindNoSubstitutionTemplateLiteral:
			return isIndexOfName(arg.AsNoSubstitutionTemplateLiteral().Text)
		}
	}
	return false
}

func isIndexOfName(s string) bool {
	return s == "indexOf" || s == "lastIndexOf"
}

// isMultiplyByOne reports whether bin is a `*` expression with an operand
// equal to the literal `1`.
func isMultiplyByOne(bin *ast.BinaryExpression) bool {
	if bin.OperatorToken.Kind != ast.KindAsteriskToken {
		return false
	}
	return isLiteralOne(ast.SkipParentheses(bin.Left)) || isLiteralOne(ast.SkipParentheses(bin.Right))
}

// isMultiplyByFractionOfOne reports whether node is the `x * 1` half of a
// `x * 1 / y` expression. In that case the whole expression is naturally
// read as `x * (1 / y)`, so there's no coercion intent worth flagging.
// Mirrors ESLint's fix for eslint/eslint#16373.
func isMultiplyByFractionOfOne(node *ast.Node) bool {
	bin := node.AsBinaryExpression()
	if bin.OperatorToken.Kind != ast.KindAsteriskToken {
		return false
	}
	if !isLiteralOne(ast.SkipParentheses(bin.Right)) {
		return false
	}
	// If the node is parenthesised, `(x * 1) / y` can't be reinterpreted as
	// `x * (1 / y)` — the parens pin the grouping to the coercion form.
	if node.Parent == nil || node.Parent.Kind == ast.KindParenthesizedExpression {
		return false
	}
	parent := node.Parent
	if parent.Kind != ast.KindBinaryExpression {
		return false
	}
	parentBin := parent.AsBinaryExpression()
	if parentBin.OperatorToken.Kind != ast.KindSlashToken {
		return false
	}
	return parentBin.Left == node
}

// getNonNumericOperand returns the non-numeric operand of a `*` expression
// when exactly one non-BinaryExpression side is non-numeric. Walked bottom-up
// (hence the right-then-left order) to match ESLint's traversal behavior.
func getNonNumericOperand(bin *ast.BinaryExpression) *ast.Node {
	left, right := ast.SkipParentheses(bin.Left), ast.SkipParentheses(bin.Right)
	if right != nil && right.Kind != ast.KindBinaryExpression && !isNumeric(right) {
		return right
	}
	if left != nil && left.Kind != ast.KindBinaryExpression && !isNumeric(left) {
		return left
	}
	return nil
}

// isNumeric reports whether node statically evaluates to a number: a numeric
// literal or a call to `Number`, `parseInt`, or `parseFloat`. Shadowing of
// these globals is ignored — matches ESLint, which also uses a syntactic check.
// Parens on the callee (e.g. `(Number)(x)`) are peeled off to match ESLint's
// paren-transparent AST.
func isNumeric(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindNumericLiteral {
		return true
	}
	if node.Kind == ast.KindCallExpression {
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		if callee != nil && callee.Kind == ast.KindIdentifier {
			switch callee.AsIdentifier().Text {
			case "Number", "parseInt", "parseFloat":
				return true
			}
		}
	}
	return false
}

// isStringType reports whether node statically evaluates to a string: any
// string/template literal, or a call to `String`. Parens on the callee
// (e.g. `(String)(x)`) are peeled off to match ESLint's paren-transparent AST.
func isStringType(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression:
		return true
	case ast.KindCallExpression:
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		if callee != nil && callee.Kind == ast.KindIdentifier && callee.AsIdentifier().Text == "String" {
			return true
		}
	}
	return false
}

// isEmptyString reports whether node is `""` or “ “ “.
func isEmptyString(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text == ""
	case ast.KindNoSubstitutionTemplateLiteral:
		return node.AsNoSubstitutionTemplateLiteral().Text == ""
	}
	return false
}

// isConcatWithEmptyString reports whether bin is `"" + x` or `x + ""` where
// `x` isn't already a string (otherwise `+` is plain concatenation).
func isConcatWithEmptyString(bin *ast.BinaryExpression) bool {
	if bin.OperatorToken.Kind != ast.KindPlusToken {
		return false
	}
	left, right := ast.SkipParentheses(bin.Left), ast.SkipParentheses(bin.Right)
	if isEmptyString(left) && !isStringType(right) {
		return true
	}
	if isEmptyString(right) && !isStringType(left) {
		return true
	}
	return false
}

// getNonEmptyOperand returns the side of a `"" + x` / `x + ""` that isn't the
// empty string. Returns the original (possibly parenthesised) node so the
// recommendation text reproduces the source verbatim.
func getNonEmptyOperand(bin *ast.BinaryExpression) *ast.Node {
	if isEmptyString(ast.SkipParentheses(bin.Left)) {
		return bin.Right
	}
	return bin.Left
}

// isLiteralOne / isLiteralZero match numeric literals whose value equals
// `1` / `0` (including `1.0`, `0x1`, `1e0`, etc.), mirroring ESLint's
// `value === 1` / `value === 0` check on Literal nodes.
func isLiteralOne(node *ast.Node) bool {
	return isLiteralNumberEqual(node, "1")
}

func isLiteralZero(node *ast.Node) bool {
	return isLiteralNumberEqual(node, "0")
}

func isLiteralNumberEqual(node *ast.Node, normalized string) bool {
	if node == nil || node.Kind != ast.KindNumericLiteral {
		return false
	}
	return utils.NormalizeNumericLiteral(node.AsNumericLiteral().Text) == normalized
}
