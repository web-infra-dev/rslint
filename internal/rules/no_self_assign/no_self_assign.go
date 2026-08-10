package no_self_assign

import (
	_ "embed"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_self_assign.schema.json
var schemaJSON []byte

// skipOuterExprKinds defines which outer expressions to skip: parentheses and type assertions.
const skipOuterExprKinds = ast.OEKParentheses | ast.OEKTypeAssertions

// https://eslint.org/docs/latest/rules/no-self-assign
var NoSelfAssignRule = rule.Rule{
	Name:   "no-self-assign",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binExpr := node.AsBinaryExpression()
				if binExpr == nil || binExpr.OperatorToken == nil {
					return
				}

				// Check for assignment operators: =, &&=, ||=, ??=
				op := binExpr.OperatorToken.Kind
				if op != ast.KindEqualsToken &&
					op != ast.KindAmpersandAmpersandEqualsToken &&
					op != ast.KindBarBarEqualsToken &&
					op != ast.KindQuestionQuestionEqualsToken {
					return
				}

				eachSelfAssignment(binExpr.Left, binExpr.Right, opts.props, func(rightNode *ast.Node) {
					reportRange := utils.TrimNodeTextRange(ctx.SourceFile, rightNode)
					name := removeWhitespace(ctx.SourceFile.Text()[reportRange.Pos():reportRange.End()])
					ctx.ReportRange(reportRange, rule.RuleMessage{
						Id:          "selfAssignment",
						Description: "'" + name + "' is assigned to itself.",
					})
				})
			},
		}
	},
}

type selfAssignOptions struct {
	props bool
}

func parseOptions(options []any) selfAssignOptions {
	result := selfAssignOptions{props: true} // default: props is true
	if len(options) == 0 {
		return result
	}

	m, _ := options[0].(map[string]any)
	if p, ok := m["props"].(bool); ok {
		result.props = p
	}

	return result
}

// removeWhitespace normalizes node text such as "a . b" to "a.b" without
// allocating when the usual compact spelling contains no whitespace. Source
// text is overwhelmingly ASCII, so only non-ASCII bytes enter unicode.IsSpace.
func removeWhitespace(text string) string {
	firstWhitespace := -1
	for i := 0; i < len(text); {
		if text[i] < utf8.RuneSelf {
			if isASCIIWhitespace(text[i]) {
				firstWhitespace = i
				break
			}
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(text[i:])
		if unicode.IsSpace(r) {
			firstWhitespace = i
			break
		}
		i += size
	}
	if firstWhitespace < 0 {
		return text
	}

	var result strings.Builder
	result.Grow(len(text) - 1)
	result.WriteString(text[:firstWhitespace])
	for i := firstWhitespace; i < len(text); {
		if text[i] < utf8.RuneSelf {
			if !isASCIIWhitespace(text[i]) {
				result.WriteByte(text[i])
			}
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(text[i:])
		if !unicode.IsSpace(r) {
			// Copy the original bytes so malformed UTF-8 is preserved instead of
			// being rewritten as the replacement rune.
			result.WriteString(text[i : i+size])
		}
		i += size
	}
	return result.String()
}

func isASCIIWhitespace(ch byte) bool {
	return ch == ' ' || ch >= '\t' && ch <= '\r'
}

const maxLinearObjectProperties = 32

type namedObjectProperty struct {
	name  string
	value *ast.Node
}

// eachSelfAssignment recursively compares the left and right nodes of an assignment,
// calling report for each right-side node that is self-assigned.
func eachSelfAssignment(left *ast.Node, right *ast.Node, props bool, report func(*ast.Node)) {
	if left == nil || right == nil {
		return
	}

	// Skip type assertions / parentheses on both sides
	left = ast.SkipOuterExpressions(left, skipOuterExprKinds)
	right = ast.SkipOuterExpressions(right, skipOuterExprKinds)

	switch {
	// Identifier = Identifier
	case left.Kind == ast.KindIdentifier && right.Kind == ast.KindIdentifier:
		if left.AsIdentifier().Text == right.AsIdentifier().Text {
			report(right)
		}

	// ArrayLiteralExpression (destructuring) = ArrayLiteralExpression
	case left.Kind == ast.KindArrayLiteralExpression && right.Kind == ast.KindArrayLiteralExpression:
		leftArr := left.AsArrayLiteralExpression()
		rightArr := right.AsArrayLiteralExpression()
		if leftArr.Elements == nil || rightArr.Elements == nil {
			return
		}
		leftElems := leftArr.Elements.Nodes
		rightElems := rightArr.Elements.Nodes
		end := len(leftElems)
		if len(rightElems) < end {
			end = len(rightElems)
		}
		for i := range end {
			l := leftElems[i]
			r := rightElems[i]

			// Handle rest element (...x) on the left.
			// If there are more elements on the right after this position,
			// the rest collects them too, so it's not a simple 1:1 match.
			// This matches ESLint's check: leftElement.type === "RestElement" && i < right.elements.length - 1
			if l.Kind == ast.KindSpreadElement {
				if i < len(rightElems)-1 {
					break
				}
				// ...x on left and ...y on right (last element)
				if r.Kind == ast.KindSpreadElement {
					eachSelfAssignment(l.AsSpreadElement().Expression, r.AsSpreadElement().Expression, props, report)
				}
				break
			}

			// Handle spread on the right
			if r.Kind == ast.KindSpreadElement {
				// can't statically determine what spread produces
				break
			}

			eachSelfAssignment(l, r, props, report)
		}

	// ObjectLiteralExpression (destructuring) = ObjectLiteralExpression
	case left.Kind == ast.KindObjectLiteralExpression && right.Kind == ast.KindObjectLiteralExpression:
		leftObj := left.AsObjectLiteralExpression()
		rightObj := right.AsObjectLiteralExpression()
		if leftObj.Properties == nil || rightObj.Properties == nil {
			return
		}

		rightProps := rightObj.Properties.Nodes

		// Find the index of the last SpreadAssignment on the right side.
		// Properties before (and including) a spread can be overwritten by it,
		// so only properties after the last spread are safe to compare.
		// This matches ESLint's startJ logic.
		startJ := 0
		for i := len(rightProps) - 1; i >= 0; i-- {
			if rightProps[i].Kind == ast.KindSpreadAssignment {
				startJ = i + 1
				break
			}
		}

		compareObjectProperties(leftObj.Properties.Nodes, rightProps[startJ:], props, report)

	// MemberExpression (PropertyAccessExpression or ElementAccessExpression) with props option.
	// Unlike destructuring patterns above, member expressions are compared as a whole
	// reference chain using utils.IsSameReference, matching ESLint's isSameReference approach.
	case props && ast.IsAccessExpression(left) && ast.IsAccessExpression(right):
		if utils.IsSameReference(left, right) {
			report(right)
		}

	// ThisKeyword = ThisKeyword
	case left.Kind == ast.KindThisKeyword && right.Kind == ast.KindThisKeyword:
		report(right)
	}
}

// compareObjectProperties compares properties after the right-hand side's last
// spread. Small patterns stay allocation-free; larger patterns build a compact
// name index so adversarial or generated destructuring assignments remain
// linear instead of repeatedly scanning the right-hand side.
func compareObjectProperties(leftProps, rightProps []*ast.Node, props bool, report func(*ast.Node)) {
	if len(rightProps) <= maxLinearObjectProperties {
		var properties [maxLinearObjectProperties]namedObjectProperty
		propertyCount := 0
		for _, rightProp := range rightProps {
			name, ok := getPropertyKeyName(rightProp)
			if !ok {
				continue
			}
			properties[propertyCount] = namedObjectProperty{
				name:  name,
				value: getPropertyValue(rightProp),
			}
			propertyCount++
		}

		for _, leftProp := range leftProps {
			leftName, ok := getPropertyKeyName(leftProp)
			if !ok {
				continue
			}
			leftValue := getPropertyValue(leftProp)
			for i := range propertyCount {
				if properties[i].name == leftName {
					eachSelfAssignment(leftValue, properties[i].value, props, report)
				}
			}
		}
		return
	}

	nextProperty := make([]int, len(rightProps))
	propertiesByName := make(map[string]int, len(rightProps))
	for i := len(rightProps) - 1; i >= 0; i-- {
		name, ok := getPropertyKeyName(rightProps[i])
		if !ok {
			continue
		}
		next := -1
		if head, exists := propertiesByName[name]; exists {
			next = head
		}
		nextProperty[i] = next
		propertiesByName[name] = i
	}

	for _, leftProp := range leftProps {
		name, ok := getPropertyKeyName(leftProp)
		if !ok {
			continue
		}
		leftValue := getPropertyValue(leftProp)
		i, exists := propertiesByName[name]
		if !exists {
			continue
		}
		for i >= 0 {
			eachSelfAssignment(leftValue, getPropertyValue(rightProps[i]), props, report)
			i = nextProperty[i]
		}
	}
}

// getPropertyKeyName returns the static property name for a property assignment
// or shorthand property. The boolean distinguishes a valid empty-string key
// from a property whose name cannot be determined statically.
func getPropertyKeyName(prop *ast.Node) (string, bool) {
	if prop == nil {
		return "", false
	}

	switch prop.Kind {
	case ast.KindPropertyAssignment:
		nameNode := prop.AsPropertyAssignment().Name()
		if nameNode == nil {
			return "", false
		}
		name, ok := utils.GetStaticPropertyName(nameNode)
		if !ok {
			return "", false
		}
		return name, true

	case ast.KindShorthandPropertyAssignment:
		nameNode := prop.AsShorthandPropertyAssignment().Name()
		if nameNode == nil {
			return "", false
		}
		return nameNode.Text(), true
	}
	return "", false
}

// getPropertyValue returns the value node for a property assignment or shorthand property.
// For shorthand properties like { a }, the name identifier is used as the value.
func getPropertyValue(prop *ast.Node) *ast.Node {
	if prop == nil {
		return nil
	}

	switch prop.Kind {
	case ast.KindPropertyAssignment:
		return prop.AsPropertyAssignment().Initializer
	case ast.KindShorthandPropertyAssignment:
		spa := prop.AsShorthandPropertyAssignment()
		// If the property has a default value (e.g. {a = 1}), it's structurally
		// different from a plain shorthand ({a}), so don't match them.
		if spa.ObjectAssignmentInitializer != nil {
			return nil
		}
		return spa.Name()
	}
	return nil
}
