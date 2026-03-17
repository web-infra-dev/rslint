package no_self_assign

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-self-assign
var NoSelfAssignRule = rule.Rule{
	Name: "no-self-assign",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
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
					name := getNodeText(ctx.SourceFile, rightNode)
					ctx.ReportNode(rightNode, rule.RuleMessage{
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

func parseOptions(opts any) selfAssignOptions {
	result := selfAssignOptions{props: true} // default: props is true

	optsMap := utils.GetOptionsMap(opts)
	if optsMap != nil {
		if p, ok := optsMap["props"].(bool); ok {
			result.props = p
		}
	}

	return result
}

// getNodeText returns the source text of a node, with whitespace removed.
func getNodeText(sourceFile *ast.SourceFile, node *ast.Node) string {
	trimmed := utils.TrimNodeTextRange(sourceFile, node)
	text := sourceFile.Text()[trimmed.Pos():trimmed.End()]
	// Remove internal whitespace to normalize e.g. "a . b" -> "a.b"
	return strings.Join(strings.Fields(text), "")
}

// eachSelfAssignment recursively compares the left and right nodes of an assignment,
// calling report for each right-side node that is self-assigned.
func eachSelfAssignment(left *ast.Node, right *ast.Node, props bool, report func(*ast.Node)) {
	if left == nil || right == nil {
		return
	}

	// Skip type assertions / parentheses on right side
	right = skipTypeAssertions(right)

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

			// Handle rest element (...x) on the left
			if l.Kind == ast.KindSpreadElement {
				// ...x on left and ...y on right
				if r.Kind == ast.KindSpreadElement {
					eachSelfAssignment(l.AsSpreadElement().Expression, r.AsSpreadElement().Expression, props, report)
				}
				// For ...x = [y, z], the rest collects remaining - not a simple match
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

		// Check if right side has any spread - if so, stop at the first spread
		// because spreads can change the meaning of properties
		rightHasSpread := false
		for _, rp := range rightObj.Properties.Nodes {
			if rp.Kind == ast.KindSpreadAssignment {
				rightHasSpread = true
				break
			}
		}
		_ = rightHasSpread

		for _, lProp := range leftObj.Properties.Nodes {
			// Handle spread on left: {...x} = {...y}
			if lProp.Kind == ast.KindSpreadAssignment {
				// Find corresponding spread on right
				for _, rProp := range rightObj.Properties.Nodes {
					if rProp.Kind == ast.KindSpreadAssignment {
						eachSelfAssignment(lProp.AsSpreadAssignment().Expression, rProp.AsSpreadAssignment().Expression, props, report)
						break
					}
				}
				continue
			}

			lName := getPropertyKeyName(lProp)
			if lName == "" {
				continue
			}

			// Find matching property on right
			for _, rProp := range rightObj.Properties.Nodes {
				if rProp.Kind == ast.KindSpreadAssignment {
					continue
				}

				rName := getPropertyKeyName(rProp)
				if rName == lName {
					lValue := getPropertyValue(lProp)
					rValue := getPropertyValue(rProp)
					eachSelfAssignment(lValue, rValue, props, report)
					break
				}
			}
		}

	// PropertyAccessExpression = PropertyAccessExpression (with props option)
	case props &&
		left.Kind == ast.KindPropertyAccessExpression && right.Kind == ast.KindPropertyAccessExpression:
		leftProp := left.AsPropertyAccessExpression()
		rightProp := right.AsPropertyAccessExpression()

		// Optional chaining on the right side (a.b = a?.b) is not self-assignment
		if rightProp.QuestionDotToken != nil {
			return
		}

		leftName := leftProp.Name()
		rightName := rightProp.Name()
		if leftName != nil && rightName != nil &&
			leftName.Text() == rightName.Text() {
			eachSelfAssignment(leftProp.Expression, rightProp.Expression, props, report)
		}

	// ElementAccessExpression = ElementAccessExpression (with props option)
	case props &&
		left.Kind == ast.KindElementAccessExpression && right.Kind == ast.KindElementAccessExpression:
		leftElem := left.AsElementAccessExpression()
		rightElem := right.AsElementAccessExpression()

		// Optional chaining on the right side (a[b] = a?.[b]) is not self-assignment
		if rightElem.QuestionDotToken != nil {
			return
		}

		if isSameReference(leftElem.ArgumentExpression, rightElem.ArgumentExpression) {
			eachSelfAssignment(leftElem.Expression, rightElem.Expression, props, report)
		}

	// ThisKeyword = ThisKeyword
	case left.Kind == ast.KindThisKeyword && right.Kind == ast.KindThisKeyword:
		report(right)
	}
}

// skipTypeAssertions skips parenthesized expressions and type assertion expressions.
func skipTypeAssertions(node *ast.Node) *ast.Node {
	for node != nil {
		switch node.Kind {
		case ast.KindParenthesizedExpression:
			node = node.AsParenthesizedExpression().Expression
		case ast.KindTypeAssertionExpression:
			node = node.AsTypeAssertion().Expression
		case ast.KindAsExpression:
			node = node.AsAsExpression().Expression
		default:
			return node
		}
	}
	return node
}

// isSameReference checks if two nodes refer to the same value.
// Used for comparing element access arguments like a[0] = a[0].
func isSameReference(left *ast.Node, right *ast.Node) bool {
	if left == nil || right == nil {
		return false
	}
	if left.Kind != right.Kind {
		return false
	}

	switch left.Kind {
	case ast.KindIdentifier:
		return left.AsIdentifier().Text == right.AsIdentifier().Text
	case ast.KindThisKeyword:
		return true
	case ast.KindStringLiteral:
		return left.AsStringLiteral().Text == right.AsStringLiteral().Text
	case ast.KindNumericLiteral:
		return left.AsNumericLiteral().Text == right.AsNumericLiteral().Text
	}
	return false
}

// getPropertyKeyName returns the static property name for a property assignment or shorthand property.
func getPropertyKeyName(prop *ast.Node) string {
	if prop == nil {
		return ""
	}

	switch prop.Kind {
	case ast.KindPropertyAssignment:
		nameNode := prop.AsPropertyAssignment().Name()
		if nameNode == nil {
			return ""
		}
		name, ok := utils.GetStaticPropertyName(nameNode)
		if !ok {
			return ""
		}
		return name

	case ast.KindShorthandPropertyAssignment:
		nameNode := prop.AsShorthandPropertyAssignment().Name()
		if nameNode == nil {
			return ""
		}
		return nameNode.Text()
	}
	return ""
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
		return prop.AsShorthandPropertyAssignment().Name()
	}
	return nil
}
