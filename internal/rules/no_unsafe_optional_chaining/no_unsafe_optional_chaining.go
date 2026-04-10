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

		checkUnsafeUsage := func(expr *ast.Node) {
			checkUndefinedShortCircuit(expr, func(chainNode *ast.Node) {
				ctx.ReportNode(chainNode, unsafeOptionalChainMsg)
			})
		}

		checkUnsafeArithmetic := func(expr *ast.Node) {
			checkUndefinedShortCircuit(expr, func(chainNode *ast.Node) {
				ctx.ReportNode(chainNode, unsafeArithmeticMsg)
			})
		}

		listeners := rule.RuleListeners{}

		// CallExpression: (obj?.foo)()
		listeners[ast.KindCallExpression] = func(node *ast.Node) {
			if ast.IsOptionalChain(node) {
				return
			}
			checkUnsafeUsage(node.AsCallExpression().Expression)
		}

		// NewExpression: new (obj?.foo)()
		listeners[ast.KindNewExpression] = func(node *ast.Node) {
			checkUnsafeUsage(node.AsNewExpression().Expression)
		}

		// TaggedTemplateExpression: (obj?.foo)`text`
		listeners[ast.KindTaggedTemplateExpression] = func(node *ast.Node) {
			checkUnsafeUsage(node.AsTaggedTemplateExpression().Tag)
		}

		// PropertyAccessExpression: (obj?.foo).bar
		listeners[ast.KindPropertyAccessExpression] = func(node *ast.Node) {
			if ast.IsOptionalChain(node) {
				return
			}
			checkUnsafeUsage(node.AsPropertyAccessExpression().Expression)
		}

		// ElementAccessExpression: (obj?.foo)[0]
		listeners[ast.KindElementAccessExpression] = func(node *ast.Node) {
			if ast.IsOptionalChain(node) {
				return
			}
			checkUnsafeUsage(node.AsElementAccessExpression().Expression)
		}

		// BinaryExpression: in/instanceof, destructuring assignment, arithmetic
		listeners[ast.KindBinaryExpression] = func(node *ast.Node) {
			bin := node.AsBinaryExpression()
			if bin == nil || bin.OperatorToken == nil {
				return
			}

			// Destructuring assignment: ({x} = obj?.foo)
			if ast.IsDestructuringAssignment(node) {
				checkUnsafeUsage(bin.Right)
				return
			}

			op := bin.OperatorToken.Kind

			switch op {
			case ast.KindInKeyword, ast.KindInstanceOfKeyword:
				checkUnsafeUsage(bin.Right)

			case ast.KindPlusToken, ast.KindMinusToken, ast.KindAsteriskToken,
				ast.KindSlashToken, ast.KindPercentToken, ast.KindAsteriskAsteriskToken:
				if opts.disallowArithmeticOperators {
					checkUnsafeArithmetic(bin.Left)
					checkUnsafeArithmetic(bin.Right)
				}

			case ast.KindPlusEqualsToken, ast.KindMinusEqualsToken,
				ast.KindAsteriskEqualsToken, ast.KindAsteriskAsteriskEqualsToken,
				ast.KindSlashEqualsToken, ast.KindPercentEqualsToken:
				if opts.disallowArithmeticOperators {
					checkUnsafeArithmetic(bin.Right)
				}
			}
		}

		// PrefixUnaryExpression: +obj?.foo, -obj?.foo
		if opts.disallowArithmeticOperators {
			listeners[ast.KindPrefixUnaryExpression] = func(node *ast.Node) {
				pue := node.AsPrefixUnaryExpression()
				if pue.Operator == ast.KindPlusToken || pue.Operator == ast.KindMinusToken {
					checkUnsafeArithmetic(pue.Operand)
				}
			}
		}

		// SpreadElement: [...obj?.foo] is unsafe, but {...obj?.foo} is safe
		listeners[ast.KindSpreadElement] = func(node *ast.Node) {
			if node.Parent != nil && node.Parent.Kind == ast.KindObjectLiteralExpression {
				return
			}
			checkUnsafeUsage(node.AsSpreadElement().Expression)
		}

		// ForOfStatement: for (const x of obj?.foo)
		listeners[ast.KindForOfStatement] = func(node *ast.Node) {
			checkUnsafeUsage(node.AsForInOrOfStatement().Expression)
		}

		// VariableDeclaration: const {x} = obj?.foo
		listeners[ast.KindVariableDeclaration] = func(node *ast.Node) {
			vd := node.AsVariableDeclaration()
			if name := vd.Name(); name != nil && ast.IsBindingPattern(name) {
				checkUnsafeUsage(vd.Initializer)
			}
		}

		// BindingElement: const {x: {y} = obj?.foo} = obj, function f({x: {y} = obj?.foo}) {}
		listeners[ast.KindBindingElement] = func(node *ast.Node) {
			be := node.AsBindingElement()
			if name := be.Name(); name != nil && ast.IsBindingPattern(name) {
				checkUnsafeUsage(be.Initializer)
			}
		}

		// WithStatement: with (obj?.foo) {}
		listeners[ast.KindWithStatement] = func(node *ast.Node) {
			checkUnsafeUsage(node.AsWithStatement().Expression)
		}

		// Class extends: class Foo extends obj?.foo {}
		classHandler := func(node *ast.Node) {
			extendsElement := ast.GetClassExtendsHeritageElement(node)
			if extendsElement == nil {
				return
			}
			exprWithType := extendsElement.AsExpressionWithTypeArguments()
			if exprWithType == nil {
				return
			}
			checkUnsafeUsage(exprWithType.Expression)
		}
		listeners[ast.KindClassDeclaration] = classHandler
		listeners[ast.KindClassExpression] = classHandler

		return listeners
	},
}

// checkUndefinedShortCircuit recursively traverses expression wrappers to find optional chain
// nodes that could short-circuit with undefined, calling report for each one found.
// Matches ESLint's checkUndefinedShortCircuit behavior.
func checkUndefinedShortCircuit(node *ast.Node, report func(chainNode *ast.Node)) {
	if node == nil {
		return
	}

	node = ast.SkipParentheses(node)

	switch node.Kind {
	case ast.KindPropertyAccessExpression,
		ast.KindElementAccessExpression,
		ast.KindCallExpression:
		if ast.IsOptionalChain(node) {
			report(node)
		}

	case ast.KindAwaitExpression:
		checkUndefinedShortCircuit(node.AsAwaitExpression().Expression, report)

	case ast.KindBinaryExpression:
		bin := node.AsBinaryExpression()
		if bin == nil || bin.OperatorToken == nil {
			return
		}
		switch bin.OperatorToken.Kind {
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			// || and ?? — only right side; left provides fallback
			checkUndefinedShortCircuit(bin.Right, report)
		case ast.KindAmpersandAmpersandToken:
			// && — both sides can propagate undefined
			checkUndefinedShortCircuit(bin.Left, report)
			checkUndefinedShortCircuit(bin.Right, report)
		case ast.KindCommaToken:
			// Comma/sequence — only last (right) expression matters
			checkUndefinedShortCircuit(bin.Right, report)
		}

	case ast.KindConditionalExpression:
		ce := node.AsConditionalExpression()
		checkUndefinedShortCircuit(ce.WhenTrue, report)
		checkUndefinedShortCircuit(ce.WhenFalse, report)
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
