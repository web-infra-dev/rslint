package no_self_assign

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// skipOuterExprKinds defines which outer expressions to skip: parentheses and type assertions.
const skipOuterExprKinds = ast.OEKParentheses | ast.OEKTypeAssertions

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

		for _, lProp := range leftObj.Properties.Nodes {
			// Handle spread on left: {...x} = {...y}
			if lProp.Kind == ast.KindSpreadAssignment {
				// Find corresponding spread on right, but only after startJ
				for j := startJ; j < len(rightProps); j++ {
					rProp := rightProps[j]
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

			// Find matching property on right, starting after the last spread
			for j := startJ; j < len(rightProps); j++ {
				rProp := rightProps[j]
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
