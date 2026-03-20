package no_constant_binary_expression

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Message builders
func buildConstantBinaryOperandMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "constantBinaryOperand",
		Description: "Unexpected constant binary expression. Comparisons will always evaluate the same.",
	}
}

func buildConstantShortCircuitMessage(property string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "constantShortCircuit",
		Description: "Unexpected constant " + property + " on the left-hand side of a `" + property + "` expression.",
	}
}

func buildAlwaysNewMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "alwaysNew",
		Description: "Unexpected comparison to newly constructed object. These two values can never be equal.",
	}
}

func buildBothAlwaysNewMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "bothAlwaysNew",
		Description: "Unexpected comparison of two newly constructed objects. These two values can never be equal.",
	}
}

// --- Operator classification helpers ---

func isNumericOrStringBinaryOperator(op ast.Kind) bool {
	switch op {
	case ast.KindPlusToken, ast.KindMinusToken, ast.KindAsteriskToken,
		ast.KindSlashToken, ast.KindPercentToken, ast.KindAsteriskAsteriskToken,
		ast.KindLessThanLessThanToken, ast.KindGreaterThanGreaterThanToken,
		ast.KindGreaterThanGreaterThanGreaterThanToken,
		ast.KindBarToken, ast.KindAmpersandToken, ast.KindCaretToken:
		return true
	}
	return false
}

func isLogicalAssignmentOperator(op ast.Kind) bool {
	switch op {
	case ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken,
		ast.KindQuestionQuestionEqualsToken:
		return true
	}
	return false
}

func isCompoundAssignmentOperator(op ast.Kind) bool {
	switch op {
	case ast.KindPlusEqualsToken, ast.KindMinusEqualsToken,
		ast.KindAsteriskEqualsToken, ast.KindSlashEqualsToken,
		ast.KindPercentEqualsToken, ast.KindAsteriskAsteriskEqualsToken,
		ast.KindLessThanLessThanEqualsToken, ast.KindGreaterThanGreaterThanEqualsToken,
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken,
		ast.KindBarEqualsToken, ast.KindAmpersandEqualsToken,
		ast.KindCaretEqualsToken:
		return true
	}
	return false
}

// --- Core helpers ---

// isNullOrUndefined checks if a node represents null or undefined
func isNullOrUndefined(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindNullKeyword:
		return true
	case ast.KindIdentifier:
		return node.Text() == "undefined"
	case ast.KindVoidExpression:
		return true
	case ast.KindParenthesizedExpression:
		paren := node.AsParenthesizedExpression()
		return paren != nil && isNullOrUndefined(paren.Expression)
	}
	return false
}

// isGlobalBuiltin checks if an identifier refers to a global built-in (not shadowed)
func isGlobalBuiltin(ctx *rule.RuleContext, node *ast.Node) bool {
	if ctx == nil || ctx.TypeChecker == nil || ctx.Program == nil || ctx.SourceFile == nil {
		return false
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
	if symbol == nil {
		return false
	}
	// Check if any declaration is from the current source file (locally shadowed)
	for _, declaration := range symbol.Declarations {
		declarationSourceFile := ast.GetSourceFileOfNode(declaration)
		if declarationSourceFile != nil && declarationSourceFile == ctx.SourceFile {
			return false
		}
	}
	// TypeScript-Go's checker may merge function declarations with same-named global
	// interfaces, causing the symbol to resolve to the global even when shadowed.
	// Manually check for top-level function declarations with the same name.
	if node.Kind == ast.KindIdentifier {
		if hasTopLevelFunctionDeclaration(ctx.SourceFile, node.Text()) {
			return false
		}
	}
	return utils.IsSymbolFromDefaultLibrary(ctx.Program, symbol)
}

// hasTopLevelFunctionDeclaration checks if the source file has a top-level function
// declaration with the given name (used to detect shadowing of global built-ins).
func hasTopLevelFunctionDeclaration(sf *ast.SourceFile, name string) bool {
	if sf == nil || sf.Statements == nil {
		return false
	}
	for _, stmt := range sf.Statements.Nodes {
		if stmt.Kind == ast.KindFunctionDeclaration {
			nameNode := stmt.Name()
			if nameNode != nil && nameNode.Text() == name {
				return true
			}
		}
	}
	return false
}

// getBooleanValue returns the boolean value of a literal, or nil if unknown
func getBooleanValue(node *ast.Node) *bool {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindTrueKeyword:
		t := true
		return &t
	case ast.KindFalseKeyword:
		f := false
		return &f
	case ast.KindNullKeyword:
		f := false
		return &f
	case ast.KindNumericLiteral:
		text := node.Text()
		if text == "0" || text == "0.0" || text == "-0" {
			f := false
			return &f
		}
		t := true
		return &t
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		// node.Text() returns the parsed value without quotes/backticks
		text := node.Text()
		if len(text) == 0 {
			f := false
			return &f
		}
		t := true
		return &t
	case ast.KindIdentifier:
		if node.Text() == "undefined" {
			f := false
			return &f
		}
	}
	return nil
}

// isLogicalIdentity checks if a node is a logical identity element for the given operator
func isLogicalIdentity(node *ast.Node, operator ast.Kind) bool {
	if node == nil {
		return false
	}

	boolVal := getBooleanValue(node)
	if boolVal != nil {
		if operator == ast.KindBarBarToken && *boolVal {
			return true
		}
		if operator == ast.KindAmpersandAmpersandToken && !*boolVal {
			return true
		}
	}

	// For ||, always-truthy values are identity
	if operator == ast.KindBarBarToken {
		switch node.Kind {
		case ast.KindRegularExpressionLiteral,
			ast.KindArrowFunction,
			ast.KindFunctionExpression,
			ast.KindClassExpression,
			ast.KindObjectLiteralExpression,
			ast.KindArrayLiteralExpression:
			return true
		}
	}

	// void operator is identity for &&
	if node.Kind == ast.KindVoidExpression {
		return operator == ast.KindAmpersandAmpersandToken
	}

	if node.Kind == ast.KindParenthesizedExpression {
		paren := node.AsParenthesizedExpression()
		if paren != nil && paren.Expression != nil {
			return isLogicalIdentity(paren.Expression, operator)
		}
	}

	if node.Kind == ast.KindBinaryExpression {
		binary := node.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil {
			nodeOp := binary.OperatorToken.Kind
			// Logical expressions with same operator
			if nodeOp == operator && (nodeOp == ast.KindBarBarToken || nodeOp == ast.KindAmpersandAmpersandToken) {
				return isLogicalIdentity(binary.Left, operator) || isLogicalIdentity(binary.Right, operator)
			}
			// Logical assignment operators
			if nodeOp == ast.KindBarBarEqualsToken || nodeOp == ast.KindAmpersandAmpersandEqualsToken {
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

// --- Main analysis functions ---

// isConstant checks if a node represents a constant value.
// inBooleanPosition indicates whether the value is used in a boolean context (e.g., left side of && or ||).
func isConstant(ctx *rule.RuleContext, node *ast.Node, inBooleanPosition bool) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindNumericLiteral, ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword,
		ast.KindRegularExpressionLiteral, ast.KindBigIntLiteral:
		return true

	case ast.KindIdentifier:
		return node.Text() == "undefined"

	case ast.KindArrowFunction, ast.KindFunctionExpression,
		ast.KindClassExpression, ast.KindObjectLiteralExpression:
		return true

	case ast.KindArrayLiteralExpression:
		if inBooleanPosition {
			return true
		}
		arrayLit := node.AsArrayLiteralExpression()
		if arrayLit != nil && arrayLit.Elements != nil {
			for _, elem := range arrayLit.Elements.Nodes {
				if elem.Kind == ast.KindOmittedExpression {
					continue
				}
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
		template := node.AsTemplateExpression()
		if template == nil {
			return false
		}
		if inBooleanPosition {
			// Constant if any static part has content (non-empty string is truthy)
			if template.Head != nil && len(template.Head.Text()) > 0 {
				return true
			}
			if template.TemplateSpans != nil {
				for _, span := range template.TemplateSpans.Nodes {
					if span.Kind == ast.KindTemplateSpan {
						templateSpan := span.AsTemplateSpan()
						if templateSpan != nil && templateSpan.Literal != nil && len(templateSpan.Literal.Text()) > 0 {
							return true
						}
					}
				}
			}
		}
		// Not in boolean position (or no static content): constant only if all expressions are constant
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

	case ast.KindVoidExpression:
		return true

	case ast.KindTypeOfExpression:
		if inBooleanPosition {
			return true
		}
		typeofExpr := node.AsTypeOfExpression()
		if typeofExpr != nil && typeofExpr.Expression != nil {
			return isConstant(ctx, typeofExpr.Expression, false)
		}
		return false

	case ast.KindDeleteExpression:
		deleteExpr := node.AsDeleteExpression()
		if deleteExpr != nil && deleteExpr.Expression != nil {
			return isConstant(ctx, deleteExpr.Expression, false)
		}
		return false

	case ast.KindPrefixUnaryExpression:
		prefix := node.AsPrefixUnaryExpression()
		if prefix == nil {
			return false
		}
		switch prefix.Operator {
		case ast.KindExclamationToken:
			return isConstant(ctx, prefix.Operand, true)
		case ast.KindPlusToken, ast.KindMinusToken, ast.KindTildeToken:
			return isConstant(ctx, prefix.Operand, false)
		case ast.KindPlusPlusToken, ast.KindMinusMinusToken:
			// Prefix ++ / -- modify a variable, not constant
			return false
		}

	case ast.KindPostfixUnaryExpression:
		return false

	case ast.KindNewExpression:
		return inBooleanPosition

	case ast.KindParenthesizedExpression:
		paren := node.AsParenthesizedExpression()
		if paren != nil && paren.Expression != nil {
			return isConstant(ctx, paren.Expression, inBooleanPosition)
		}

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil {
			return false
		}
		op := binary.OperatorToken.Kind

		// Comma operator (sequence expression): constant if last (right) is constant
		if op == ast.KindCommaToken {
			return isConstant(ctx, binary.Right, inBooleanPosition)
		}

		// Simple assignment: constant if right side is constant
		if op == ast.KindEqualsToken {
			return isConstant(ctx, binary.Right, inBooleanPosition)
		}

		// Logical assignment (||=, &&=)
		if op == ast.KindBarBarEqualsToken || op == ast.KindAmpersandAmpersandEqualsToken {
			if !inBooleanPosition {
				return false
			}
			var baseOp ast.Kind
			if op == ast.KindBarBarEqualsToken {
				baseOp = ast.KindBarBarToken
			} else {
				baseOp = ast.KindAmpersandAmpersandToken
			}
			return isLogicalIdentity(binary.Right, baseOp)
		}

		// Other compound/logical assignments
		if isLogicalAssignmentOperator(op) || isCompoundAssignmentOperator(op) {
			return false
		}

		// Logical operators (&&, ||, ??)
		switch op {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			isLeftConstant := isConstant(ctx, binary.Left, inBooleanPosition)
			isRightConstant := isConstant(ctx, binary.Right, inBooleanPosition)
			isLeftShortCircuit := isLeftConstant && isLogicalIdentity(binary.Left, op)
			isRightShortCircuit := inBooleanPosition && isRightConstant && isLogicalIdentity(binary.Right, op)
			return (isLeftConstant && isRightConstant) || isLeftShortCircuit || isRightShortCircuit
		}

		// Arithmetic / comparison / bitwise operators (both sides must be constant)
		if isNumericOrStringBinaryOperator(op) {
			return isConstant(ctx, binary.Left, false) && isConstant(ctx, binary.Right, false)
		}

		// Comparison operators
		switch op {
		case ast.KindLessThanToken, ast.KindLessThanEqualsToken,
			ast.KindGreaterThanToken, ast.KindGreaterThanEqualsToken,
			ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken,
			ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken,
			ast.KindInstanceOfKeyword:
			return isConstant(ctx, binary.Left, false) && isConstant(ctx, binary.Right, false)
		case ast.KindInKeyword:
			return false
		}

	case ast.KindCommaListExpression:
		children := node.Children()
		if children != nil && len(children.Nodes) > 0 {
			return isConstant(ctx, children.Nodes[len(children.Nodes)-1], inBooleanPosition)
		}

	case ast.KindSpreadElement:
		spread := node.AsSpreadElement()
		if spread != nil && spread.Expression != nil {
			return isConstant(ctx, spread.Expression, inBooleanPosition)
		}

	case ast.KindCallExpression:
		callExpr := node.AsCallExpression()
		if callExpr == nil || callExpr.Expression == nil {
			return false
		}
		if callExpr.Expression.Kind != ast.KindIdentifier {
			return false
		}
		name := callExpr.Expression.Text()
		if name != "Boolean" {
			return false
		}
		if !isGlobalBuiltin(ctx, callExpr.Expression) {
			return false
		}
		if callExpr.Arguments == nil || len(callExpr.Arguments.Nodes) == 0 {
			return true
		}
		return isConstant(ctx, callExpr.Arguments.Nodes[0], true)
	}

	return false
}

// isStaticBoolean checks if a node is a static (unchanging) boolean value
func isStaticBoolean(ctx *rule.RuleContext, node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true
	case ast.KindParenthesizedExpression:
		paren := node.AsParenthesizedExpression()
		return paren != nil && isStaticBoolean(ctx, paren.Expression)
	case ast.KindPrefixUnaryExpression:
		prefix := node.AsPrefixUnaryExpression()
		if prefix != nil && prefix.Operator == ast.KindExclamationToken {
			return isConstant(ctx, prefix.Operand, true)
		}
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		if call == nil || call.Expression == nil || call.Expression.Kind != ast.KindIdentifier {
			return false
		}
		if call.Expression.Text() != "Boolean" {
			return false
		}
		if !isGlobalBuiltin(ctx, call.Expression) {
			return false
		}
		if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
			return true
		}
		return isConstant(ctx, call.Arguments.Nodes[0], true)
	}
	return false
}

// isAlwaysNew checks if an expression always creates a new object
func isAlwaysNew(ctx *rule.RuleContext, node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression,
		ast.KindArrowFunction, ast.KindFunctionExpression,
		ast.KindClassExpression, ast.KindRegularExpressionLiteral:
		return true

	case ast.KindNewExpression:
		// Only built-in constructors are guaranteed to always return new objects.
		// User-defined constructors could return a sentinel object.
		newExpr := node.AsNewExpression()
		if newExpr == nil || newExpr.Expression == nil {
			return false
		}
		if newExpr.Expression.Kind != ast.KindIdentifier {
			return false
		}
		return isGlobalBuiltin(ctx, newExpr.Expression)

	case ast.KindParenthesizedExpression:
		paren := node.AsParenthesizedExpression()
		return paren != nil && isAlwaysNew(ctx, paren.Expression)

	case ast.KindConditionalExpression:
		cond := node.AsConditionalExpression()
		if cond != nil {
			return isAlwaysNew(ctx, cond.WhenTrue) && isAlwaysNew(ctx, cond.WhenFalse)
		}

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil {
			return false
		}
		op := binary.OperatorToken.Kind
		if op == ast.KindCommaToken {
			return isAlwaysNew(ctx, binary.Right)
		}
		if op == ast.KindEqualsToken {
			return isAlwaysNew(ctx, binary.Right)
		}

	case ast.KindCommaListExpression:
		children := node.Children()
		if children != nil && len(children.Nodes) > 0 {
			return isAlwaysNew(ctx, children.Nodes[len(children.Nodes)-1])
		}
	}

	return false
}

// hasConstantNullishness checks if a node always resolves to a constant nullish or non-nullish value.
// When nonNullish is true, nullish values are NOT considered constant (used for ?? right-side checks).
func hasConstantNullishness(ctx *rule.RuleContext, node *ast.Node, nonNullish bool) bool {
	if node == nil {
		return false
	}

	if nonNullish && isNullOrUndefined(node) {
		return false
	}

	switch node.Kind {
	// Always non-nullish
	case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression,
		ast.KindArrowFunction, ast.KindFunctionExpression,
		ast.KindClassExpression, ast.KindNewExpression,
		ast.KindRegularExpressionLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true

	// Literal values: nullish (null) or non-nullish (numbers, strings) — all constant
	case ast.KindNullKeyword, ast.KindNumericLiteral, ast.KindStringLiteral,
		ast.KindNoSubstitutionTemplateLiteral, ast.KindBigIntLiteral:
		return true

	case ast.KindTemplateExpression:
		return true // strings are never nullish

	case ast.KindIdentifier:
		return node.Text() == "undefined"

	case ast.KindVoidExpression:
		return true // always undefined (nullish, but constantly so)

	case ast.KindTypeOfExpression:
		return true // always returns a string (non-nullish)

	case ast.KindDeleteExpression:
		return true // always returns a boolean (non-nullish)

	case ast.KindPrefixUnaryExpression:
		// !, +, -, ~, ++, -- all produce non-nullish values (boolean or number)
		return true

	case ast.KindPostfixUnaryExpression:
		return true // ++/-- produce numbers (non-nullish)

	case ast.KindParenthesizedExpression:
		paren := node.AsParenthesizedExpression()
		return paren != nil && hasConstantNullishness(ctx, paren.Expression, nonNullish)

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil {
			return false
		}
		op := binary.OperatorToken.Kind

		// Arithmetic/bitwise operators always produce non-nullish values
		if isNumericOrStringBinaryOperator(op) {
			return true
		}

		// Nullish coalescing: a ?? b has constant nullishness if b is constantly non-nullish
		if op == ast.KindQuestionQuestionToken {
			return hasConstantNullishness(ctx, binary.Right, true)
		}

		// Comma: check last expression
		if op == ast.KindCommaToken {
			return hasConstantNullishness(ctx, binary.Right, nonNullish)
		}

		// Simple assignment: check right side
		if op == ast.KindEqualsToken {
			return hasConstantNullishness(ctx, binary.Right, nonNullish)
		}

		// Logical assignment: can't determine without scope walking
		if isLogicalAssignmentOperator(op) {
			return false
		}

		// Compound assignments (+=, -=, etc.) always produce numeric/string (non-nullish)
		if isCompoundAssignmentOperator(op) {
			return true
		}

		// All remaining binary operators (comparison, arithmetic, bitwise, instanceof, in)
		// produce non-nullish values (numbers, booleans, or strings)
		return true

	case ast.KindCommaListExpression:
		children := node.Children()
		if children != nil && len(children.Nodes) > 0 {
			return hasConstantNullishness(ctx, children.Nodes[len(children.Nodes)-1], nonNullish)
		}

	case ast.KindCallExpression:
		call := node.AsCallExpression()
		if call == nil || call.Expression == nil || call.Expression.Kind != ast.KindIdentifier {
			return false
		}
		name := call.Expression.Text()
		// Boolean(), String(), Number() always return non-nullish values
		if name == "Boolean" || name == "String" || name == "Number" {
			return isGlobalBuiltin(ctx, call.Expression)
		}
	}

	return false
}

// hasConstantLooseBooleanComparison checks if a value always gives the same result
// when compared loosely (==) to a boolean value.
func hasConstantLooseBooleanComparison(ctx *rule.RuleContext, node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindObjectLiteralExpression, ast.KindClassExpression:
		return true

	case ast.KindArrayLiteralExpression:
		// Empty arrays coerce to 0 (constant). Multi-element arrays coerce to NaN (constant).
		// Single-element arrays depend on the element value (not constant).
		arrayLit := node.AsArrayLiteralExpression()
		if arrayLit == nil || arrayLit.Elements == nil {
			return true // empty
		}
		elements := arrayLit.Elements.Nodes
		if len(elements) == 0 {
			return true
		}
		nonSpreadCount := 0
		for _, elem := range elements {
			if elem.Kind != ast.KindOmittedExpression && elem.Kind != ast.KindSpreadElement {
				nonSpreadCount++
			}
		}
		return nonSpreadCount > 1

	case ast.KindArrowFunction, ast.KindFunctionExpression:
		return true

	case ast.KindVoidExpression:
		return true // always undefined

	case ast.KindTypeOfExpression:
		return true // typeof strings, when coerced to number, are not 0 or 1

	case ast.KindPrefixUnaryExpression:
		prefix := node.AsPrefixUnaryExpression()
		if prefix == nil {
			return false
		}
		if prefix.Operator == ast.KindExclamationToken {
			return isConstant(ctx, prefix.Operand, true)
		}
		// +, -, ~, ++, -- : we won't try to reason about these
		return false

	case ast.KindNewExpression:
		return false // objects might have custom valueOf/toString

	case ast.KindNumericLiteral, ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword,
		ast.KindRegularExpressionLiteral, ast.KindBigIntLiteral:
		return true

	case ast.KindIdentifier:
		return node.Text() == "undefined"

	case ast.KindParenthesizedExpression:
		paren := node.AsParenthesizedExpression()
		return paren != nil && hasConstantLooseBooleanComparison(ctx, paren.Expression)

	case ast.KindTemplateExpression:
		// Template with expressions: result varies depending on expressions
		return false

	case ast.KindCallExpression:
		call := node.AsCallExpression()
		if call == nil || call.Expression == nil || call.Expression.Kind != ast.KindIdentifier {
			return false
		}
		if call.Expression.Text() != "Boolean" {
			return false
		}
		if !isGlobalBuiltin(ctx, call.Expression) {
			return false
		}
		if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
			return true
		}
		return isConstant(ctx, call.Arguments.Nodes[0], true)

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil {
			return false
		}
		op := binary.OperatorToken.Kind
		if op == ast.KindCommaToken {
			return hasConstantLooseBooleanComparison(ctx, binary.Right)
		}
		if op == ast.KindEqualsToken {
			return hasConstantLooseBooleanComparison(ctx, binary.Right)
		}

	case ast.KindCommaListExpression:
		children := node.Children()
		if children != nil && len(children.Nodes) > 0 {
			return hasConstantLooseBooleanComparison(ctx, children.Nodes[len(children.Nodes)-1])
		}
	}

	return false
}

// hasConstantStrictBooleanComparison checks if a value always gives the same result
// when strictly compared (===) to a boolean value.
func hasConstantStrictBooleanComparison(ctx *rule.RuleContext, node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	// Non-boolean types: strict comparison to boolean is always false (constant result)
	case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression,
		ast.KindArrowFunction, ast.KindFunctionExpression,
		ast.KindClassExpression, ast.KindNewExpression,
		ast.KindRegularExpressionLiteral,
		ast.KindNumericLiteral, ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindBigIntLiteral,
		ast.KindNullKeyword:
		return true

	// Boolean literals: strict comparison is constant
	case ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true

	case ast.KindTemplateExpression:
		return true // strings are not booleans

	case ast.KindIdentifier:
		return node.Text() == "undefined"

	case ast.KindVoidExpression:
		return true // undefined is not boolean

	case ast.KindTypeOfExpression:
		return true // typeof returns string, not boolean

	case ast.KindDeleteExpression:
		return false // delete returns boolean, result varies

	case ast.KindPrefixUnaryExpression:
		prefix := node.AsPrefixUnaryExpression()
		if prefix == nil {
			return false
		}
		if prefix.Operator == ast.KindExclamationToken {
			return isConstant(ctx, prefix.Operand, true)
		}
		// +, -, ~: return numbers (not booleans)
		// ++, --: return numbers (not booleans)
		return true

	case ast.KindPostfixUnaryExpression:
		return true // numbers are not booleans

	case ast.KindParenthesizedExpression:
		paren := node.AsParenthesizedExpression()
		return paren != nil && hasConstantStrictBooleanComparison(ctx, paren.Expression)

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil {
			return false
		}
		op := binary.OperatorToken.Kind

		// Numeric/string operators produce non-boolean values
		if isNumericOrStringBinaryOperator(op) {
			return true
		}

		// Comma: check last
		if op == ast.KindCommaToken {
			return hasConstantStrictBooleanComparison(ctx, binary.Right)
		}

		// Assignment =: check right
		if op == ast.KindEqualsToken {
			return hasConstantStrictBooleanComparison(ctx, binary.Right)
		}

		// Logical assignment: can't determine
		if isLogicalAssignmentOperator(op) {
			return false
		}

		// Compound assignments produce numeric/string (not boolean)
		if isCompoundAssignmentOperator(op) {
			return true
		}

	case ast.KindCommaListExpression:
		children := node.Children()
		if children != nil && len(children.Nodes) > 0 {
			return hasConstantStrictBooleanComparison(ctx, children.Nodes[len(children.Nodes)-1])
		}

	case ast.KindCallExpression:
		call := node.AsCallExpression()
		if call == nil || call.Expression == nil || call.Expression.Kind != ast.KindIdentifier {
			return false
		}
		name := call.Expression.Text()
		// String() and Number() always return non-boolean types
		if name == "String" || name == "Number" {
			return isGlobalBuiltin(ctx, call.Expression)
		}
		// Boolean() returns a boolean — constant only if argument is constant
		if name == "Boolean" {
			if !isGlobalBuiltin(ctx, call.Expression) {
				return false
			}
			if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
				return true
			}
			return isConstant(ctx, call.Arguments.Nodes[0], true)
		}
	}

	return false
}

// findBinaryExpressionConstantOperand checks if operand `a` being a specific type (null/boolean)
// makes the comparison with operand `b` always constant. Returns `b` if constant, nil otherwise.
func findBinaryExpressionConstantOperand(ctx *rule.RuleContext, a, b *ast.Node, operator ast.Kind) *ast.Node {
	switch operator {
	case ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken:
		if (isNullOrUndefined(a) && hasConstantNullishness(ctx, b, false)) ||
			(isStaticBoolean(ctx, a) && hasConstantLooseBooleanComparison(ctx, b)) {
			return b
		}
	case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
		if (isNullOrUndefined(a) && hasConstantNullishness(ctx, b, false)) ||
			(isStaticBoolean(ctx, a) && hasConstantStrictBooleanComparison(ctx, b)) {
			return b
		}
	}
	return nil
}

// NoConstantBinaryExpressionRule detects constant binary expressions
var NoConstantBinaryExpressionRule = rule.Rule{
	Name: "no-constant-binary-expression",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary == nil || binary.OperatorToken == nil {
					return
				}

				operator := binary.OperatorToken.Kind

				// 1. Logical operators: && and ||
				if operator == ast.KindAmpersandAmpersandToken || operator == ast.KindBarBarToken {
					if isConstant(&ctx, binary.Left, true) {
						var prop string
						if operator == ast.KindAmpersandAmpersandToken {
							prop = "&&"
						} else {
							prop = "||"
						}
						ctx.ReportNode(node, buildConstantShortCircuitMessage(prop))
					}
					return
				}

				// 2. Nullish coalescing: ??
				if operator == ast.KindQuestionQuestionToken {
					if hasConstantNullishness(&ctx, binary.Left, false) {
						ctx.ReportNode(node, buildConstantShortCircuitMessage("??"))
					}
					return
				}

				// 3. Equality operators
				switch operator {
				case ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken,
					ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:

					// Check constantBinaryOperand FIRST (in both directions)
					rightConstant := findBinaryExpressionConstantOperand(&ctx, binary.Left, binary.Right, operator)
					leftConstant := findBinaryExpressionConstantOperand(&ctx, binary.Right, binary.Left, operator)

					if rightConstant != nil {
						ctx.ReportNode(node, buildConstantBinaryOperandMessage())
					} else if leftConstant != nil {
						ctx.ReportNode(node, buildConstantBinaryOperandMessage())
					} else if operator == ast.KindEqualsEqualsEqualsToken || operator == ast.KindExclamationEqualsEqualsToken {
						// For === / !==: single-side isAlwaysNew is enough
						if isAlwaysNew(&ctx, binary.Left) {
							ctx.ReportNode(node, buildAlwaysNewMessage())
						} else if isAlwaysNew(&ctx, binary.Right) {
							ctx.ReportNode(node, buildAlwaysNewMessage())
						}
					} else {
						// For == / !=: BOTH sides must be isAlwaysNew
						if isAlwaysNew(&ctx, binary.Left) && isAlwaysNew(&ctx, binary.Right) {
							ctx.ReportNode(node, buildBothAlwaysNewMessage())
						}
					}
				}
			},
		}
	},
}
