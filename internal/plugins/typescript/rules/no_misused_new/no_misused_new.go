package no_misused_new

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Only identifiers and static computed string names can equal the valid
// identifiers this rule checks. Keep the hot path on the AST and avoid the
// generic source-text fallback, which creates a scanner for dynamic names.
func isMemberNamed(name *ast.Node, expected string) bool {
	if name == nil {
		return false
	}

	switch name.Kind {
	case ast.KindIdentifier:
		return name.AsIdentifier().Text == expected
	case ast.KindComputedPropertyName:
		expression := name.AsComputedPropertyName().Expression
		return ast.IsStringLiteralLike(expression) && expression.Text() == expected
	}
	return false
}

func returnsParentType(typeNode *ast.Node, parent *ast.Node) bool {
	if typeNode == nil || !ast.IsTypeReferenceNode(typeNode) {
		return false
	}

	typeName := typeNode.AsTypeReferenceNode().TypeName
	if !ast.IsIdentifier(typeName) {
		return false
	}

	parentName := parent.Name()
	return parentName != nil && typeName.Text() == parentName.Text()
}

var NoMisusedNewRule = rule.CreateRule(rule.Rule{
	Name: "no-misused-new",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindMethodDeclaration: func(node *ast.Node) {
				method := node.AsMethodDeclaration()
				// A method with an implementation is valid regardless of its name.
				if method.Body != nil {
					return
				}

				parentKind := node.Parent.Kind
				if parentKind != ast.KindClassDeclaration && parentKind != ast.KindClassExpression {
					return
				}

				if !isMemberNamed(method.Name(), "new") {
					return
				}

				if returnsParentType(method.Type, node.Parent) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "errorMessageClass",
						Description: "Class cannot have method named `new`.",
					})
				}
			},
			ast.KindConstructSignature: func(node *ast.Node) {
				if node.Parent.Kind != ast.KindInterfaceDeclaration {
					return
				}

				if returnsParentType(node.AsConstructSignatureDeclaration().Type, node.Parent) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "errorMessageInterface",
						Description: "interfaces cannot be constructed, only classes.",
					})
				}
			},
			ast.KindMethodSignature: func(node *ast.Node) {
				if isMemberNamed(node.AsMethodSignatureDeclaration().Name(), "constructor") {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "errorMessageInterface",
						Description: "interfaces cannot be constructed, only classes.",
					})
				}
			},
		}
	},
})
