package no_constant_condition

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed no_constant_condition.schema.json
var schemaJSON []byte

// Message builder
func buildUnexpectedMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: "Unexpected constant condition.",
	}
}

// Options structure
type Options struct {
	CheckLoops string // "all", "allExceptWhileTrue", "none"
}

// parseOptions parses the rule options
func parseOptions(options []any) Options {
	opts := Options{
		CheckLoops: "allExceptWhileTrue", // default
	}

	if len(options) == 0 {
		return opts
	}

	optMap, _ := options[0].(map[string]any)
	if checkLoops, ok := optMap["checkLoops"].(string); ok {
		opts.CheckLoops = checkLoops
	} else if checkLoopsBool, ok := optMap["checkLoops"].(bool); ok {
		// Handle boolean values: true = "all", false = "none"
		if checkLoopsBool {
			opts.CheckLoops = "all"
		} else {
			opts.CheckLoops = "none"
		}
	}

	return opts
}

// getBooleanValue evaluates only ESTree Literal equivalents. Template literals
// and identifiers are not logical identities in the upstream helper.
func getBooleanValue(node *ast.Node) *bool {
	if node == nil {
		return nil
	}
	var value bool
	switch node.Kind {
	case ast.KindTrueKeyword, ast.KindRegularExpressionLiteral:
		value = true
	case ast.KindFalseKeyword, ast.KindNullKeyword:
		value = false
	case ast.KindNumericLiteral:
		number, ok := ecmascript.StringToNumber(node.Text())
		if !ok {
			return nil
		}
		value = number != 0
	case ast.KindBigIntLiteral:
		value = utils.NormalizeBigIntLiteral(node.Text()) != "0"
	case ast.KindStringLiteral:
		// ts-go stores cooked text without the source quotes.
		value = len(node.Text()) != 0
	default:
		return nil
	}
	return &value
}

// isLogicalIdentity checks if a node is a logical identity element
func isLogicalIdentity(node *ast.Node, operator ast.Kind) bool {
	if node == nil {
		return false
	}

	// Check literals
	boolVal := getBooleanValue(node)
	if boolVal != nil {
		if operator == ast.KindBarBarToken && *boolVal {
			return true
		}
		if operator == ast.KindAmpersandAmpersandToken && !*boolVal {
			return true
		}
	}

	// void operator is identity for &&
	if node.Kind == ast.KindVoidExpression {
		return operator == ast.KindAmpersandAmpersandToken
	}

	// Parenthesized expressions
	if node.Kind == ast.KindParenthesizedExpression {
		paren := node.AsParenthesizedExpression()
		if paren != nil && paren.Expression != nil {
			return isLogicalIdentity(paren.Expression, operator)
		}
	}

	// Logical expressions with same operator
	if node.Kind == ast.KindBinaryExpression {
		binary := node.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil {
			nodeOp := binary.OperatorToken.Kind
			if nodeOp == operator && (nodeOp == ast.KindBarBarToken || nodeOp == ast.KindAmpersandAmpersandToken) {
				return isLogicalIdentity(binary.Left, operator) || isLogicalIdentity(binary.Right, operator)
			}
		}
	}

	// Assignment expressions
	if node.Kind == ast.KindBinaryExpression {
		binary := node.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil {
			nodeOp := binary.OperatorToken.Kind
			if nodeOp == ast.KindBarBarEqualsToken || nodeOp == ast.KindAmpersandAmpersandEqualsToken {
				// Extract the base operator (|| or &&)
				var baseOp ast.Kind
				if nodeOp == ast.KindBarBarEqualsToken {
					baseOp = ast.KindBarBarToken
				} else {
					baseOp = ast.KindAmpersandAmpersandToken
				}
				return operator == baseOp && isLogicalIdentity(binary.Right, operator)
			}
		}
	}

	return false
}

// isConstant checks if a node represents a constant value
func isConstant(ctx *rule.RuleContext, node *ast.Node, inBooleanPosition bool) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	// Literal values
	case ast.KindNumericLiteral,
		ast.KindStringLiteral,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword,
		ast.KindRegularExpressionLiteral,
		ast.KindBigIntLiteral:
		return true

	case ast.KindIdentifier:
		return node.Text() == "undefined" && isReferenceToGlobal(ctx, node)

	case ast.KindArrowFunction,
		ast.KindFunctionExpression,
		ast.KindClassExpression:
		return true

	case ast.KindObjectLiteralExpression:
		// Object literals are always truthy in boolean context
		// In non-boolean context, they're considered constant (new object)
		return true

	case ast.KindArrayLiteralExpression:
		// Array literals are always truthy in boolean context
		if inBooleanPosition {
			return true
		}
		// In non-boolean context, only constant if all elements are constant
		arrayLit := node.AsArrayLiteralExpression()
		if arrayLit != nil && arrayLit.Elements != nil {
			for _, elem := range arrayLit.Elements.Nodes {
				// Skip omitted expressions
				if elem.Kind == ast.KindOmittedExpression {
					continue
				}
				// Spread elements: check the spread argument
				if elem.Kind == ast.KindSpreadElement {
					spread := elem.AsSpreadElement()
					if spread != nil && spread.Expression != nil {
						if !isConstant(ctx, spread.Expression, false) {
							return false
						}
					}
					continue
				}
				if !isConstant(ctx, elem, false) {
					return false
				}
			}
		}
		return true

	case ast.KindTemplateExpression:
		// Template literals: constant if any static part has length (in boolean position)
		// or all expressions are constant (not in boolean position)
		template := node.AsTemplateExpression()
		if template == nil {
			return false
		}

		// In boolean position: constant if any quasi has content OR all expressions are constant
		if inBooleanPosition {
			// Check template head for non-empty content
			hasContent := false
			if template.Head != nil {
				text := template.Head.Text()
				if len(text) > 0 {
					hasContent = true
				}
			}

			// If we found static content, it's constant
			if hasContent {
				return true
			}

			// Check template spans for static content
			if template.TemplateSpans != nil {
				for _, span := range template.TemplateSpans.Nodes {
					if span.Kind == ast.KindTemplateSpan {
						templateSpan := span.AsTemplateSpan()
						if templateSpan != nil && templateSpan.Literal != nil {
							text := templateSpan.Literal.Text()
							if len(text) > 0 {
								hasContent = true
								break
							}
						}
					}
				}
			}

			if hasContent {
				return true
			}

			// No static content, so check if all expressions are constant
			// Fall through to check expressions below
		}

		// Check if all expressions are constant
		if template.TemplateSpans != nil {
			for _, span := range template.TemplateSpans.Nodes {
				if span.Kind == ast.KindTemplateSpan {
					templateSpan := span.AsTemplateSpan()
					if templateSpan != nil && templateSpan.Expression != nil {
						if !isConstant(ctx, templateSpan.Expression, false) {
							return false
						}
					}
				}
			}
		}
		return true

	case ast.KindParenthesizedExpression:
		paren := node.AsParenthesizedExpression()
		if paren != nil && paren.Expression != nil {
			return isConstant(ctx, paren.Expression, inBooleanPosition)
		}

	case ast.KindVoidExpression:
		// void operator always returns undefined (constant falsy value)
		// The operand's constantness doesn't matter - void always returns undefined
		return true

	case ast.KindTypeOfExpression:
		// typeof always returns a non-empty string (always truthy in boolean context)
		// In boolean position, typeof is always constant (always truthy)
		// In non-boolean position, typeof is only constant if the operand is constant
		if inBooleanPosition {
			return true
		}
		typeofExpr := node.AsTypeOfExpression()
		if typeofExpr != nil && typeofExpr.Expression != nil {
			return isConstant(ctx, typeofExpr.Expression, false)
		}
		return false

	case ast.KindPrefixUnaryExpression:
		prefix := node.AsPrefixUnaryExpression()
		if prefix == nil {
			return false
		}

		switch prefix.Operator {
		case ast.KindExclamationToken: // !
			if prefix.Operand != nil {
				// ! operator: constant if operand is constant in boolean context
				return isConstant(ctx, prefix.Operand, true)
			}
			return false
		case ast.KindPlusToken, ast.KindMinusToken, ast.KindTildeToken: // +, -, ~
			if prefix.Operand != nil {
				// Unary +, -, ~: constant if operand is constant
				// In boolean position, result is constant if operand is constant
				return isConstant(ctx, prefix.Operand, false)
			}
			return false
		}

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil {
			return false
		}

		operator := binary.OperatorToken.Kind

		// Comma operator (sequence expression): constant if right side is constant
		if operator == ast.KindCommaToken {
			return isConstant(ctx, binary.Right, inBooleanPosition)
		}

		// Assignment expressions
		if operator == ast.KindEqualsToken {
			// Simple assignment: constant if right side is constant
			return isConstant(ctx, binary.Right, inBooleanPosition)
		}

		// Logical assignment operators (||=, &&=)
		if operator == ast.KindBarBarEqualsToken || operator == ast.KindAmpersandAmpersandEqualsToken {
			// Logical assignments are only constant in boolean context when right side is identity
			if !inBooleanPosition {
				return false
			}
			var baseOp ast.Kind
			if operator == ast.KindBarBarEqualsToken {
				baseOp = ast.KindBarBarToken
			} else {
				baseOp = ast.KindAmpersandAmpersandToken
			}
			// The assignment is constant ONLY if the right side is a logical identity
			// For ||=: identity is true (1, "foo", etc - always truthy)
			// For &&=: identity is false (0, null, undefined, etc - always falsy)
			return isLogicalIdentity(binary.Right, baseOp)
		}

		// Logical operators (&&, ||, ??)
		switch operator {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			isLeftConstant := isConstant(ctx, binary.Left, inBooleanPosition)
			isRightConstant := isConstant(ctx, binary.Right, inBooleanPosition)
			isLeftShortCircuit := isLeftConstant && isLogicalIdentity(binary.Left, operator)
			isRightShortCircuit := inBooleanPosition && isRightConstant && isLogicalIdentity(binary.Right, operator)

			return (isLeftConstant && isRightConstant) || isLeftShortCircuit || isRightShortCircuit
		}

		// Comparison operators - both sides must be constant
		switch operator {
		case ast.KindLessThanToken,
			ast.KindLessThanEqualsToken,
			ast.KindGreaterThanToken,
			ast.KindGreaterThanEqualsToken,
			ast.KindEqualsEqualsToken,
			ast.KindExclamationEqualsToken,
			ast.KindEqualsEqualsEqualsToken,
			ast.KindExclamationEqualsEqualsToken:
			return isConstant(ctx, binary.Left, false) && isConstant(ctx, binary.Right, false)

		case ast.KindInKeyword:
			return false

		case ast.KindInstanceOfKeyword:
			return isConstant(ctx, binary.Left, false) && isConstant(ctx, binary.Right, false)
		}

		// Arithmetic operators - both sides must be constant
		switch operator {
		case ast.KindPlusToken,
			ast.KindMinusToken,
			ast.KindAsteriskToken,
			ast.KindSlashToken,
			ast.KindPercentToken,
			ast.KindAsteriskAsteriskToken,
			ast.KindLessThanLessThanToken,
			ast.KindGreaterThanGreaterThanToken,
			ast.KindGreaterThanGreaterThanGreaterThanToken,
			ast.KindBarToken,
			ast.KindAmpersandToken,
			ast.KindCaretToken:
			return isConstant(ctx, binary.Left, false) && isConstant(ctx, binary.Right, false)
		}

	case ast.KindNewExpression:
		return inBooleanPosition

	case ast.KindSpreadElement:
		return isConstant(ctx, node.AsSpreadElement().Expression, inBooleanPosition)

	case ast.KindCallExpression:
		if ast.IsOptionalChain(node) {
			return false
		}
		call := node.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		// Boolean's first argument alone controls its result, also when the
		// call is compared or coerced outside a boolean position.
		return callee != nil && callee.Kind == ast.KindIdentifier &&
			callee.Text() == "Boolean" &&
			(call.Arguments == nil || len(call.Arguments.Nodes) == 0 ||
				isConstant(ctx, call.Arguments.Nodes[0], true)) &&
			isReferenceToGlobal(ctx, callee)

	case ast.KindPostfixUnaryExpression:
		// ++ and -- are not constant (they modify variables)
		return false
	}

	return false
}

// isReferenceToGlobal combines lexical binding identity with the effective
// language globals. RefStore works in source-only Programs too.
func isReferenceToGlobal(ctx *rule.RuleContext, node *ast.Node) bool {
	if ctx == nil || !ctx.Globals.Access(node.Text()).IsDeclared() {
		return false
	}
	if ctx.Refs != nil {
		return ctx.Refs.IsGlobalReference(node)
	}
	return !utils.IsShadowed(node, node.Text())
}

// isWhileTrueLoop checks if a loop is a `while (true)` loop
func isWhileTrueLoop(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindWhileStatement {
		return false
	}

	whileStmt := node.AsWhileStatement()
	if whileStmt == nil || whileStmt.Expression == nil {
		return false
	}

	// Check if the condition is literally `true`
	if whileStmt.Expression.Kind == ast.KindTrueKeyword {
		return true
	}

	// Check for parenthesized true: (true), ((true)), etc.
	expr := whileStmt.Expression
	for expr.Kind == ast.KindParenthesizedExpression {
		paren := expr.AsParenthesizedExpression()
		if paren == nil || paren.Expression == nil {
			return false
		}
		expr = paren.Expression
	}

	return expr.Kind == ast.KindTrueKeyword
}

// containsYield checks if a node contains a yield expression (not inside nested functions)
func containsYield(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// If this is a yield expression, return true
	if node.Kind == ast.KindYieldExpression {
		return true
	}

	// Don't traverse into nested function bodies
	switch node.Kind {
	case ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindArrowFunction:
		return false
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindConstructor:
		// ESTree visits a computed key outside the method's FunctionExpression,
		// but yields in its body belong to that function's own loop set.
		return containsYield(node.Name())
	}

	// Recursively check children using ForEachChild
	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if containsYield(child) {
			found = true
			return true // Stop iteration
		}
		return false // Continue iteration
	})

	return found
}

// shouldCheckLoop determines if a loop should be checked based on options
func shouldCheckLoop(node *ast.Node, opts Options) bool {
	if opts.CheckLoops == "none" {
		return false
	}

	if opts.CheckLoops == "allExceptWhileTrue" {
		// Allow while (true) loops
		if isWhileTrueLoop(node) {
			return false
		}
	}

	// Yields encountered after a loop is tracked clear it upstream. A for
	// initializer runs before its test is re-tracked and must be excluded.
	// Checking the entire loop body also catches yields in nested loop headers.
	var body, condition, incrementor *ast.Node
	switch node.Kind {
	case ast.KindWhileStatement:
		stmt := node.AsWhileStatement()
		body, condition = stmt.Statement, stmt.Expression
	case ast.KindDoStatement:
		stmt := node.AsDoStatement()
		body, condition = stmt.Statement, stmt.Expression
	case ast.KindForStatement:
		stmt := node.AsForStatement()
		body, condition, incrementor = stmt.Statement, stmt.Condition, stmt.Incrementor
	}
	if containsYield(body) || containsYield(condition) || containsYield(incrementor) {
		return false
	}

	return true
}

// getTestExpression returns the condition/test expression for a statement
func getTestExpression(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}

	switch node.Kind {
	case ast.KindIfStatement:
		ifStmt := node.AsIfStatement()
		if ifStmt != nil {
			return ifStmt.Expression
		}
	case ast.KindWhileStatement:
		whileStmt := node.AsWhileStatement()
		if whileStmt != nil {
			return whileStmt.Expression
		}
	case ast.KindDoStatement:
		doStmt := node.AsDoStatement()
		if doStmt != nil {
			return doStmt.Expression
		}
	case ast.KindForStatement:
		forStmt := node.AsForStatement()
		if forStmt != nil {
			return forStmt.Condition
		}
	case ast.KindConditionalExpression:
		condExpr := node.AsConditionalExpression()
		if condExpr != nil {
			return condExpr.Condition
		}
	}

	return nil
}

// checkCondition reports if a condition is constant
func checkCondition(ctx *rule.RuleContext, node *ast.Node, opts Options) {
	testExpr := getTestExpression(node)
	if testExpr == nil {
		return
	}
	testExpr = ast.SkipParentheses(testExpr)

	// For loops, check if we should skip based on options
	switch node.Kind {
	case ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement:
		if !shouldCheckLoop(node, opts) {
			return
		}
	}

	// Check if the test expression is constant
	if isConstant(ctx, testExpr, true) {
		ctx.ReportNode(testExpr, buildUnexpectedMessage())
	}
}

// NoConstantConditionRule disallows constant expressions in conditions
var NoConstantConditionRule = rule.Rule{
	Name:   "no-constant-condition",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			// Check if statements
			ast.KindIfStatement: func(node *ast.Node) {
				checkCondition(&ctx, node, opts)
			},

			// Check conditional expressions (ternary)
			ast.KindConditionalExpression: func(node *ast.Node) {
				checkCondition(&ctx, node, opts)
			},

			// Check while loops
			ast.KindWhileStatement: func(node *ast.Node) {
				checkCondition(&ctx, node, opts)
			},

			// Check do-while loops
			ast.KindDoStatement: func(node *ast.Node) {
				checkCondition(&ctx, node, opts)
			},

			// Check for loops
			ast.KindForStatement: func(node *ast.Node) {
				checkCondition(&ctx, node, opts)
			},
		}
	},
}
