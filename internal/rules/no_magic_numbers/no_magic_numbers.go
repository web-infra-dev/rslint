package no_magic_numbers

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type NoMagicNumbersOptions struct {
	DetectObjects                 bool  `json:"detectObjects"`
	EnforceConst                  bool  `json:"enforceConst"`
	Ignore                        []any `json:"ignore"`
	IgnoreArrayIndexes            bool  `json:"ignoreArrayIndexes"`
	IgnoreDefaultValues           bool  `json:"ignoreDefaultValues"`
	IgnoreClassFieldInitialValues bool  `json:"ignoreClassFieldInitialValues"`
	IgnoreEnums                   bool  `json:"ignoreEnums"`
	IgnoreNumericLiteralTypes     bool  `json:"ignoreNumericLiteralTypes"`
	IgnoreReadonlyClassProperties bool  `json:"ignoreReadonlyClassProperties"`
	IgnoreTypeIndexes             bool  `json:"ignoreTypeIndexes"`
}

// normalizeIgnoreValue converts string bigint values to actual numeric values
func normalizeIgnoreValue(value any) any {
	if strVal, ok := value.(string); ok {
		// Handle bigint notation (ends with 'n')
		if strings.HasSuffix(strVal, "n") {
			// Remove 'n' suffix and parse as big.Int
			numStr := strVal[:len(strVal)-1]
			if bigInt, ok := new(big.Int).SetString(numStr, 10); ok {
				return bigInt
			}
		}
	}
	return value
}

// normalizeLiteralValue converts the node to its numeric value, handling prefixed numbers (-1 / +1)
func normalizeLiteralValue(node *ast.Node) any {
	if node.Kind == ast.KindNumericLiteral {
		numLit := node.AsNumericLiteral()
		val := parseNumericValue(numLit.Text)

		// Check if parent is unary expression with - operator
		if node.Parent != nil && node.Parent.Kind == ast.KindPrefixUnaryExpression {
			unary := node.Parent.AsPrefixUnaryExpression()
			if unary.Operator == ast.KindMinusToken {
				switch v := val.(type) {
				case float64:
					return -v
				case int64:
					return -v
				case *big.Int:
					return new(big.Int).Neg(v)
				}
			}
		}
		return val
	} else if node.Kind == ast.KindBigIntLiteral {
		bigIntLit := node.AsBigIntLiteral()
		// Remove the 'n' suffix
		text := bigIntLit.Text
		if strings.HasSuffix(text, "n") {
			text = text[:len(text)-1]
		}
		bigInt, _ := new(big.Int).SetString(text, 10)

		// Check if parent is unary expression with - operator
		if node.Parent != nil && node.Parent.Kind == ast.KindPrefixUnaryExpression {
			unary := node.Parent.AsPrefixUnaryExpression()
			if unary.Operator == ast.KindMinusToken {
				return new(big.Int).Neg(bigInt)
			}
		}
		return bigInt
	}
	return nil
}

// parseNumericValue parses a numeric literal text into a float64 but preserves the raw text for comparison
func parseNumericValue(text string) any {
	// Handle different number formats
	if strings.HasPrefix(text, "0x") || strings.HasPrefix(text, "0X") {
		// Hexadecimal - return as integer to preserve the exact value
		if val, err := strconv.ParseInt(text[2:], 16, 64); err == nil {
			return val
		}
	} else if strings.HasPrefix(text, "0b") || strings.HasPrefix(text, "0B") {
		// Binary
		if val, err := strconv.ParseInt(text[2:], 2, 64); err == nil {
			return val
		}
	} else if strings.HasPrefix(text, "0o") || strings.HasPrefix(text, "0O") {
		// Octal
		if val, err := strconv.ParseInt(text[2:], 8, 64); err == nil {
			return val
		}
	} else if len(text) > 1 && text[0] == '0' && text[1] >= '0' && text[1] <= '7' {
		// Legacy octal
		if val, err := strconv.ParseInt(text, 8, 64); err == nil {
			return val
		}
	} else {
		// Decimal (including scientific notation)
		if val, err := strconv.ParseFloat(text, 64); err == nil {
			return val
		}
	}
	return 0.0
}

// valuesEqual checks if two values are equal, handling bigint comparison and numeric type conversion
func valuesEqual(a, b any) bool {
	// Handle bigint comparison
	if aBig, ok := a.(*big.Int); ok {
		if bBig, ok := b.(*big.Int); ok {
			return aBig.Cmp(bBig) == 0
		}
		// Try to convert b to bigint if it's a numeric type
		if bFloat, ok := b.(float64); ok && bFloat == math.Trunc(bFloat) {
			bBig := big.NewInt(int64(bFloat))
			return aBig.Cmp(bBig) == 0
		}
		if bInt, ok := b.(int64); ok {
			bBig := big.NewInt(bInt)
			return aBig.Cmp(bBig) == 0
		}
		return false
	}

	if bBig, ok := b.(*big.Int); ok {
		if aFloat, ok := a.(float64); ok && aFloat == math.Trunc(aFloat) {
			aBig := big.NewInt(int64(aFloat))
			return aBig.Cmp(bBig) == 0
		}
		if aInt, ok := a.(int64); ok {
			aBig := big.NewInt(aInt)
			return aBig.Cmp(bBig) == 0
		}
		return false
	}

	// Handle numeric type conversions
	if aFloat, ok := a.(float64); ok {
		if bFloat, ok := b.(float64); ok {
			return aFloat == bFloat
		}
		if bInt, ok := b.(int64); ok {
			return aFloat == float64(bInt)
		}
		if bIntInterface, ok := b.(int); ok {
			return aFloat == float64(bIntInterface)
		}
	}

	if aInt, ok := a.(int64); ok {
		if bFloat, ok := b.(float64); ok {
			return float64(aInt) == bFloat
		}
		if bInt, ok := b.(int64); ok {
			return aInt == bInt
		}
		if bIntInterface, ok := b.(int); ok {
			return aInt == int64(bIntInterface)
		}
	}

	// Handle int interface from JSON
	if aIntInterface, ok := a.(int); ok {
		if bFloat, ok := b.(float64); ok {
			return float64(aIntInterface) == bFloat
		}
		if bInt, ok := b.(int64); ok {
			return int64(aIntInterface) == bInt
		}
		if bIntInterface, ok := b.(int); ok {
			return aIntInterface == bIntInterface
		}
	}

	// Regular comparison
	return a == b
}

// getLiteralParent gets the true parent of the literal, handling prefixed numbers (-1 / +1)
func getLiteralParent(node *ast.Node) *ast.Node {
	if node.Parent != nil && node.Parent.Kind == ast.KindPrefixUnaryExpression {
		unary := node.Parent.AsPrefixUnaryExpression()
		if unary.Operator == ast.KindMinusToken || unary.Operator == ast.KindPlusToken {
			return node.Parent.Parent
		}
	}
	return node.Parent
}

// isGrandparentTSTypeAliasDeclaration checks if the node grandparent is a TypeScript type alias declaration
func isGrandparentTSTypeAliasDeclaration(node *ast.Node) bool {
	return node.Parent != nil && node.Parent.Parent != nil &&
		node.Parent.Parent.Kind == ast.KindTypeAliasDeclaration
}

// isGrandparentTSUnionType checks if the node grandparent is a TypeScript union type and its parent is a type alias declaration
func isGrandparentTSUnionType(node *ast.Node) bool {
	if node.Parent != nil && node.Parent.Parent != nil &&
		node.Parent.Parent.Kind == ast.KindUnionType {
		return isGrandparentTSTypeAliasDeclaration(node.Parent)
	}
	return false
}

// isParentTSEnumDeclaration checks if the node parent is a TypeScript enum member
func isParentTSEnumDeclaration(node *ast.Node) bool {
	parent := getLiteralParent(node)
	return parent != nil && parent.Kind == ast.KindEnumMember
}

// isParentTSLiteralType checks if the node parent is a TypeScript literal type
func isParentTSLiteralType(node *ast.Node) bool {
	return node.Parent != nil && node.Parent.Kind == ast.KindLiteralType
}

// isTSNumericLiteralType checks if the node is a valid TypeScript numeric literal type
func isTSNumericLiteralType(node *ast.Node) bool {
	actualNode := node

	// For negative numbers, use the parent node
	if node.Parent != nil && node.Parent.Kind == ast.KindPrefixUnaryExpression {
		unary := node.Parent.AsPrefixUnaryExpression()
		if unary.Operator == ast.KindMinusToken {
			actualNode = node.Parent
		}
	}

	// If the parent node is not a TSLiteralType, early return
	if !isParentTSLiteralType(actualNode) {
		return false
	}

	// If the grandparent is a TSTypeAliasDeclaration, ignore
	if isGrandparentTSTypeAliasDeclaration(actualNode) {
		return true
	}

	// If the grandparent is a TSUnionType and it's parent is a TSTypeAliasDeclaration, ignore
	if isGrandparentTSUnionType(actualNode) {
		return true
	}

	return false
}

// isParentTSReadonlyPropertyDefinition checks if the node parent is a readonly class property
func isParentTSReadonlyPropertyDefinition(node *ast.Node) bool {
	parent := getLiteralParent(node)

	if parent != nil && parent.Kind == ast.KindPropertyDeclaration {
		propDecl := parent.AsPropertyDeclaration()
		// Check if property has readonly modifier
		if propDecl.Modifiers() != nil {
			for _, mod := range propDecl.Modifiers().Nodes {
				if mod.Kind == ast.KindReadonlyKeyword {
					return true
				}
			}
		}
	}

	return false
}

// isAncestorTSIndexedAccessType checks if the node is part of a type indexed access (eg. Foo[4])
func isAncestorTSIndexedAccessType(node *ast.Node) bool {
	// Handle unary expressions (eg. -4)
	ancestor := getLiteralParent(node)

	// Go up through any nesting of union/intersection types and parentheses
	for ancestor != nil && ancestor.Parent != nil {
		switch ancestor.Parent.Kind {
		case ast.KindUnionType, ast.KindIntersectionType, ast.KindParenthesizedType:
			ancestor = ancestor.Parent
		case ast.KindIndexedAccessType:
			return true
		default:
			return false
		}
	}

	return false
}

// isArrayIndex checks if the node is being used as an array index
func isArrayIndex(node *ast.Node) bool {
	parent := getLiteralParent(node)
	return parent != nil && parent.Kind == ast.KindElementAccessExpression &&
		parent.AsElementAccessExpression().ArgumentExpression == node
}

// isDefaultValue checks if the node is a default value in a function parameter or property
func isDefaultValue(node *ast.Node) bool {
	parent := getLiteralParent(node)
	if parent == nil {
		return false
	}

	// Check for default parameter values
	if parent.Kind == ast.KindParameter {
		param := parent.AsParameterDeclaration()
		return param.Initializer == node
	}

	// Check for default property values (handled by ignoreClassFieldInitialValues)

	return false
}

// isClassFieldInitialValue checks if the node is a class field initial value
func isClassFieldInitialValue(node *ast.Node) bool {
	parent := getLiteralParent(node)
	if parent == nil {
		return false
	}

	// Check for class property initializers
	if parent.Kind == ast.KindPropertyDeclaration {
		propDecl := parent.AsPropertyDeclaration()
		return propDecl.Initializer == node
	}

	return false
}

// isObjectProperty checks if the node is an object property value
func isObjectProperty(node *ast.Node) bool {
	parent := getLiteralParent(node)
	return parent != nil && parent.Kind == ast.KindPropertyAssignment
}

// checkNode checks a single numeric literal node
func checkNode(ctx rule.RuleContext, node *ast.Node, opts NoMagicNumbersOptions, ignored map[string]bool) {
	// This will be `true` if we're configured to ignore this case
	// It will be `false` if we're not configured to ignore this case
	// It will remain unset if this is not one of our exception cases
	var isAllowed *bool

	// Get the numeric value
	value := normalizeLiteralValue(node)
	if value == nil {
		return
	}

	// Check if the node is ignored by value
	for _, ignoreVal := range opts.Ignore {
		normalized := normalizeIgnoreValue(ignoreVal)
		if valuesEqual(value, normalized) {
			allowed := true
			isAllowed = &allowed
			break
		}
	}

	// Check if the node is a TypeScript enum declaration
	if isAllowed == nil && isParentTSEnumDeclaration(node) {
		allowed := opts.IgnoreEnums
		isAllowed = &allowed
	}

	// Check TypeScript specific nodes for Numeric Literal
	if isAllowed == nil && isTSNumericLiteralType(node) {
		allowed := opts.IgnoreNumericLiteralTypes
		isAllowed = &allowed
	}

	// Check if the node is a type index
	if isAllowed == nil && isAncestorTSIndexedAccessType(node) {
		allowed := opts.IgnoreTypeIndexes
		isAllowed = &allowed
	}

	// Check if the node is a readonly class property
	if isAllowed == nil && isParentTSReadonlyPropertyDefinition(node) {
		allowed := opts.IgnoreReadonlyClassProperties
		isAllowed = &allowed
	}

	// Check if the node is an array index
	if isAllowed == nil && opts.IgnoreArrayIndexes && isArrayIndex(node) {
		allowed := true
		isAllowed = &allowed
	}

	// Check if the node is a default value
	if isAllowed == nil && opts.IgnoreDefaultValues && isDefaultValue(node) {
		allowed := true
		isAllowed = &allowed
	}

	// Check if the node is a class field initial value
	if isAllowed == nil && opts.IgnoreClassFieldInitialValues && isClassFieldInitialValue(node) {
		allowed := true
		isAllowed = &allowed
	}

	// Check if the node is an object property and detectObjects is false
	if isAllowed == nil && !opts.DetectObjects && isObjectProperty(node) {
		allowed := true
		isAllowed = &allowed
	}

	// If we've hit a case where the ignore option is true we can return now
	if isAllowed != nil && *isAllowed {
		return
	}

	// Report the error
	fullNumberNode := node
	raw := ""

	if node.Kind == ast.KindNumericLiteral {
		raw = node.AsNumericLiteral().Text
	} else if node.Kind == ast.KindBigIntLiteral {
		raw = node.AsBigIntLiteral().Text
	}

	// Handle negative numbers - report on the unary expression but use original raw for negative hex
	if node.Parent != nil && node.Parent.Kind == ast.KindPrefixUnaryExpression {
		unary := node.Parent.AsPrefixUnaryExpression()
		if unary.Operator == ast.KindMinusToken {
			fullNumberNode = node.Parent
			// For hex numbers, preserve the original format in the error message
			if strings.HasPrefix(raw, "0x") || strings.HasPrefix(raw, "0X") {
				raw = fmt.Sprintf("-%s", raw)
			} else {
				raw = fmt.Sprintf("-%s", raw)
			}
		}
	}

	message := rule.RuleMessage{
		Id:          "noMagic",
		Description: fmt.Sprintf("No magic number: %s.", raw),
	}

	// Check if enforceConst is enabled and suggest const declaration
	if opts.EnforceConst {
		// For now, just report without suggestions
		// TODO: Add fix suggestions for const declarations
		ctx.ReportNode(fullNumberNode, message)
	} else {
		ctx.ReportNode(fullNumberNode, message)
	}
}

var NoMagicNumbersRule = rule.Rule{
	Name: "no-magic-numbers",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Set default options
		opts := NoMagicNumbersOptions{
			DetectObjects:                 false,
			EnforceConst:                  false,
			Ignore:                        []any{},
			IgnoreArrayIndexes:            false,
			IgnoreDefaultValues:           false,
			IgnoreClassFieldInitialValues: false,
			IgnoreEnums:                   false,
			IgnoreNumericLiteralTypes:     false,
			IgnoreReadonlyClassProperties: false,
			IgnoreTypeIndexes:             false,
		}

		// Parse options with dual-format support (handles both array and object formats)
		if options != nil {
			var optsMap map[string]interface{}
			var ok bool

			// Handle array format: [{ option: value }]
			if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
				optsMap, ok = optArray[0].(map[string]interface{})
			} else {
				// Handle direct object format: { option: value }
				optsMap, ok = options.(map[string]interface{})
			}

			if ok {
				if val, ok := optsMap["detectObjects"].(bool); ok {
					opts.DetectObjects = val
				}
				if val, ok := optsMap["enforceConst"].(bool); ok {
					opts.EnforceConst = val
				}
				if val, ok := optsMap["ignore"]; ok {
					if ignoreList, ok := val.([]interface{}); ok {
						opts.Ignore = ignoreList
					}
				}
				if val, ok := optsMap["ignoreArrayIndexes"].(bool); ok {
					opts.IgnoreArrayIndexes = val
				}
				if val, ok := optsMap["ignoreDefaultValues"].(bool); ok {
					opts.IgnoreDefaultValues = val
				}
				if val, ok := optsMap["ignoreClassFieldInitialValues"].(bool); ok {
					opts.IgnoreClassFieldInitialValues = val
				}
				if val, ok := optsMap["ignoreEnums"].(bool); ok {
					opts.IgnoreEnums = val
				}
				if val, ok := optsMap["ignoreNumericLiteralTypes"].(bool); ok {
					opts.IgnoreNumericLiteralTypes = val
				}
				if val, ok := optsMap["ignoreReadonlyClassProperties"].(bool); ok {
					opts.IgnoreReadonlyClassProperties = val
				}
				if val, ok := optsMap["ignoreTypeIndexes"].(bool); ok {
					opts.IgnoreTypeIndexes = val
				}
			}
		}

		// Create a map for faster ignore lookups
		ignored := make(map[string]bool)

		return rule.RuleListeners{
			ast.KindNumericLiteral: func(node *ast.Node) {
				checkNode(ctx, node, opts, ignored)
			},
			ast.KindBigIntLiteral: func(node *ast.Node) {
				checkNode(ctx, node, opts, ignored)
			},
		}
	},
}
