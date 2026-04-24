package no_magic_numbers

import (
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Maximum array length by the ECMAScript Specification.
const maxArrayLength = 1<<32 - 1

type options struct {
	detectObjects                 bool
	enforceConst                  bool
	ignore                        map[string]bool // normalized values as strings
	ignoreArrayIndexes            bool
	ignoreDefaultValues           bool
	ignoreClassFieldInitialValues bool
	ignoreEnums                   bool
	ignoreNumericLiteralTypes     bool
	ignoreReadonlyClassProperties bool
	ignoreTypeIndexes             bool
}

func parseOptions(rawOpts any) options {
	opts := options{
		ignore: make(map[string]bool),
	}
	optsMap := utils.GetOptionsMap(rawOpts)
	if optsMap == nil {
		return opts
	}
	if v, ok := optsMap["detectObjects"].(bool); ok {
		opts.detectObjects = v
	}
	if v, ok := optsMap["enforceConst"].(bool); ok {
		opts.enforceConst = v
	}
	if v, ok := optsMap["ignoreArrayIndexes"].(bool); ok {
		opts.ignoreArrayIndexes = v
	}
	if v, ok := optsMap["ignoreDefaultValues"].(bool); ok {
		opts.ignoreDefaultValues = v
	}
	if v, ok := optsMap["ignoreClassFieldInitialValues"].(bool); ok {
		opts.ignoreClassFieldInitialValues = v
	}
	if v, ok := optsMap["ignoreEnums"].(bool); ok {
		opts.ignoreEnums = v
	}
	if v, ok := optsMap["ignoreNumericLiteralTypes"].(bool); ok {
		opts.ignoreNumericLiteralTypes = v
	}
	if v, ok := optsMap["ignoreReadonlyClassProperties"].(bool); ok {
		opts.ignoreReadonlyClassProperties = v
	}
	if v, ok := optsMap["ignoreTypeIndexes"].(bool); ok {
		opts.ignoreTypeIndexes = v
	}
	if arr, ok := optsMap["ignore"].([]interface{}); ok {
		for _, item := range arr {
			switch v := item.(type) {
			case float64:
				opts.ignore[normalizeIgnoreFloat(v)] = true
			case string:
				// BigInt string like "100n" or "-100n"
				opts.ignore[normalizeIgnoreBigIntString(v)] = true
			}
		}
	}
	return opts
}

// normalizeIgnoreFloat converts a float64 ignore value to a canonical string key.
// Sign is preserved: -2 → "-2", 2 → "2".
func normalizeIgnoreFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// normalizeIgnoreBigIntString converts a bigint ignore string like "100n" or "-100n"
// to a canonical key: "bigint:100" or "bigint:-100".
func normalizeIgnoreBigIntString(s string) string {
	negative := false
	text := s
	if strings.HasPrefix(text, "-") {
		negative = true
		text = text[1:]
	} else if strings.HasPrefix(text, "+") {
		text = text[1:]
	}
	text = strings.TrimSuffix(text, "n")
	n, ok := new(big.Int).SetString(text, 10)
	if !ok {
		return "bigint:" + s
	}
	if negative {
		n.Neg(n)
	}
	return "bigint:" + n.String()
}

// parseRawNumericValue parses a raw numeric literal string (which may be hex, octal, binary)
// into a float64. Unlike strconv.ParseFloat, this handles all JS numeric prefixes.
func parseRawNumericValue(raw string) (float64, bool) {
	if len(raw) > 2 && raw[0] == '0' {
		switch raw[1] {
		case 'x', 'X', 'o', 'O', 'b', 'B':
			n, ok := new(big.Int).SetString(raw, 0)
			if !ok {
				return 0, false
			}
			f, _ := new(big.Float).SetInt(n).Float64()
			return f, true
		}
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil && !math.IsInf(f, 0) {
		return 0, false
	}
	return f, true
}

// normalizeLiteralValue returns a canonical string key for a numeric or bigint literal,
// handling unary minus/plus prefix.
func normalizeLiteralValue(node *ast.Node, raw string, isBigInt bool) string {
	negate := false
	if node.Parent != nil && node.Parent.Kind == ast.KindPrefixUnaryExpression {
		pref := node.Parent.AsPrefixUnaryExpression()
		if pref.Operator == ast.KindMinusToken {
			negate = true
		}
	}
	if isBigInt {
		text := strings.TrimSuffix(raw, "n")
		n, ok := new(big.Int).SetString(text, 0)
		if !ok {
			key := "bigint:" + raw
			if negate {
				key = "bigint:-" + raw
			}
			return key
		}
		if negate {
			n.Neg(n)
		}
		return "bigint:" + n.String()
	}
	// Numeric literal
	f, ok := parseRawNumericValue(raw)
	if !ok {
		return raw
	}
	if negate {
		f = -f
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

var noMagicMessage = rule.RuleMessage{
	Id:          "noMagic",
	Description: "No magic number: {{raw}}.",
}

var useConstMessage = rule.RuleMessage{
	Id:          "useConst",
	Description: "Number constants declarations must use 'const'.",
}

// NoMagicNumbersRule implements the @typescript-eslint/no-magic-numbers rule.
var NoMagicNumbersRule = rule.CreateRule(rule.Rule{
	Name: "no-magic-numbers",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)

		// Types allowed when detectObjects is false.
		// In ESLint: ObjectExpression, Property, AssignmentExpression (but NOT assignment to an identifier)
		// When detectObjects is true, none of these are allowed.

		handleNumericNode := func(node *ast.Node, isBigInt bool) {
			raw := utils.TrimmedNodeText(ctx.SourceFile, node)

			// --- TS-specific checks (from @typescript-eslint extension) ---
			// These are checked first and may short-circuit with their own reporting.
			// isAllowed: true = skip, false = report as TS violation, nil = fall through to base logic
			var isAllowed *bool
			trueVal := true
			falseVal := false

			// Check if the value is in the ignore list (sign-aware)
			if opts.ignore[normalizeLiteralValue(node, raw, isBigInt)] {
				isAllowed = &trueVal
			}

			// Check TS enum member
			if isAllowed == nil && isParentTSEnumDeclaration(node) {
				if opts.ignoreEnums {
					isAllowed = &trueVal
				} else {
					isAllowed = &falseVal
				}
			}

			// Check TS numeric literal type
			if isAllowed == nil && isTSNumericLiteralType(node) {
				if opts.ignoreNumericLiteralTypes {
					isAllowed = &trueVal
				} else {
					isAllowed = &falseVal
				}
			}

			// Check TS type index
			if isAllowed == nil && isAncestorTSIndexedAccessType(node) {
				if opts.ignoreTypeIndexes {
					isAllowed = &trueVal
				} else {
					isAllowed = &falseVal
				}
			}

			// Check readonly class property
			if isAllowed == nil && isParentTSReadonlyPropertyDefinition(node) {
				if opts.ignoreReadonlyClassProperties {
					isAllowed = &trueVal
				} else {
					isAllowed = &falseVal
				}
			}

			if isAllowed != nil && *isAllowed {
				return
			}

			if isAllowed != nil && !*isAllowed {
				// Report as TS violation: only prepend '-' for negative numbers (not '+')
				reportNode := node
				reportRaw := raw
				if node.Parent != nil && node.Parent.Kind == ast.KindPrefixUnaryExpression {
					pref := node.Parent.AsPrefixUnaryExpression()
					if pref.Operator == ast.KindMinusToken {
						reportNode = node.Parent
						reportRaw = "-" + raw
					}
				}
				ctx.ReportNode(reportNode, rule.RuleMessage{
					Id:          noMagicMessage.Id,
					Description: strings.Replace(noMagicMessage.Description, "{{raw}}", reportRaw, 1),
				})
				return
			}

			// --- Core ESLint base rule logic ---
			// Determine fullNumberNode and raw (handling unary +/-)
			fullNumberNode := node
			fullRaw := raw
			var numericValue float64
			var bigintValue *big.Int

			if isBigInt {
				text := strings.TrimSuffix(raw, "n")
				bigintValue, _ = new(big.Int).SetString(text, 0)
				if bigintValue == nil {
					bigintValue = new(big.Int)
				}
			} else {
				numericValue, _ = parseRawNumericValue(raw)
			}

			if node.Parent != nil && node.Parent.Kind == ast.KindPrefixUnaryExpression {
				pref := node.Parent.AsPrefixUnaryExpression()
				if pref.Operator == ast.KindMinusToken || pref.Operator == ast.KindPlusToken {
					fullNumberNode = node.Parent
					fullRaw = utils.TrimmedNodeText(ctx.SourceFile, node.Parent)
					if pref.Operator == ast.KindMinusToken {
						if isBigInt {
							bigintValue = new(big.Int).Neg(bigintValue)
						} else {
							numericValue = -numericValue
						}
					}
				}
			}

			parent := fullNumberNode.Parent
			if parent == nil {
				return
			}

			// Check ignore list with the full (signed) value
			var valueKey string
			if isBigInt {
				valueKey = "bigint:" + bigintValue.String()
			} else {
				valueKey = strconv.FormatFloat(numericValue, 'f', -1, 64)
			}
			if opts.ignore[valueKey] {
				return
			}

			// Always allow parseInt radix and JSX numbers
			if isParseIntRadix(fullNumberNode) || isJSXNumber(fullNumberNode) {
				return
			}

			// Check optional ignore conditions
			if opts.ignoreDefaultValues && isDefaultValue(fullNumberNode) {
				return
			}
			if opts.ignoreClassFieldInitialValues && isClassFieldInitialValue(fullNumberNode) {
				return
			}
			if opts.ignoreArrayIndexes && isArrayIndex(fullNumberNode, numericValue, bigintValue, isBigInt) {
				return
			}

			// Report
			if parent.Kind == ast.KindVariableDeclaration {
				if opts.enforceConst {
					// Check if the variable declaration list uses 'const'
					declList := parent.Parent
					if declList != nil && declList.Kind == ast.KindVariableDeclarationList && !ast.IsVarConst(declList) {
						ctx.ReportNode(fullNumberNode, useConstMessage)
					}
				}
			} else if !isOkParent(parent, opts.detectObjects) {
				ctx.ReportNode(fullNumberNode, rule.RuleMessage{
					Id:          noMagicMessage.Id,
					Description: strings.Replace(noMagicMessage.Description, "{{raw}}", fullRaw, 1),
				})
			}
		}

		return rule.RuleListeners{
			ast.KindNumericLiteral: func(node *ast.Node) {
				handleNumericNode(node, false)
			},
			ast.KindBigIntLiteral: func(node *ast.Node) {
				handleNumericNode(node, true)
			},
		}
	},
})

// isOkParent checks if the parent node type is one that allows numbers without reporting.
// When detectObjects is false (default), numbers in object literals, property assignments,
// and assignment expressions (to non-identifiers) are allowed.
func isOkParent(parent *ast.Node, detectObjects bool) bool {
	if detectObjects {
		return false
	}
	switch parent.Kind {
	case ast.KindObjectLiteralExpression:
		return true
	case ast.KindPropertyAssignment:
		return true
	case ast.KindShorthandPropertyAssignment:
		// ShorthandPropertyAssignment with ObjectAssignmentInitializer means this number is a
		// destructuring default (e.g. {one = 1} = {}), not a property value. Don't suppress.
		spa := parent.AsShorthandPropertyAssignment()
		if spa.ObjectAssignmentInitializer != nil {
			return false
		}
		return true
	case ast.KindComputedPropertyName:
		// In ESTree, computed property keys like {[42]: true} have Property as parent (okType).
		// In tsgo, they have ComputedPropertyName as parent. Check if the grandparent is a
		// property assignment in an object literal (not a class field).
		gp := parent.Parent
		if gp != nil && (gp.Kind == ast.KindPropertyAssignment || gp.Kind == ast.KindShorthandPropertyAssignment) {
			return true
		}
		return false
	case ast.KindBinaryExpression:
		// AssignmentExpression in ESTree maps to BinaryExpression with assignment operator in tsgo
		op := parent.AsBinaryExpression().OperatorToken.Kind
		if isAssignmentOperator(op) {
			// If assigning to an identifier, it's NOT ok (magic number in identifier assignment)
			left := parent.AsBinaryExpression().Left
			if left != nil && ast.SkipParentheses(left).Kind == ast.KindIdentifier {
				return false
			}
			return true
		}
	}
	return false
}

// isAssignmentOperator returns true for = and compound assignment operators.
func isAssignmentOperator(kind ast.Kind) bool {
	switch kind {
	case ast.KindEqualsToken,
		ast.KindPlusEqualsToken,
		ast.KindMinusEqualsToken,
		ast.KindAsteriskEqualsToken,
		ast.KindSlashEqualsToken,
		ast.KindPercentEqualsToken,
		ast.KindAmpersandEqualsToken,
		ast.KindBarEqualsToken,
		ast.KindCaretEqualsToken,
		ast.KindLessThanLessThanEqualsToken,
		ast.KindGreaterThanGreaterThanEqualsToken,
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken,
		ast.KindAsteriskAsteriskEqualsToken,
		ast.KindBarBarEqualsToken,
		ast.KindAmpersandAmpersandEqualsToken,
		ast.KindQuestionQuestionEqualsToken:
		return true
	}
	return false
}

// getLiteralParent returns the "logical parent" of a numeric literal node,
// skipping any unary +/- prefix to get the containing statement/expression.
func getLiteralParent(node *ast.Node) *ast.Node {
	if node.Parent != nil && node.Parent.Kind == ast.KindPrefixUnaryExpression {
		pref := node.Parent.AsPrefixUnaryExpression()
		if pref.Operator == ast.KindMinusToken || pref.Operator == ast.KindPlusToken {
			return node.Parent.Parent
		}
	}
	return node.Parent
}

// isParentTSEnumDeclaration checks if the numeric literal is inside a TS enum member.
func isParentTSEnumDeclaration(node *ast.Node) bool {
	parent := getLiteralParent(node)
	return parent != nil && parent.Kind == ast.KindEnumMember
}

// isTSNumericLiteralType checks if the numeric literal is used as a TypeScript numeric literal type.
// Returns true for patterns like `type Foo = 1`, `type Foo = 1 | 2 | 3`.
func isTSNumericLiteralType(node *ast.Node) bool {
	// For negative numbers, step up past the unary minus
	current := node
	if current.Parent != nil && current.Parent.Kind == ast.KindPrefixUnaryExpression {
		pref := current.Parent.AsPrefixUnaryExpression()
		if pref.Operator == ast.KindMinusToken {
			current = current.Parent
		}
	}

	// Parent must be a LiteralType
	if current.Parent == nil || current.Parent.Kind != ast.KindLiteralType {
		return false
	}

	// Check if grandparent is TypeAliasDeclaration
	gp := current.Parent.Parent
	if gp == nil {
		return false
	}
	// Skip parenthesized types
	for gp != nil && gp.Kind == ast.KindParenthesizedType {
		gp = gp.Parent
	}
	if gp != nil && gp.Kind == ast.KindTypeAliasDeclaration {
		return true
	}

	// Check if grandparent is UnionType whose ancestor is TypeAliasDeclaration
	if gp != nil && gp.Kind == ast.KindUnionType {
		ancestor := gp.Parent
		// Walk up through nested union types and parenthesized types
		for ancestor != nil && (ancestor.Kind == ast.KindUnionType || ancestor.Kind == ast.KindParenthesizedType) {
			ancestor = ancestor.Parent
		}
		return ancestor != nil && ancestor.Kind == ast.KindTypeAliasDeclaration
	}

	return false
}

// isParentTSReadonlyPropertyDefinition checks if the numeric literal is inside a readonly class property.
func isParentTSReadonlyPropertyDefinition(node *ast.Node) bool {
	parent := getLiteralParent(node)
	if parent == nil || parent.Kind != ast.KindPropertyDeclaration {
		return false
	}
	return ast.HasSyntacticModifier(parent, ast.ModifierFlagsReadonly)
}

// isAncestorTSIndexedAccessType checks if the numeric literal is part of a type indexed access (e.g. Bar[0]).
func isAncestorTSIndexedAccessType(node *ast.Node) bool {
	// Get the logical parent (skip unary +/-)
	ancestor := getLiteralParent(node)

	// Walk up through union types, intersection types, and parenthesized types
	for ancestor != nil && ancestor.Parent != nil &&
		(ancestor.Parent.Kind == ast.KindUnionType ||
			ancestor.Parent.Kind == ast.KindIntersectionType ||
			ancestor.Parent.Kind == ast.KindParenthesizedType) {
		ancestor = ancestor.Parent
	}

	return ancestor != nil && ancestor.Parent != nil && ancestor.Parent.Kind == ast.KindIndexedAccessType
}

// isDefaultValue checks if the fullNumberNode is a default value assignment.
// In tsgo, default values appear as initializers in BindingElement, Parameter,
// or as BinaryExpression(=) inside destructuring assignments.
func isDefaultValue(fullNumberNode *ast.Node) bool {
	parent := fullNumberNode.Parent
	if parent == nil {
		return false
	}
	// Parameter default: function(param = 123)
	if parent.Kind == ast.KindParameter {
		return parent.AsParameterDeclaration().Initializer == fullNumberNode
	}
	// Binding element default: const { param = 123 } = obj; const [a = 1] = arr;
	if parent.Kind == ast.KindBindingElement {
		return parent.AsBindingElement().Initializer == fullNumberNode
	}
	// Shorthand property destructuring default: ({one = 1} = {})
	// In tsgo, this is a ShorthandPropertyAssignment with ObjectAssignmentInitializer
	if parent.Kind == ast.KindShorthandPropertyAssignment {
		spa := parent.AsShorthandPropertyAssignment()
		return spa.ObjectAssignmentInitializer == fullNumberNode
	}
	// Destructuring assignment default: [one = 1, two = 2] = arr
	// In tsgo, this is a BinaryExpression with = operator where the right operand is the default value
	if parent.Kind == ast.KindBinaryExpression {
		binExpr := parent.AsBinaryExpression()
		if binExpr.OperatorToken.Kind == ast.KindEqualsToken && binExpr.Right == fullNumberNode {
			// Check if this assignment is inside an array/object destructuring context
			// (the BinaryExpression's parent should be an array/object literal that's the left side of a destructuring assignment)
			return isInsideDestructuringAssignment(parent)
		}
	}
	return false
}

// isInsideDestructuringAssignment checks if a node is part of a destructuring assignment pattern.
func isInsideDestructuringAssignment(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil {
		switch parent.Kind {
		case ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression:
			// Continue walking up
			parent = parent.Parent
			continue
		case ast.KindSpreadElement, ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
			parent = parent.Parent
			continue
		case ast.KindBinaryExpression:
			// Check if this is a destructuring assignment (e.g., [...] = ...)
			binExpr := parent.AsBinaryExpression()
			if binExpr.OperatorToken.Kind == ast.KindEqualsToken {
				left := binExpr.Left
				return left != nil && (left.Kind == ast.KindArrayLiteralExpression || left.Kind == ast.KindObjectLiteralExpression)
			}
			return false
		case ast.KindForOfStatement, ast.KindForInStatement:
			return true
		default:
			return false
		}
	}
	return false
}

// isClassFieldInitialValue checks if the fullNumberNode is the direct initializer
// of a class field (PropertyDeclaration), not a computed key.
func isClassFieldInitialValue(fullNumberNode *ast.Node) bool {
	parent := fullNumberNode.Parent
	if parent == nil || parent.Kind != ast.KindPropertyDeclaration {
		return false
	}
	return parent.AsPropertyDeclaration().Initializer == fullNumberNode
}

// isParseIntRadix checks if the fullNumberNode is used as the radix argument in parseInt() or Number.parseInt().
func isParseIntRadix(fullNumberNode *ast.Node) bool {
	parent := fullNumberNode.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	call := parent.AsCallExpression()
	args := call.Arguments.Nodes
	if len(args) < 2 || args[1] != fullNumberNode {
		return false
	}
	callee := ast.SkipParentheses(call.Expression)
	if callee == nil {
		return false
	}
	// Check for parseInt(y, 10) or parseInt?.(y, 10)
	if callee.Kind == ast.KindIdentifier && callee.AsIdentifier().Text == "parseInt" {
		return true
	}
	// Check for Number.parseInt(y, 10) or Number?.parseInt(y, 10)
	return utils.IsSpecificMemberAccess(call.Expression, "Number", "parseInt")
}

// isJSXNumber checks if the fullNumberNode is a direct child of a JSX node.
func isJSXNumber(fullNumberNode *ast.Node) bool {
	parent := fullNumberNode.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindJsxExpression, ast.KindJsxAttribute, ast.KindJsxElement,
		ast.KindJsxSelfClosingElement, ast.KindJsxOpeningElement,
		ast.KindJsxFragment, ast.KindJsxSpreadAttribute:
		return true
	}
	return false
}

// isArrayIndex checks if the fullNumberNode is used as a valid array index in an element access expression.
func isArrayIndex(fullNumberNode *ast.Node, numericValue float64, bigintValue *big.Int, isBigInt bool) bool {
	parent := fullNumberNode.Parent
	if parent == nil || parent.Kind != ast.KindElementAccessExpression {
		return false
	}
	elemAccess := parent.AsElementAccessExpression()
	if elemAccess.ArgumentExpression != fullNumberNode {
		return false
	}

	if isBigInt {
		if bigintValue == nil {
			return false
		}
		// BigInt must be >= 0 and < maxArrayLength
		if bigintValue.Sign() < 0 {
			return false
		}
		maxIdx := new(big.Int).SetUint64(maxArrayLength)
		return bigintValue.Cmp(maxIdx) < 0
	}

	// Numeric value must be a non-negative integer < maxArrayLength
	if !isIntegerValue(numericValue) {
		return false
	}
	return numericValue >= 0 && numericValue < maxArrayLength
}

// isIntegerValue checks if a float64 is an integer value.
func isIntegerValue(v float64) bool {
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return false
	}
	return v == math.Floor(v)
}
