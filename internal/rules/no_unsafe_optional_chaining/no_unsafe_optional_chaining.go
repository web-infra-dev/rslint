package no_unsafe_optional_chaining

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-unsafe-optional-chaining
var NoUnsafeOptionalChainingRule = rule.Rule{
	Name: "no-unsafe-optional-chaining",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		unsafeOptionalChainMsg := rule.RuleMessage{
			Id:          "unsafeOptionalChain",
			Description: "Unsafe usage of optional chaining. If it short-circuits with 'undefined' the evaluation will throw TypeError.",
		}

		unsafeArithmeticMsg := rule.RuleMessage{
			Id:          "unsafeArithmetic",
			Description: "Unsafe arithmetic operation on optional chaining. It can result in NaN.",
		}

		// reportIfOptionalChain reports on the node if the given expression contains an optional chain.
		reportIfOptionalChain := func(node *ast.Node, expr *ast.Node, msg rule.RuleMessage) {
			if expr != nil && hasOptionalChain(expr) {
				ctx.ReportNode(node, msg)
			}
		}

		listeners := rule.RuleListeners{}

		// 1. CallExpression: if callee has optional chain and the call itself is NOT optional
		listeners[ast.KindCallExpression] = func(node *ast.Node) {
			ce := node.AsCallExpression()
			if ce == nil {
				return
			}
			// If this call itself uses optional chaining (?.), it's safe
			if ast.IsOptionalChain(node) {
				return
			}
			reportIfOptionalChain(node, ce.Expression, unsafeOptionalChainMsg)
		}

		// 2. NewExpression: if callee has optional chain
		listeners[ast.KindNewExpression] = func(node *ast.Node) {
			ne := node.AsNewExpression()
			if ne == nil {
				return
			}
			reportIfOptionalChain(node, ne.Expression, unsafeOptionalChainMsg)
		}

		// 3. TaggedTemplateExpression: if tag has optional chain and is NOT optional
		listeners[ast.KindTaggedTemplateExpression] = func(node *ast.Node) {
			tte := node.AsTaggedTemplateExpression()
			if tte == nil {
				return
			}
			if ast.IsOptionalChain(node) {
				return
			}
			reportIfOptionalChain(node, tte.Tag, unsafeOptionalChainMsg)
		}

		// 4. PropertyAccessExpression: if object has optional chain and access is NOT optional
		listeners[ast.KindPropertyAccessExpression] = func(node *ast.Node) {
			pae := node.AsPropertyAccessExpression()
			if pae == nil {
				return
			}
			if ast.IsOptionalChain(node) {
				return
			}
			reportIfOptionalChain(node, pae.Expression, unsafeOptionalChainMsg)
		}

		// 5. ElementAccessExpression: if object has optional chain and access is NOT optional
		listeners[ast.KindElementAccessExpression] = func(node *ast.Node) {
			eae := node.AsElementAccessExpression()
			if eae == nil {
				return
			}
			if ast.IsOptionalChain(node) {
				return
			}
			reportIfOptionalChain(node, eae.Expression, unsafeOptionalChainMsg)
		}

		// 6. BinaryExpression: handles in/instanceof, destructuring assignment, comma (sequence),
		//    and arithmetic operators
		listeners[ast.KindBinaryExpression] = func(node *ast.Node) {
			bin := node.AsBinaryExpression()
			if bin == nil || bin.OperatorToken == nil {
				return
			}

			op := bin.OperatorToken.Kind

			switch op {
			case ast.KindInKeyword, ast.KindInstanceOfKeyword:
				// Right operand of in/instanceof must not be optional chain
				reportIfOptionalChain(node, bin.Right, unsafeOptionalChainMsg)

			case ast.KindEqualsToken:
				// Destructuring assignment: if left side is a binding/destructuring pattern
				// and right side has optional chain
				if isDestructuringPattern(bin.Left) {
					reportIfOptionalChain(node, bin.Right, unsafeOptionalChainMsg)
				}

			case ast.KindPlusToken, ast.KindMinusToken, ast.KindAsteriskToken,
				ast.KindSlashToken, ast.KindPercentToken, ast.KindAsteriskAsteriskToken:
				// Arithmetic operators (only when option is enabled)
				if opts.disallowArithmeticOperators {
					if hasOptionalChain(bin.Left) || hasOptionalChain(bin.Right) {
						ctx.ReportNode(node, unsafeArithmeticMsg)
					}
				}
			}
		}

		// 7. PrefixUnaryExpression: unary +/- with optional chain operand
		if opts.disallowArithmeticOperators {
			listeners[ast.KindPrefixUnaryExpression] = func(node *ast.Node) {
				pue := node.AsPrefixUnaryExpression()
				if pue == nil {
					return
				}
				if pue.Operator == ast.KindPlusToken || pue.Operator == ast.KindMinusToken {
					reportIfOptionalChain(node, pue.Operand, unsafeArithmeticMsg)
				}
			}
		}

		// 8. SpreadElement: if argument has optional chain
		listeners[ast.KindSpreadElement] = func(node *ast.Node) {
			se := node.AsSpreadElement()
			if se == nil {
				return
			}
			reportIfOptionalChain(node, se.Expression, unsafeOptionalChainMsg)
		}

		// 9. ForOfStatement: if right side has optional chain
		listeners[ast.KindForOfStatement] = func(node *ast.Node) {
			foStmt := node.AsForInOrOfStatement()
			if foStmt == nil {
				return
			}
			reportIfOptionalChain(node, foStmt.Expression, unsafeOptionalChainMsg)
		}

		// 10. VariableDeclaration: destructuring with optional chain initializer
		listeners[ast.KindVariableDeclaration] = func(node *ast.Node) {
			vd := node.AsVariableDeclaration()
			if vd == nil {
				return
			}
			// Check if the name is a destructuring pattern (object or array binding)
			name := vd.Name()
			if name == nil {
				return
			}
			if name.Kind == ast.KindObjectBindingPattern || name.Kind == ast.KindArrayBindingPattern {
				reportIfOptionalChain(node, vd.Initializer, unsafeOptionalChainMsg)
			}
		}

		// 11. Class extends: check heritage clauses for optional chain in super class
		classHandler := func(node *ast.Node) {
			var heritageClauses *ast.NodeList
			switch node.Kind {
			case ast.KindClassDeclaration:
				heritageClauses = node.AsClassDeclaration().HeritageClauses
			case ast.KindClassExpression:
				heritageClauses = node.AsClassExpression().HeritageClauses
			}
			if heritageClauses == nil {
				return
			}
			for _, clause := range heritageClauses.Nodes {
				hc := clause.AsHeritageClause()
				if hc == nil || hc.Token != ast.KindExtendsKeyword {
					continue
				}
				if hc.Types == nil {
					continue
				}
				for _, typeNode := range hc.Types.Nodes {
					ewta := typeNode.AsExpressionWithTypeArguments()
					if ewta == nil {
						continue
					}
					reportIfOptionalChain(typeNode, ewta.Expression, unsafeOptionalChainMsg)
				}
			}
		}
		listeners[ast.KindClassDeclaration] = classHandler
		listeners[ast.KindClassExpression] = classHandler

		return listeners
	},
}

// hasOptionalChain recursively checks whether a node tree contains an optional chain
// that could produce undefined. This handles:
// - Direct optional chain nodes (PropertyAccessExpression, ElementAccessExpression, CallExpression with ?.)
// - ParenthesizedExpression: check inner expression
// - LogicalExpression with || or ??: check right side only (left provides fallback)
// - LogicalExpression with &&: check both sides
// - ConditionalExpression: check both consequent and alternate
// - BinaryExpression with , (comma/sequence): check right side only
func hasOptionalChain(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindPropertyAccessExpression,
		ast.KindElementAccessExpression,
		ast.KindCallExpression,
		ast.KindNonNullExpression:
		// Check if this node is part of an optional chain
		if ast.IsOptionalChain(node) {
			return true
		}
		return false

	case ast.KindParenthesizedExpression:
		pe := node.AsParenthesizedExpression()
		if pe == nil {
			return false
		}
		return hasOptionalChain(pe.Expression)

	case ast.KindBinaryExpression:
		bin := node.AsBinaryExpression()
		if bin == nil || bin.OperatorToken == nil {
			return false
		}
		op := bin.OperatorToken.Kind
		switch op {
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			// || and ?? — only the right side matters because if left short-circuits
			// to undefined, the right side provides the fallback
			return hasOptionalChain(bin.Right)
		case ast.KindAmpersandAmpersandToken:
			// && — both sides can propagate undefined
			return hasOptionalChain(bin.Left) || hasOptionalChain(bin.Right)
		case ast.KindCommaToken:
			// Comma/sequence — only the last (right) expression matters
			return hasOptionalChain(bin.Right)
		default:
			return false
		}

	case ast.KindConditionalExpression:
		ce := node.AsConditionalExpression()
		if ce == nil {
			return false
		}
		return hasOptionalChain(ce.WhenTrue) || hasOptionalChain(ce.WhenFalse)

	default:
		return false
	}
}

// isDestructuringPattern checks if a node is a destructuring assignment pattern
// (object or array literal used as assignment target).
func isDestructuringPattern(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression:
		return true
	case ast.KindParenthesizedExpression:
		pe := node.AsParenthesizedExpression()
		if pe == nil {
			return false
		}
		return isDestructuringPattern(pe.Expression)
	default:
		return false
	}
}

type noUnsafeOptionalChainingOptions struct {
	disallowArithmeticOperators bool
}

func parseOptions(opts any) noUnsafeOptionalChainingOptions {
	result := noUnsafeOptionalChainingOptions{
		disallowArithmeticOperators: false,
	}

	optsMap := utils.GetOptionsMap(opts)
	if optsMap != nil {
		if disallow, ok := optsMap["disallowArithmeticOperators"].(bool); ok {
			result.disallowArithmeticOperators = disallow
		}
	}

	return result
}
