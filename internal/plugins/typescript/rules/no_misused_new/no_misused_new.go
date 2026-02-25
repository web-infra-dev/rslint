package no_misused_new

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

/**
 * check whether the name of the return type of the node is the same as the name of the parent node.
 */
func check(node *ast.Node) bool {
	parentName := node.Parent.Name()
	if parentName == nil {
		return false
	}

	nodeType := node.Type()
	if nodeType != nil && ast.IsTypeReferenceNode(nodeType) {
		typeName := nodeType.AsTypeReference().TypeName
		if ast.IsIdentifier(typeName) {
			return typeName.Text() == parentName.Text()
		}
	}
	return false
}

var NoMisusedNewRule = rule.CreateRule(rule.Rule{
	Name: "no-misused-new",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindMethodDeclaration: func(node *ast.Node) {
				parentKind := node.Parent.Kind
				if parentKind != ast.KindClassDeclaration && parentKind != ast.KindClassExpression {
					return
				}

				nodeNameText, _ := utils.GetNameFromMember(ctx.SourceFile, node.Name())
				if nodeNameText != "new" {
					return
				}
				// If the function body exists, it's valid for this rule.
				body := node.Body()
				if body != nil {
					return
				}

				if check(node) {
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
				if check(node) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "errorMessageInterface",
						Description: "interfaces cannot be constructed, only classes.",
					})
				}
			},
			ast.KindMethodSignature: func(node *ast.Node) {
				nodeNameText, _ := utils.GetNameFromMember(ctx.SourceFile, node.Name())
				if nodeNameText == "constructor" {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "errorMessageInterface",
						Description: "interfaces cannot be constructed, only classes.",
					})
				}
			},
		}
	},
})
