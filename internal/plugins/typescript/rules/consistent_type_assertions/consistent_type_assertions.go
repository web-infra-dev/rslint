package consistent_type_assertions

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type AssertionStyle string

const (
	AssertionStyleAs           AssertionStyle = "as"
	AssertionStyleAngleBracket AssertionStyle = "angle-bracket"
	AssertionStyleNever        AssertionStyle = "never"
)

type LiteralAssertion string

const (
	LiteralAssertionAllow        LiteralAssertion = "allow"
	LiteralAssertionNever        LiteralAssertion = "never"
	LiteralAssertionAllowAsParam LiteralAssertion = "allow-as-parameter"
)

type ConsistentTypeAssertionsOptions struct {
	AssertionStyle              AssertionStyle   `json:"assertionStyle"`
	ObjectLiteralTypeAssertions LiteralAssertion `json:"objectLiteralTypeAssertions"`
	ArrayLiteralTypeAssertions  LiteralAssertion `json:"arrayLiteralTypeAssertions"`
}

// ConsistentTypeAssertionsRule enforces consistent type assertions
var ConsistentTypeAssertionsRule = rule.CreateRule(rule.Rule{
	Name: "consistent-type-assertions",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := ConsistentTypeAssertionsOptions{
		AssertionStyle:              AssertionStyleAs,
		ObjectLiteralTypeAssertions: LiteralAssertionAllow,
		ArrayLiteralTypeAssertions:  LiteralAssertionAllow,
	}

	// Parse options
	if options != nil {
		if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
			if optsMap, ok := optArray[0].(map[string]interface{}); ok {
				parseOptionsMap(optsMap, &opts)
			}
		} else if optsMap, ok := options.(map[string]interface{}); ok {
			parseOptionsMap(optsMap, &opts)
		}
	}

	// Helper to check if a node is a const assertion
	isConstAssertion := func(node *ast.Node) bool {
		if node == nil {
			return false
		}

		var typeNode *ast.Node

		switch node.Kind {
		case ast.KindAsExpression:
			asExpr := node.AsAsExpression()
			if asExpr != nil {
				typeNode = asExpr.Type
			}
		case ast.KindTypeAssertionExpression:
			typeAssertion := node.AsTypeAssertion()
			if typeAssertion != nil {
				typeNode = typeAssertion.Type
			}
		}

		if typeNode != nil && typeNode.Kind == ast.KindTypeReference {
			typeRef := typeNode.AsTypeReference()
			if typeRef != nil && typeRef.TypeName != nil {
				typeName := typeRef.TypeName
				if typeName.Kind == ast.KindIdentifier {
					ident := typeName.AsIdentifier()
					if ident != nil && ident.Text == "const" {
						return true
					}
				}
			}
		}

		return false
	}

	// Helper to check if type is any or unknown
	isAnyOrUnknown := func(typeNode *ast.Node) bool {
		if typeNode == nil {
			return false
		}

		if typeNode.Kind == ast.KindAnyKeyword || typeNode.Kind == ast.KindUnknownKeyword {
			return true
		}

		// Check for union types containing any/unknown
		if typeNode.Kind == ast.KindUnionType {
			unionType := typeNode.AsUnionTypeNode()
			if unionType != nil {
				types := unionType.Types.Nodes
				for _, t := range types {
					if t.Kind == ast.KindAnyKeyword || t.Kind == ast.KindUnknownKeyword {
						return true
					}
				}
			}
		}

		return false
	}

	// Helper to check if assertion is used as a parameter
	var isAsParameter func(node *ast.Node) bool
	isAsParameter = func(node *ast.Node) bool {
		if node == nil || node.Parent == nil {
			return false
		}

		parent := node.Parent

		switch parent.Kind {
		case ast.KindCallExpression:
			// Check if node is in arguments
			callExpr := parent.AsCallExpression()
			if callExpr != nil {
				args := callExpr.Arguments.Nodes
				for _, arg := range args {
					if arg == node {
						return true
					}
				}
			}
		case ast.KindNewExpression:
			// Check if node is in arguments
			newExpr := parent.AsNewExpression()
			if newExpr != nil && newExpr.Arguments != nil {
				args := newExpr.Arguments.Nodes
				for _, arg := range args {
					if arg == node {
						return true
					}
				}
			}
		case ast.KindThrowStatement:
			return true
		case ast.KindTemplateSpan:
			return true
		case ast.KindParameter:
			return true
		case ast.KindPropertyAssignment:
			// Check if it's a default value in a parameter
			propAssignment := parent.AsPropertyAssignment()
			if propAssignment != nil && propAssignment.Initializer == node {
				return isAsParameter(parent)
			}
		}

		return false
	}

	// Helper to check if expression is an object literal
	isObjectLiteral := func(node *ast.Node) bool {
		return node != nil && node.Kind == ast.KindObjectLiteralExpression
	}

	// Helper to check if expression is an array literal
	isArrayLiteral := func(node *ast.Node) bool {
		return node != nil && node.Kind == ast.KindArrayLiteralExpression
	}

	// Helper to check if a variable declaration can use type annotation
	canUseTypeAnnotation := func(node *ast.Node) bool {
		if node == nil || node.Parent == nil {
			return false
		}

		parent := node.Parent

		// Check if parent is a variable declaration
		if parent.Kind == ast.KindVariableDeclaration {
			return true
		}

		// Check if parent is a property declaration
		if parent.Kind == ast.KindPropertyDeclaration {
			return true
		}

		return false
	}

	checkAsExpression := func(node *ast.Node) {
		// Always allow const assertions
		if isConstAssertion(node) {
			return
		}

		asExpr := node.AsAsExpression()
		if asExpr == nil {
			return
		}

		expression := asExpr.Expression
		typeNode := asExpr.Type

		// Check assertion style
		if opts.AssertionStyle == AssertionStyleNever {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "never",
				Description: "Do not use any type assertions.",
			})
			return
		}

		if opts.AssertionStyle == AssertionStyleAngleBracket {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "as",
				Description: "Use angle-bracket type assertions instead of 'as' assertions.",
			})
			return
		}

		// Check object literal assertions
		if isObjectLiteral(expression) && !isAnyOrUnknown(typeNode) {
			if opts.ObjectLiteralTypeAssertions == LiteralAssertionNever {
				if canUseTypeAnnotation(node) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "object-literal-with-type-annotation",
						Description: "Use a type annotation instead of a type assertion for object literals.",
					})
				} else {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "never-object-literal",
						Description: "Use a type annotation or satisfies instead of a type assertion for object literals.",
					})
				}
				return
			}

			if opts.ObjectLiteralTypeAssertions == LiteralAssertionAllowAsParam {
				if !isAsParameter(node) {
					if canUseTypeAnnotation(node) {
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "object-literal-with-type-annotation",
							Description: "Use a type annotation instead of a type assertion for object literals.",
						})
					} else {
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "never-object-literal",
							Description: "Use a type annotation or satisfies instead of a type assertion for object literals.",
						})
					}
					return
				}
			}
		}

		// Check array literal assertions
		if isArrayLiteral(expression) && !isAnyOrUnknown(typeNode) {
			if opts.ArrayLiteralTypeAssertions == LiteralAssertionNever {
				if canUseTypeAnnotation(node) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "array-literal-with-type-annotation",
						Description: "Use a type annotation instead of a type assertion for array literals.",
					})
				} else {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "never-array-literal",
						Description: "Use a type annotation or satisfies instead of a type assertion for array literals.",
					})
				}
				return
			}

			if opts.ArrayLiteralTypeAssertions == LiteralAssertionAllowAsParam {
				if !isAsParameter(node) {
					if canUseTypeAnnotation(node) {
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "array-literal-with-type-annotation",
							Description: "Use a type annotation instead of a type assertion for array literals.",
						})
					} else {
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "never-array-literal",
							Description: "Use a type annotation or satisfies instead of a type assertion for array literals.",
						})
					}
					return
				}
			}
		}
	}

	checkTypeAssertion := func(node *ast.Node) {
		// Always allow const assertions
		if isConstAssertion(node) {
			return
		}

		typeAssertion := node.AsTypeAssertion()
		if typeAssertion == nil {
			return
		}

		expression := typeAssertion.Expression
		typeNode := typeAssertion.Type

		// Check assertion style 'never' first
		if opts.AssertionStyle == AssertionStyleNever {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "never",
				Description: "Do not use any type assertions.",
			})
			return
		}

		// Check object literal assertions BEFORE checking assertion style
		if isObjectLiteral(expression) && !isAnyOrUnknown(typeNode) {
			if opts.ObjectLiteralTypeAssertions == LiteralAssertionNever {
				if canUseTypeAnnotation(node) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "object-literal-with-type-annotation",
						Description: "Use a type annotation instead of a type assertion for object literals.",
					})
				} else {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "never-object-literal",
						Description: "Use a type annotation or satisfies instead of a type assertion for object literals.",
					})
				}
				return
			}

			if opts.ObjectLiteralTypeAssertions == LiteralAssertionAllowAsParam {
				if !isAsParameter(node) {
					if canUseTypeAnnotation(node) {
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "object-literal-with-type-annotation",
							Description: "Use a type annotation instead of a type assertion for object literals.",
						})
					} else {
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "never-object-literal",
							Description: "Use a type annotation or satisfies instead of a type assertion for object literals.",
						})
					}
					return
				}
			}
		}

		// Check array literal assertions
		if isArrayLiteral(expression) && !isAnyOrUnknown(typeNode) {
			if opts.ArrayLiteralTypeAssertions == LiteralAssertionNever {
				if canUseTypeAnnotation(node) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "array-literal-with-type-annotation",
						Description: "Use a type annotation instead of a type assertion for array literals.",
					})
				} else {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "never-array-literal",
						Description: "Use a type annotation or satisfies instead of a type assertion for array literals.",
					})
				}
				return
			}

			if opts.ArrayLiteralTypeAssertions == LiteralAssertionAllowAsParam {
				if !isAsParameter(node) {
					if canUseTypeAnnotation(node) {
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "array-literal-with-type-annotation",
							Description: "Use a type annotation instead of a type assertion for array literals.",
						})
					} else {
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "never-array-literal",
							Description: "Use a type annotation or satisfies instead of a type assertion for array literals.",
						})
					}
					return
				}
			}
		}

		// Check assertion style (after literal checks)
		if opts.AssertionStyle == AssertionStyleAs {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "angle-bracket",
				Description: "Use 'as' assertions instead of angle-bracket type assertions.",
			})
			return
		}
	}

	return rule.RuleListeners{
		ast.KindAsExpression:            checkAsExpression,
		ast.KindTypeAssertionExpression: checkTypeAssertion,
	}
}

func parseOptionsMap(optsMap map[string]interface{}, opts *ConsistentTypeAssertionsOptions) {
	if v, exists := optsMap["assertionStyle"]; exists {
		if str, ok := v.(string); ok {
			opts.AssertionStyle = AssertionStyle(str)
		}
	}

	if v, exists := optsMap["objectLiteralTypeAssertions"]; exists {
		if str, ok := v.(string); ok {
			opts.ObjectLiteralTypeAssertions = LiteralAssertion(str)
		}
	}

	if v, exists := optsMap["arrayLiteralTypeAssertions"]; exists {
		if str, ok := v.(string); ok {
			opts.ArrayLiteralTypeAssertions = LiteralAssertion(str)
		}
	}
}
